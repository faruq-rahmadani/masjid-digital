package fridayduty

import (
	"context"
	"errors"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

func TestChangesPublishActivityLogs(t *testing.T) {
	ctx := context.Background()
	uc, repo, activityLog := newTestUsecaseWithActivityLog(t)

	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(domain.FridayDuty{Date: "2026-11-06", Khatib: "Ustadz Abdullah"}, nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityFridayDutySaved, Summary: "mengisi petugas Jumat 6 Nov 2026"})
	if _, err := uc.Save(ctx, domain.SaveFridayDutyRequest{Date: "2026-11-06", Khatib: "Ustadz Abdullah"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	repo.EXPECT().Delete(gomock.Any(), "2026-11-06").Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityFridayDutyDeleted, Summary: "menghapus petugas Jumat 6 Nov 2026"})
	if err := uc.Delete(ctx, "2026-11-06"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestFailedChangesPublishNothing(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newTestUsecaseWithActivityLog(t) // no Publish expected
	errDatabase := apperror.New(apperror.CodeInternal, errors.New("db down"), "gagal mengakses data petugas Jumat")

	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(domain.FridayDuty{}, errDatabase)
	repo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errDatabase)

	_, _ = uc.Save(ctx, domain.SaveFridayDutyRequest{Date: "2026-11-06", Khatib: "Ustadz Abdullah"})
	_ = uc.Delete(ctx, "2026-11-06")
}
