package activitylog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

func newTestUsecase(t *testing.T) (*UseCase, *mock.MockActivityLogRepositoryItf) {
	t.Helper()

	repo := mock.NewMockActivityLogRepositoryItf(gomock.NewController(t))
	return NewActivityLogUsecase(repo), repo
}

func TestSave(t *testing.T) {
	uc, repo := newTestUsecase(t)
	activityLog := domain.ActivityLog{
		ActorID:   7,
		Action:    domain.ActivitySlideCreated,
		EntityID:  4,
		Summary:   "menambah slide Kajian Ahad",
		CreatedAt: time.Date(2026, 11, 3, 1, 0, 0, 0, time.UTC),
	}
	repo.EXPECT().Create(gomock.Any(), activityLog).Return(nil)

	uc.Save(context.Background(), activityLog)
}

func TestSaveSwallowsErrors(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(apperror.New(apperror.CodeInternal, errors.New("db down"), "gagal mengakses log aktivitas"))

	uc.Save(context.Background(), domain.ActivityLog{Action: domain.ActivitySlideCreated}) // must not panic
}

func TestListRecent(t *testing.T) {
	tests := []struct {
		name      string
		limit     string
		wantLimit int
	}{
		{name: "default", limit: "", wantLimit: domain.DefaultActivityLogLimit},
		{name: "given", limit: "4", wantLimit: 4},
		{name: "max", limit: "50", wantLimit: domain.MaxActivityLogLimit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := newTestUsecase(t)
			want := []domain.ActivityLog{{ID: 1, Summary: "mengubah pengaturan iqamah"}}
			repo.EXPECT().FindRecent(gomock.Any(), tt.wantLimit).Return(want, nil)

			got, err := uc.ListRecent(context.Background(), domain.ListActivityLogsRequest{Limit: tt.limit})
			if err != nil || len(got) != 1 || got[0].ID != 1 {
				t.Errorf("got %+v, %v", got, err)
			}
		})
	}
}

func TestListRecentRejectsBadLimit(t *testing.T) {
	for _, limit := range []string{"0", "51", "-1", "abc"} {
		t.Run(limit, func(t *testing.T) {
			uc, _ := newTestUsecase(t)

			_, err := uc.ListRecent(context.Background(), domain.ListActivityLogsRequest{Limit: limit})
			var appErr *apperror.AppError
			if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalidInput {
				t.Errorf("err = %v, want INVALID_INPUT", err)
			}
		})
	}
}
