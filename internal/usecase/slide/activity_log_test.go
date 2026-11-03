package slide

import (
	"context"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"go.uber.org/mock/gomock"
)

func TestChangesPublishActivityLogs(t *testing.T) {
	ctx := context.Background()
	uc, repo, activityLog := newTestUsecaseWithActivityLog(t)

	repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.Slide{ID: 4, Title: "Kajian Ahad"}, nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivitySlideCreated, EntityID: 4, Summary: `menambah slide "Kajian Ahad"`})
	if _, err := uc.Create(ctx, domain.CreateSlideRequest{Title: "Kajian Ahad", Data: pngData}); err != nil {
		t.Fatalf("create: %v", err)
	}

	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.Slide{ID: 4, Title: "Kajian Ahad"}, nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivitySlideUpdated, EntityID: 4, Summary: `mengubah slide "Kajian Ahad"`})
	if _, err := uc.Update(ctx, domain.UpdateSlideRequest{ID: 4, Title: "Kajian Ahad", DurationSeconds: 10}); err != nil {
		t.Fatalf("update: %v", err)
	}

	repo.EXPECT().Delete(gomock.Any(), int64(4)).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivitySlideDeleted, EntityID: 4, Summary: "menghapus slide"})
	if err := uc.Delete(ctx, 4); err != nil {
		t.Fatalf("delete: %v", err)
	}

	repo.EXPECT().FindAll(gomock.Any()).Return(saved, nil).Times(2)
	repo.EXPECT().Reorder(gomock.Any(), []int64{3, 1, 2}).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivitySlidesReordered, Summary: "mengurutkan ulang slide"})
	if _, err := uc.Reorder(ctx, domain.ReorderSlidesRequest{IDs: []int64{3, 1, 2}}); err != nil {
		t.Fatalf("reorder: %v", err)
	}
}

func TestFailedChangesPublishNothing(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newTestUsecaseWithActivityLog(t) // no Publish expected

	repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.Slide{}, errDatabase)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.Slide{}, errDatabase)
	repo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errDatabase)

	_, _ = uc.Create(ctx, domain.CreateSlideRequest{Title: "Kajian Ahad", Data: pngData})
	_, _ = uc.Update(ctx, domain.UpdateSlideRequest{ID: 4, Title: "Kajian Ahad", DurationSeconds: 10})
	_ = uc.Delete(ctx, 4)
}
