package screen

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/realtime/screenhub"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

// now is 2026-11-02 10:00 UTC, 17:00 WIB
var now = time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)

var (
	errScreenNotFound   = apperror.New(apperror.CodeNotFound, domain.ErrScreenNotFound, "layar tidak ditemukan")
	errPresenceNotFound = apperror.New(apperror.CodeNotFound, domain.ErrScreenPresenceNotFound, "layar belum pernah terlihat")
	errProfileNotFound  = apperror.New(apperror.CodeNotFound, domain.ErrMosqueProfileNotFound, "profil masjid belum diisi")
)

type fixture struct {
	uc       *useCase
	screens  *mock.MockScreenRepositoryItf
	presence *mock.MockScreenPresenceRepositoryItf
	profile  *mock.MockMosqueProfileRepositoryItf
	hub      interface {
		Subscribe(string) (<-chan domain.ScreenEvent, func())
	}
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	ctrl := gomock.NewController(t)
	f := fixture{
		screens:  mock.NewMockScreenRepositoryItf(ctrl),
		presence: mock.NewMockScreenPresenceRepositoryItf(ctrl),
		profile:  mock.NewMockMosqueProfileRepositoryItf(ctrl),
	}
	hub := screenhub.New()
	f.hub = hub
	f.uc = New(f.screens, f.presence, hub, f.profile)
	f.uc.now = func() time.Time { return now }
	return f
}

func receive(ch <-chan domain.ScreenEvent) domain.ScreenEvent {
	select {
	case ev := <-ch:
		return ev
	default:
		return domain.ScreenEvent{}
	}
}

func TestConnectRegistersOnceAndRecordsPresence(t *testing.T) {
	f := newFixture(t)
	req := domain.ConnectScreenRequest{DeviceID: "device-0001", NameHint: "teras", AppVersion: "1.4.0"}
	registered := domain.Screen{ID: 1, DeviceID: "device-0001", NameHint: "teras", ForcedMode: domain.ScreenModeAuto}

	f.screens.EXPECT().FindByDeviceID(gomock.Any(), "device-0001").Return(domain.Screen{}, errScreenNotFound)
	f.screens.EXPECT().FindAll(gomock.Any()).Return(nil, nil)
	f.screens.EXPECT().Register(gomock.Any(), domain.Screen{DeviceID: "device-0001", NameHint: "teras"}).Return(registered, nil)
	f.presence.EXPECT().Set(gomock.Any(), "device-0001", domain.ScreenPresence{LastSeenAt: now, AppVersion: "1.4.0"}).Return(nil).Times(2) // connect and close

	conn, err := f.uc.Connect(context.Background(), req)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if conn.State.ForcedMode != domain.ScreenModeAuto {
		t.Errorf("state = %+v", conn.State)
	}

	f.uc.NotifyContentChanged()
	if ev := receive(conn.Events); ev.Type != domain.ScreenEventRefresh {
		t.Errorf("event = %q, want refresh", ev.Type)
	}
	conn.Close()
}

func TestConnectRefusesBlockedAndFull(t *testing.T) {
	f := newFixture(t)
	blockedAt := now.Add(-time.Hour)
	f.screens.EXPECT().FindByDeviceID(gomock.Any(), "device-0001").Return(domain.Screen{ID: 1, DeviceID: "device-0001", BlockedAt: &blockedAt}, nil)

	_, err := f.uc.Connect(context.Background(), domain.ConnectScreenRequest{DeviceID: "device-0001"})
	if !errors.Is(err, domain.ErrScreenBlocked) {
		t.Errorf("blocked: err = %v, want ErrScreenBlocked", err)
	}

	// a new device when the screens are full; a known one still connects (checked above)
	full := make([]domain.Screen, domain.MaxScreens)
	f.screens.EXPECT().FindByDeviceID(gomock.Any(), "device-0002").Return(domain.Screen{}, errScreenNotFound)
	f.screens.EXPECT().FindAll(gomock.Any()).Return(full, nil)
	_, err = f.uc.Connect(context.Background(), domain.ConnectScreenRequest{DeviceID: "device-0002"})
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeBusinessRule {
		t.Errorf("full: err = %v, want BUSINESS_RULE", err)
	}
}

func TestListStatus(t *testing.T) {
	f := newFixture(t)
	f.screens.EXPECT().FindAll(gomock.Any()).Return([]domain.Screen{
		{ID: 1, DeviceID: "connected"},
		{ID: 2, DeviceID: "just-left"},
		{ID: 3, DeviceID: "long-gone"},
		{ID: 4, DeviceID: "never-cached"},
	}, nil)
	f.presence.EXPECT().Get(gomock.Any(), "connected").Return(domain.ScreenPresence{LastSeenAt: now.Add(-time.Hour)}, nil)
	f.presence.EXPECT().Get(gomock.Any(), "just-left").Return(domain.ScreenPresence{LastSeenAt: now.Add(-time.Minute), AppVersion: "1.4.0"}, nil)
	f.presence.EXPECT().Get(gomock.Any(), "long-gone").Return(domain.ScreenPresence{LastSeenAt: now.Add(-3 * time.Minute)}, nil)
	f.presence.EXPECT().Get(gomock.Any(), "never-cached").Return(domain.ScreenPresence{}, errPresenceNotFound)
	_, unsubscribe := f.hub.Subscribe("connected")
	defer unsubscribe()

	screens, err := f.uc.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	want := []domain.ScreenStatus{domain.ScreenStatusOnline, domain.ScreenStatusOnline, domain.ScreenStatusOffline, domain.ScreenStatusOffline}
	for i, s := range screens {
		if s.Status != want[i] {
			t.Errorf("%s: status = %s, want %s", s.DeviceID, s.Status, want[i])
		}
	}
	if screens[1].AppVersion != "1.4.0" || screens[3].LastSeenAt != nil {
		t.Errorf("presence fields = %+v / %+v", screens[1], screens[3])
	}
}

func TestUpdateForcesUntilMidnight(t *testing.T) {
	f := newFixture(t)
	screen := domain.Screen{ID: 1, DeviceID: "device-0001", ForcedMode: domain.ScreenModeAuto}
	events, unsubscribe := f.hub.Subscribe("device-0001")
	defer unsubscribe()

	f.screens.EXPECT().FindByID(gomock.Any(), int64(1)).Return(screen, nil)
	f.profile.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{Timezone: domain.MosqueTimezoneWITA}, nil)
	f.screens.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, s domain.Screen) (domain.Screen, error) {
		// 18:00 WITA now, so the mode ends at 00:00 WITA = 16:00 UTC
		if s.Name != "TV Teras" || s.ForcedMode != domain.ScreenModeDark || s.ForcedUntil == nil || !s.ForcedUntil.Equal(time.Date(2026, 11, 2, 16, 0, 0, 0, time.UTC)) {
			t.Errorf("saved = %+v", s)
		}
		return s, nil
	})
	f.presence.EXPECT().Get(gomock.Any(), "device-0001").Return(domain.ScreenPresence{}, errPresenceNotFound)

	if _, err := f.uc.Update(context.Background(), domain.UpdateScreenRequest{ID: 1, Name: " TV Teras ", ForcedMode: domain.ScreenModeDark}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if ev := receive(events); ev.Type != domain.ScreenEventState {
		t.Errorf("TV got %q, want state", ev.Type)
	}

	// back to automatic: no profile needed, no end time
	f.screens.EXPECT().FindByID(gomock.Any(), int64(1)).Return(screen, nil)
	f.screens.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, s domain.Screen) (domain.Screen, error) {
		if s.ForcedUntil != nil {
			t.Errorf("auto kept until %v", s.ForcedUntil)
		}
		return s, nil
	})
	f.presence.EXPECT().Get(gomock.Any(), "device-0001").Return(domain.ScreenPresence{}, errPresenceNotFound)
	if _, err := f.uc.Update(context.Background(), domain.UpdateScreenRequest{ID: 1, Name: "TV Teras", ForcedMode: domain.ScreenModeAuto}); err != nil {
		t.Fatalf("update auto: %v", err)
	}
}

func TestForcedModeWithoutProfileUsesWIB(t *testing.T) {
	f := newFixture(t)
	f.profile.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{}, errProfileNotFound)

	until, err := f.uc.forcedUntil(context.Background(), domain.ScreenModeNormal)
	if err != nil || until == nil || !until.Equal(time.Date(2026, 11, 2, 17, 0, 0, 0, time.UTC)) {
		t.Errorf("until = %v, %v; want WIB midnight", until, err)
	}
}

func TestBlockClosesTheTVAndUnblock(t *testing.T) {
	f := newFixture(t)
	events, unsubscribe := f.hub.Subscribe("device-0001")
	defer unsubscribe()
	screen := domain.Screen{ID: 1, DeviceID: "device-0001"}

	f.screens.EXPECT().FindByID(gomock.Any(), int64(1)).Return(screen, nil)
	f.screens.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, s domain.Screen) (domain.Screen, error) {
		if s.BlockedAt == nil || !s.BlockedAt.Equal(now) {
			t.Errorf("blocked_at = %v", s.BlockedAt)
		}
		return s, nil
	})
	f.presence.EXPECT().Get(gomock.Any(), gomock.Any()).Return(domain.ScreenPresence{}, errPresenceNotFound).AnyTimes()

	blocked, err := f.uc.Block(context.Background(), 1)
	if err != nil || !blocked.IsBlocked() {
		t.Fatalf("block: %+v, %v", blocked, err)
	}
	if ev := receive(events); ev.Type != domain.ScreenEventBlocked {
		t.Errorf("TV got %q, want blocked", ev.Type)
	}

	f.screens.EXPECT().FindByID(gomock.Any(), int64(1)).Return(blocked, nil)
	f.screens.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, s domain.Screen) (domain.Screen, error) { return s, nil })
	if unblocked, err := f.uc.Unblock(context.Background(), 1); err != nil || unblocked.IsBlocked() {
		t.Errorf("unblock: %+v, %v", unblocked, err)
	}
}

func TestSendCommand(t *testing.T) {
	f := newFixture(t)
	tv1, close1 := f.hub.Subscribe("device-0001")
	defer close1()
	tv2, close2 := f.hub.Subscribe("device-0002")
	defer close2()

	f.screens.EXPECT().FindByID(gomock.Any(), int64(1)).Return(domain.Screen{ID: 1, DeviceID: "device-0001"}, nil)
	if err := f.uc.SendCommand(context.Background(), domain.SendScreenCommandRequest{ScreenID: 1, Command: domain.ScreenCommandReload}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if receive(tv1).Type != domain.ScreenEventReload || receive(tv2).Type != "" {
		t.Error("reload should reach only device-0001")
	}

	if err := f.uc.SendCommand(context.Background(), domain.SendScreenCommandRequest{Command: domain.ScreenCommandTestSound}); err != nil {
		t.Fatalf("broadcast: %v", err)
	}
	if receive(tv1).Type != domain.ScreenEventTestSound || receive(tv2).Type != domain.ScreenEventTestSound {
		t.Error("test sound should reach every TV")
	}

	if err := f.uc.SendCommand(context.Background(), domain.SendScreenCommandRequest{Command: "shutdown"}); err == nil {
		t.Error("want a validation error")
	}
}
