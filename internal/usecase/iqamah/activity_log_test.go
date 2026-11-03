package iqamah

import (
	"context"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"go.uber.org/mock/gomock"
)

func TestUpdateSettingsPublishesActivityLog(t *testing.T) {
	uc, repo, activityLog := newTestUsecaseWithActivityLog(t)
	repo.EXPECT().UpdateSettings(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityIqamahSettingsUpdated, Summary: "mengubah pengaturan iqamah"})

	if _, err := uc.UpdateSettings(context.Background(), domain.UpdateIqamahSettingsRequest{IqamahSettings: domain.DefaultIqamahSettings()}); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestFailedUpdatePublishesNothing(t *testing.T) {
	uc, repo, _ := newTestUsecaseWithActivityLog(t) // no Publish expected
	repo.EXPECT().UpdateSettings(gomock.Any(), gomock.Any(), gomock.Any()).Return(errDatabase)

	_, _ = uc.UpdateSettings(context.Background(), domain.UpdateIqamahSettingsRequest{IqamahSettings: domain.DefaultIqamahSettings()})
}
