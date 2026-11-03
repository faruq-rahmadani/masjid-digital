package screen

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

type useCase struct {
	screens     usecase.ScreenRepositoryItf
	presence    usecase.ScreenPresenceRepositoryItf
	hub         usecase.ScreenHubItf
	profile     usecase.MosqueProfileRepositoryItf
	activityLog usecase.ActivityLogPublisherItf
	now         func() time.Time
}

func New(screens usecase.ScreenRepositoryItf, presence usecase.ScreenPresenceRepositoryItf, hub usecase.ScreenHubItf, profile usecase.MosqueProfileRepositoryItf, activityLog usecase.ActivityLogPublisherItf) *useCase {
	return &useCase{screens: screens, presence: presence, hub: hub, profile: profile, activityLog: activityLog, now: time.Now}
}

// List returns every screen, blocked ones too, with whether each is online and when it was last
// seen. The admin shows the blocked ones apart.
func (u *useCase) List(ctx context.Context) ([]domain.Screen, error) {
	screens, err := u.screens.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range screens {
		u.withPresence(ctx, &screens[i])
	}
	return screens, nil
}

// withPresence fills the live fields. A screen is online while connected, and for
// domain.ScreenOfflineAfter after it disconnects. Presence is a nice-to-have: when the cache
// can't answer, the screen just shows no last seen.
func (u *useCase) withPresence(ctx context.Context, screen *domain.Screen) {
	screen.Status = domain.ScreenStatusOffline
	if presence, err := u.presence.Get(ctx, screen.DeviceID); err == nil {
		screen.LastSeenAt, screen.AppVersion = &presence.LastSeenAt, presence.AppVersion
	}
	if u.hub.IsConnected(screen.DeviceID) {
		screen.Status = domain.ScreenStatusOnline
		return
	}
	if screen.LastSeenAt != nil && u.now().Sub(*screen.LastSeenAt) < domain.ScreenOfflineAfter {
		screen.Status = domain.ScreenStatusOnline
	}
}

// Connect opens a TV's connection: the device gets its row the first time, a blocked one is
// refused with domain.ErrScreenBlocked inside the error.
func (u *useCase) Connect(ctx context.Context, req domain.ConnectScreenRequest) (domain.ScreenConnection, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.ScreenConnection{}, err
	}

	screen, err := u.register(ctx, req)
	if err != nil {
		return domain.ScreenConnection{}, err
	}
	if screen.IsBlocked() {
		return domain.ScreenConnection{}, apperror.New(apperror.CodeForbidden, domain.ErrScreenBlocked, "layar ini diblokir")
	}

	u.seen(ctx, req)
	events, unsubscribe := u.hub.Subscribe(req.DeviceID)
	return domain.ScreenConnection{
		State:  screen.State(),
		Events: events,
		Close: func() {
			unsubscribe()
			// the request's context is gone by now
			u.seen(context.Background(), req)
		},
	}, nil
}

// register returns the device's row, inserting it the first time unless the screens are full
func (u *useCase) register(ctx context.Context, req domain.ConnectScreenRequest) (domain.Screen, error) {
	screen, err := u.screens.FindByDeviceID(ctx, req.DeviceID)
	if err == nil || !errors.Is(err, domain.ErrScreenNotFound) {
		return screen, err
	}

	all, err := u.screens.FindAll(ctx)
	if err != nil {
		return domain.Screen{}, err
	}
	if len(all) >= domain.MaxScreens {
		return domain.Screen{}, apperror.New(apperror.CodeBusinessRule, fmt.Errorf("%d screens registered", len(all)), fmt.Sprintf("jumlah layar sudah mencapai batas %d", domain.MaxScreens))
	}
	return u.screens.Register(ctx, domain.Screen{DeviceID: req.DeviceID, NameHint: req.NameHint, UserAgent: req.UserAgent})
}

// seen records the device as seen now. Best effort: a cache outage must not cut TVs off.
func (u *useCase) seen(ctx context.Context, req domain.ConnectScreenRequest) {
	_ = u.presence.Set(ctx, req.DeviceID, domain.ScreenPresence{LastSeenAt: u.now().UTC(), AppVersion: req.AppVersion})
}

// Update names a screen and sets its mode. A forced mode lasts until the mosque's next midnight,
// so a forgotten "Layar gelap" never keeps a TV dark for days. The TV gets its new state at once.
func (u *useCase) Update(ctx context.Context, req domain.UpdateScreenRequest) (domain.Screen, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.Screen{}, err
	}
	screen, err := u.screens.FindByID(ctx, req.ID)
	if err != nil {
		return domain.Screen{}, err
	}

	until, err := u.forcedUntil(ctx, req.ForcedMode)
	if err != nil {
		return domain.Screen{}, err
	}
	screen.Name, screen.ForcedMode, screen.ForcedUntil = req.Name, req.ForcedMode, until

	saved, err := u.save(ctx, screen, domain.ScreenEvent{Type: domain.ScreenEventState, Data: screen.State()})
	if err != nil {
		return domain.Screen{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog())
	return saved, nil
}

// forcedUntil is when a forced mode ends: nil for automatic, else the mosque's next midnight
func (u *useCase) forcedUntil(ctx context.Context, mode domain.ScreenMode) (*time.Time, error) {
	if mode == domain.ScreenModeAuto {
		return nil, nil
	}
	profile, err := u.profile.Get(ctx)
	if err != nil && !errors.Is(err, domain.ErrMosqueProfileNotFound) {
		return nil, err
	}
	// without a profile the zone is empty, which reads as WIB
	until := profile.Timezone.NextMidnight(u.now()).UTC()
	return &until, nil
}

// Block hides a screen from the list and refuses its connections; an open one is closed now
func (u *useCase) Block(ctx context.Context, id int64) (domain.Screen, error) {
	screen, err := u.screens.FindByID(ctx, id)
	if err != nil {
		return domain.Screen{}, err
	}
	if screen.IsBlocked() {
		return u.present(ctx, screen), nil
	}
	now := u.now().UTC()
	screen.BlockedAt = &now
	saved, err := u.save(ctx, screen, domain.ScreenEvent{Type: domain.ScreenEventBlocked})
	if err != nil {
		return domain.Screen{}, err
	}
	u.activityLog.Publish(ctx, saved.BlockedActivityLog())
	return saved, nil
}

// Unblock puts a blocked screen back; the TV reconnects on its next retry
func (u *useCase) Unblock(ctx context.Context, id int64) (domain.Screen, error) {
	screen, err := u.screens.FindByID(ctx, id)
	if err != nil {
		return domain.Screen{}, err
	}
	if !screen.IsBlocked() {
		return u.present(ctx, screen), nil
	}
	screen.BlockedAt = nil
	saved, err := u.save(ctx, screen, domain.ScreenEvent{})
	if err != nil {
		return domain.Screen{}, err
	}
	u.activityLog.Publish(ctx, saved.UnblockedActivityLog())
	return saved, nil
}

// save stores the screen, then tells its TV (an empty event type sends nothing)
func (u *useCase) save(ctx context.Context, screen domain.Screen, event domain.ScreenEvent) (domain.Screen, error) {
	saved, err := u.screens.Update(ctx, screen)
	if err != nil {
		return domain.Screen{}, err
	}
	if event.Type != "" {
		u.hub.Send(saved.DeviceID, event)
	}
	return u.present(ctx, saved), nil
}

func (u *useCase) present(ctx context.Context, screen domain.Screen) domain.Screen {
	u.withPresence(ctx, &screen)
	return screen
}

// SendCommand sends a reload or a test sound to one screen, or to all when req.ScreenID is 0
func (u *useCase) SendCommand(ctx context.Context, req domain.SendScreenCommandRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	event := domain.ScreenEvent{Type: domain.ScreenEventType(req.Command)}
	if req.ScreenID == 0 {
		u.hub.Broadcast(event)
		return nil
	}

	screen, err := u.screens.FindByID(ctx, req.ScreenID)
	if err != nil {
		return err
	}
	u.hub.Send(screen.DeviceID, event)
	return nil
}

// NotifyContentChanged tells every TV to reload its data now instead of at the next 5-minute mark
func (u *useCase) NotifyContentChanged() {
	u.hub.Broadcast(domain.ScreenEvent{Type: domain.ScreenEventRefresh})
}
