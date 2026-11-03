package fridayduty

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

// now is 2026-11-01 20:00 UTC: already 2026-11-02 in Jakarta, still 2026-11-01 in UTC
var now = time.Date(2026, 11, 1, 20, 0, 0, 0, time.UTC)

// newTestUsecase is for tests not about the activity log: any log may be published
func newTestUsecase(t *testing.T) (*UseCase, *mock.MockFridayDutyRepositoryItf) {
	t.Helper()

	uc, repo, activityLog := newTestUsecaseWithActivityLog(t)
	activityLog.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes()
	return uc, repo
}

func newTestUsecaseWithActivityLog(t *testing.T) (*UseCase, *mock.MockFridayDutyRepositoryItf, *mock.MockActivityLogPublisherItf) {
	t.Helper()

	ctrl := gomock.NewController(t)
	repo := mock.NewMockFridayDutyRepositoryItf(ctrl)
	activityLog := mock.NewMockActivityLogPublisherItf(ctrl)
	uc := NewFridayDutyUsecase(repo, activityLog)
	uc.now = func() time.Time { return now }
	return uc, repo, activityLog
}

func TestList(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().FindBetween(gomock.Any(), "2026-11-01", "2026-11-30").Return(nil, nil)

	if _, err := uc.List(context.Background(), domain.FridayDutyRange{From: " 2026-11-01", To: "2026-11-30"}); err != nil {
		t.Errorf("list: %v", err)
	}

	// invalid: never reaches the repository
	if _, err := uc.List(context.Background(), domain.FridayDutyRange{From: "2026-11-30", To: "2026-11-01"}); err == nil {
		t.Error("want a validation error")
	}
}

func TestListUpcoming(t *testing.T) {
	uc, repo := newTestUsecase(t)
	duties := []domain.FridayDuty{{Date: "2026-11-06", Khatib: "Ustadz Abdullah"}}
	repo.EXPECT().FindBetween(gomock.Any(), "2026-11-01", "2026-12-01").Return(duties, nil)

	got, err := uc.ListUpcoming(context.Background())
	if err != nil || !reflect.DeepEqual(got, duties) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestSave(t *testing.T) {
	uc, repo := newTestUsecase(t)
	want := domain.FridayDuty{Date: "2026-11-06", Khatib: "Ustadz Abdullah", UpdatedBy: 7}
	repo.EXPECT().Save(gomock.Any(), want).Return(want, nil)

	got, err := uc.Save(context.Background(), domain.SaveFridayDutyRequest{UpdatedBy: 7, Date: "2026-11-06", Khatib: " Ustadz Abdullah "})
	if err != nil || got != want {
		t.Errorf("got %+v, %v", got, err)
	}

	_, err = uc.Save(context.Background(), domain.SaveFridayDutyRequest{Date: "2026-10-30", Khatib: "A"})
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.MessageError != "tanggal sudah lewat" {
		t.Errorf("past friday: err = %v", err)
	}
}
