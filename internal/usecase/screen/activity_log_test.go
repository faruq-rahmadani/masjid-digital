package screen

import (
	"context"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"go.uber.org/mock/gomock"
)

func TestChangesPublishActivityLogs(t *testing.T) {
	ctx := context.Background()
	f := newFixtureWithActivityLog(t)
	screen := domain.Screen{ID: 1, DeviceID: "device-0001", Name: "TV Teras"}
	f.screens.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, s domain.Screen) (domain.Screen, error) { return s, nil }).Times(3)
	f.presence.EXPECT().Get(gomock.Any(), "device-0001").Return(domain.ScreenPresence{}, errPresenceNotFound).AnyTimes()

	f.screens.EXPECT().FindByID(gomock.Any(), int64(1)).Return(screen, nil)
	f.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityScreenUpdated, EntityID: 1, Summary: "mengubah layar TV Teras"})
	if _, err := f.uc.Update(ctx, domain.UpdateScreenRequest{ID: 1, Name: "TV Teras", ForcedMode: domain.ScreenModeAuto}); err != nil {
		t.Fatalf("update: %v", err)
	}

	f.screens.EXPECT().FindByID(gomock.Any(), int64(1)).Return(screen, nil)
	f.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityScreenBlocked, EntityID: 1, Summary: "memblokir layar TV Teras"})
	if _, err := f.uc.Block(ctx, 1); err != nil {
		t.Fatalf("block: %v", err)
	}

	blocked := screen
	blocked.Name, blocked.NameHint, blocked.BlockedAt = "", "masjid-depan", &now // unnamed: the hint names it
	f.screens.EXPECT().FindByID(gomock.Any(), int64(1)).Return(blocked, nil)
	f.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityScreenUnblocked, EntityID: 1, Summary: "membuka blokir layar masjid-depan"})
	if _, err := f.uc.Unblock(ctx, 1); err != nil {
		t.Fatalf("unblock: %v", err)
	}
}

func TestNoChangePublishesNothing(t *testing.T) {
	ctx := context.Background()
	f := newFixtureWithActivityLog(t) // no Publish expected
	f.presence.EXPECT().Get(gomock.Any(), gomock.Any()).Return(domain.ScreenPresence{}, errPresenceNotFound).AnyTimes()

	// blocking a blocked screen and unblocking an unblocked one save nothing
	f.screens.EXPECT().FindByID(gomock.Any(), int64(1)).Return(domain.Screen{ID: 1, DeviceID: "device-0001", BlockedAt: &now}, nil)
	f.screens.EXPECT().FindByID(gomock.Any(), int64(2)).Return(domain.Screen{ID: 2, DeviceID: "device-0002"}, nil)
	f.screens.EXPECT().FindByID(gomock.Any(), int64(3)).Return(domain.Screen{}, errScreenNotFound)
	_, _ = f.uc.Block(ctx, 1)
	_, _ = f.uc.Unblock(ctx, 2)
	_, _ = f.uc.Update(ctx, domain.UpdateScreenRequest{ID: 3, Name: "TV Teras", ForcedMode: domain.ScreenModeAuto})
}
