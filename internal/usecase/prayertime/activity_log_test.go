package prayertime

import (
	"context"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"go.uber.org/mock/gomock"
)

func TestChangesPublishActivityLogs(t *testing.T) {
	ctx := context.Background()
	uc, repo, activityLog := newTestUsecaseWithActivityLog(t)

	repo.EXPECT().UpdateSettings(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityPrayerSettingsUpdated, Summary: "mengubah pengaturan waktu sholat"})
	if _, err := uc.UpdateSettings(ctx, domain.UpdatePrayerSettingsRequest{PrayerSettings: settings}); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	repo.EXPECT().CreateOverride(gomock.Any(), gomock.Any()).Return(domain.PrayerTimeOverride{ID: 5, Date: "2026-11-06", Prayer: domain.PrayerMaghrib, Time: "17:55"}, nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityPrayerOverrideCreated, EntityID: 5, Summary: "menambah jadwal khusus Maghrib 6 Nov 2026"})
	if _, err := uc.CreateOverride(ctx, domain.CreatePrayerTimeOverrideRequest{Date: "2026-11-06", Prayer: domain.PrayerMaghrib, Time: "17:55"}); err != nil {
		t.Fatalf("create override: %v", err)
	}

	repo.EXPECT().DeleteOverride(gomock.Any(), int64(5)).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityPrayerOverrideDeleted, EntityID: 5, Summary: "menghapus jadwal khusus"})
	if err := uc.DeleteOverride(ctx, 5); err != nil {
		t.Fatalf("delete override: %v", err)
	}
}

func TestFailedChangesPublishNothing(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newTestUsecaseWithActivityLog(t) // no Publish expected

	repo.EXPECT().UpdateSettings(gomock.Any(), gomock.Any(), gomock.Any()).Return(errDatabase)
	repo.EXPECT().CreateOverride(gomock.Any(), gomock.Any()).Return(domain.PrayerTimeOverride{}, errDatabase)
	repo.EXPECT().DeleteOverride(gomock.Any(), gomock.Any()).Return(errDatabase)

	_, _ = uc.UpdateSettings(ctx, domain.UpdatePrayerSettingsRequest{PrayerSettings: settings})
	_, _ = uc.CreateOverride(ctx, domain.CreatePrayerTimeOverrideRequest{Date: "2026-11-06", Prayer: domain.PrayerMaghrib, Time: "17:55"})
	_ = uc.DeleteOverride(ctx, 5)
}
