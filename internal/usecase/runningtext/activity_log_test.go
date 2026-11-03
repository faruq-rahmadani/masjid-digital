package runningtext

import (
	"context"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"go.uber.org/mock/gomock"
)

func TestChangesPublishActivityLogs(t *testing.T) {
	ctx := context.Background()
	uc, repo, activityLog := newTestUsecaseWithActivityLog(t)

	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.RunningText{ID: 4, Text: "Kerja bakti"}, nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityRunningTextCreated, EntityID: 4, Summary: `menambah running text "Kerja bakti"`})
	if _, err := uc.Create(ctx, domain.CreateRunningTextRequest{Text: "Kerja bakti"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.RunningText{ID: 4, Text: "Kerja bakti Ahad"}, nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityRunningTextUpdated, EntityID: 4, Summary: `mengubah running text "Kerja bakti Ahad"`})
	if _, err := uc.Update(ctx, domain.UpdateRunningTextRequest{ID: 4, Text: "Kerja bakti Ahad"}); err != nil {
		t.Fatalf("update: %v", err)
	}

	repo.EXPECT().Delete(gomock.Any(), int64(4)).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityRunningTextDeleted, EntityID: 4, Summary: "menghapus running text"})
	if err := uc.Delete(ctx, 4); err != nil {
		t.Fatalf("delete: %v", err)
	}

	repo.EXPECT().FindAll(gomock.Any()).Return(saved, nil).Times(2)
	repo.EXPECT().Reorder(gomock.Any(), []int64{3, 1, 2}).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityRunningTextsReordered, Summary: "mengurutkan ulang running text"})
	if _, err := uc.Reorder(ctx, domain.ReorderRunningTextsRequest{IDs: []int64{3, 1, 2}}); err != nil {
		t.Fatalf("reorder: %v", err)
	}

	repo.EXPECT().UpdateSetting(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityRunningTextSettingUpdated, Summary: "mengubah kecepatan running text"})
	if _, err := uc.UpdateSetting(ctx, domain.UpdateRunningTextSettingRequest{RunningTextSetting: domain.RunningTextSetting{Speed: domain.ScrollSpeedFast}}); err != nil {
		t.Fatalf("update setting: %v", err)
	}
}

func TestFailedChangesPublishNothing(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newTestUsecaseWithActivityLog(t) // no Publish expected

	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.RunningText{}, errDatabase)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.RunningText{}, errDatabase)
	repo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errDatabase)
	repo.EXPECT().UpdateSetting(gomock.Any(), gomock.Any(), gomock.Any()).Return(errDatabase)

	_, _ = uc.Create(ctx, domain.CreateRunningTextRequest{Text: "Kerja bakti"})
	_, _ = uc.Update(ctx, domain.UpdateRunningTextRequest{ID: 4, Text: "Kerja bakti"})
	_ = uc.Delete(ctx, 4)
	_, _ = uc.UpdateSetting(ctx, domain.UpdateRunningTextSettingRequest{RunningTextSetting: domain.RunningTextSetting{Speed: domain.ScrollSpeedFast}})
}
