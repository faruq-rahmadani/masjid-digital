package mosqueprofile

import (
	"context"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"go.uber.org/mock/gomock"
)

func TestChangesPublishActivityLogs(t *testing.T) {
	ctx := context.Background()
	uc, repo, activityLog := newTestUsecaseWithActivityLog(t)

	repo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityMosqueProfileUpdated, Summary: "mengubah profil masjid"})
	if _, err := uc.Update(ctx, validProfileRequest()); err != nil {
		t.Fatalf("update: %v", err)
	}

	repo.EXPECT().UpdateLogo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityMosqueLogoUpdated, Summary: "mengganti logo masjid"})
	if err := uc.UpdateLogo(ctx, domain.UpdateMosqueLogoRequest{Data: pngHeader}); err != nil {
		t.Fatalf("update logo: %v", err)
	}

	repo.EXPECT().DeleteLogo(gomock.Any()).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityMosqueLogoDeleted, Summary: "menghapus logo masjid"})
	if err := uc.DeleteLogo(ctx); err != nil {
		t.Fatalf("delete logo: %v", err)
	}
}

func TestFailedChangesPublishNothing(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newTestUsecaseWithActivityLog(t) // no Publish expected

	repo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(errDatabase)
	repo.EXPECT().UpdateLogo(gomock.Any(), gomock.Any(), gomock.Any()).Return(errDatabase)
	repo.EXPECT().DeleteLogo(gomock.Any()).Return(errDatabase)

	_, _ = uc.Update(ctx, validProfileRequest())
	_ = uc.UpdateLogo(ctx, domain.UpdateMosqueLogoRequest{Data: pngHeader})
	_ = uc.DeleteLogo(ctx)
}
