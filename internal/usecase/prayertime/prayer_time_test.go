package prayertime

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

// repository errors, shaped like the real repository's AppErrors
var (
	errSettingsNotFound = apperror.New(apperror.CodeNotFound, domain.ErrPrayerSettingsNotFound, "pengaturan waktu sholat belum disimpan")
	errDatabase         = apperror.New(apperror.CodeInternal, errors.New("db down"), "gagal mengakses data waktu sholat")
)

// now is 2026-11-01 20:00 UTC: already 2026-11-02 in Jakarta, still 2026-11-01 in UTC
var now = time.Date(2026, 11, 1, 20, 0, 0, 0, time.UTC)

func newTestUsecase(t *testing.T) (*UseCase, *mock.MockPrayerTimeRepositoryItf) {
	t.Helper()

	repo := mock.NewMockPrayerTimeRepositoryItf(gomock.NewController(t))
	uc := NewPrayerTimeUsecase(repo)
	uc.now = func() time.Time { return now }
	return uc, repo
}

var settings = domain.PrayerSettings{
	Corrections:     domain.PrayerCorrections{Subuh: 2, Maghrib: 2},
	ImsakOffsetMin:  10,
	WeeklyOverrides: []domain.WeeklyPrayerOverride{{Weekday: "friday", Prayer: domain.PrayerDzuhur, Time: "12:00", Remarks: "Sholat Jumat"}},
}

func TestGetSettingsDefaultsWhenNeverSaved(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().GetSettings(gomock.Any()).Return(domain.PrayerSettings{}, errSettingsNotFound)

	got, err := uc.GetSettings(context.Background())
	if err != nil || !reflect.DeepEqual(got, domain.DefaultPrayerSettings()) {
		t.Errorf("got %+v, %v; want defaults", got, err)
	}
}

func TestGetSettingsPassesOtherErrors(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().GetSettings(gomock.Any()).Return(domain.PrayerSettings{}, errDatabase)

	if _, err := uc.GetSettings(context.Background()); !errors.Is(err, errDatabase) {
		t.Errorf("err = %v, want errDatabase", err)
	}
}

func TestUpdateSettings(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().UpdateSettings(gomock.Any(), settings, int64(7)).Return(nil)

	req := domain.UpdatePrayerSettingsRequest{UpdatedBy: 7, PrayerSettings: settings}
	req.WeeklyOverrides = []domain.WeeklyPrayerOverride{{Weekday: "friday", Prayer: domain.PrayerDzuhur, Time: " 12:00", Remarks: "Sholat Jumat "}}

	got, err := uc.UpdateSettings(context.Background(), req)
	if err != nil || !reflect.DeepEqual(got, settings) {
		t.Errorf("got %+v, %v; want %+v", got, err, settings)
	}
}

func TestUpdateSettingsInvalidNeverSaves(t *testing.T) {
	uc, _ := newTestUsecase(t)

	req := domain.UpdatePrayerSettingsRequest{PrayerSettings: settings}
	req.ImsakOffsetMin = 0
	if _, err := uc.UpdateSettings(context.Background(), req); err == nil {
		t.Error("want a validation error")
	}
}

func TestCreateOverride(t *testing.T) {
	uc, repo := newTestUsecase(t)
	want := domain.PrayerTimeOverride{ID: 3, Date: "2026-11-01", Prayer: domain.PrayerMaghrib, Time: "17:55", CreatedBy: 7}
	repo.EXPECT().CreateOverride(gomock.Any(), domain.PrayerTimeOverride{Date: "2026-11-01", Prayer: domain.PrayerMaghrib, Time: "17:55", CreatedBy: 7}).Return(want, nil)

	// the UTC today is still allowed: it may be today for the mosque
	got, err := uc.CreateOverride(context.Background(), domain.CreatePrayerTimeOverrideRequest{CreatedBy: 7, Date: "2026-11-01", Prayer: domain.PrayerMaghrib, Time: "17:55"})
	if err != nil || got != want {
		t.Errorf("got %+v, %v; want %+v", got, err, want)
	}
}

func TestCreateOverridePastDate(t *testing.T) {
	uc, _ := newTestUsecase(t)

	_, err := uc.CreateOverride(context.Background(), domain.CreatePrayerTimeOverrideRequest{Date: "2026-10-31", Prayer: domain.PrayerMaghrib, Time: "17:55"})
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.MessageError != "tanggal sudah lewat" {
		t.Errorf("err = %v, want tanggal sudah lewat", err)
	}
}

func TestListOverrides(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().FindOverridesBetween(gomock.Any(), "2026-11-01", "2026-11-30").Return([]domain.PrayerTimeOverride{}, nil)

	if _, err := uc.ListOverrides(context.Background(), domain.PrayerTimeOverrideRange{From: "2026-11-01", To: "2026-11-30"}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if _, err := uc.ListOverrides(context.Background(), domain.PrayerTimeOverrideRange{From: "2026-11-30"}); err == nil {
		t.Error("missing to: want a validation error")
	}
}

func TestGetConfig(t *testing.T) {
	uc, repo := newTestUsecase(t)
	overrides := []domain.PrayerTimeOverride{{ID: 1, Date: "2026-11-06", Prayer: domain.PrayerDzuhur, Time: "11:58"}}
	repo.EXPECT().GetSettings(gomock.Any()).Return(domain.PrayerSettings{}, errSettingsNotFound)
	repo.EXPECT().FindOverridesBetween(gomock.Any(), "2026-11-01", "2026-12-01").Return(overrides, nil)

	got, err := uc.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if !reflect.DeepEqual(got, domain.PrayerConfig{Settings: domain.DefaultPrayerSettings(), Overrides: overrides}) {
		t.Errorf("config = %+v", got)
	}
}
