package controller

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

const defaultSettingsJSON = `{"corrections":{"subuh":0,"terbit":0,"dzuhur":0,"ashar":0,"maghrib":0,"isya":0},"imsak_offset_min":10,"weekly_overrides":[],"hijri_offset_days":0,"hijri_offset_from":""}`

func TestGetPrayerSettings(t *testing.T) {
	prayerTimes := mock.NewMockPrayerTimeUsecaseItf(gomock.NewController(t))
	prayerTimes.EXPECT().GetSettings(gomock.Any()).Return(domain.DefaultPrayerSettings(), nil)

	rec, err := serve(authenticated(t, NewController(nil, nil, nil, prayerTimes).GetPrayerSettings), http.MethodGet, "/api/v1/prayer-time/config", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":`+defaultSettingsJSON+`}` {
		t.Errorf("body = %s", got)
	}
}

func TestUpdatePrayerSettingsTakesUpdaterFromSession(t *testing.T) {
	prayerTimes := mock.NewMockPrayerTimeUsecaseItf(gomock.NewController(t))
	prayerTimes.EXPECT().UpdateSettings(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, req domain.UpdatePrayerSettingsRequest) (domain.PrayerSettings, error) {
			if req.UpdatedBy != loggedInAs.UserID {
				t.Errorf("updated by = %d, want %d", req.UpdatedBy, loggedInAs.UserID)
			}
			if req.Corrections.Maghrib != 2 || len(req.WeeklyOverrides) != 1 {
				t.Errorf("request = %+v", req)
			}
			return req.PrayerSettings, nil
		})

	body := `{"corrections":{"maghrib":2},"imsak_offset_min":10,"weekly_overrides":[{"weekday":"friday","prayer":"dzuhur","time":"12:00","remarks":""}],"updated_by":99}`
	rec, err := serve(authenticated(t, NewController(nil, nil, nil, prayerTimes).UpdatePrayerSettings), http.MethodPut, "/api/v1/prayer-time/config", body, "Bearer valid-token")
	if err != nil || rec.Code != http.StatusOK {
		t.Fatalf("handler: %d, %v", rec.Code, err)
	}
}

func TestCreatePrayerOverride(t *testing.T) {
	created := time.Date(2026, 11, 1, 8, 0, 0, 0, time.UTC)
	prayerTimes := mock.NewMockPrayerTimeUsecaseItf(gomock.NewController(t))
	prayerTimes.EXPECT().CreateOverride(gomock.Any(), domain.CreatePrayerTimeOverrideRequest{
		CreatedBy: loggedInAs.UserID, Date: "2026-11-06", Prayer: domain.PrayerDzuhur, Time: "11:58", Remarks: "Jumat",
	}).Return(domain.PrayerTimeOverride{ID: 4, Date: "2026-11-06", Prayer: domain.PrayerDzuhur, Time: "11:58", Remarks: "Jumat", CreatedBy: loggedInAs.UserID, CreatedAt: created, UpdatedAt: created}, nil)

	body := `{"date":"2026-11-06","prayer":"dzuhur","time":"11:58","remarks":"Jumat"}`
	rec, err := serve(authenticated(t, NewController(nil, nil, nil, prayerTimes).CreatePrayerOverride), http.MethodPost, "/api/v1/prayer-time/overrides", body, "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"id":4`) || !strings.Contains(got, `"remarks":"Jumat"`) {
		t.Errorf("body = %s", got)
	}
}

func TestListPrayerOverridesPassesRange(t *testing.T) {
	prayerTimes := mock.NewMockPrayerTimeUsecaseItf(gomock.NewController(t))
	prayerTimes.EXPECT().ListOverrides(gomock.Any(), domain.PrayerTimeOverrideRange{From: "2026-11-01", To: "2026-11-30"}).Return([]domain.PrayerTimeOverride{}, nil)

	rec, err := serve(authenticated(t, NewController(nil, nil, nil, prayerTimes).ListPrayerOverrides), http.MethodGet, "/api/v1/prayer-time/overrides?from=2026-11-01&to=2026-11-30", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":[]}` {
		t.Errorf("body = %s", got)
	}
}

func TestDeletePrayerOverride(t *testing.T) {
	prayerTimes := mock.NewMockPrayerTimeUsecaseItf(gomock.NewController(t))
	prayerTimes.EXPECT().DeleteOverride(gomock.Any(), int64(4)).Return(nil)

	rec, err := serve(authenticated(t, NewController(nil, nil, nil, prayerTimes).DeletePrayerOverride), http.MethodDelete, "/api/v1/prayer-time/overrides/4", "", "Bearer valid-token", "id", "4")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"message":"jadwal khusus dihapus"}` {
		t.Errorf("body = %s", got)
	}
}

func TestDeletePrayerOverrideBadID(t *testing.T) {
	_, err := serve(authenticated(t, NewController(nil, nil, nil, nil).DeletePrayerOverride), http.MethodDelete, "/api/v1/prayer-time/overrides/x", "", "Bearer valid-token", "id", "x")
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalidInput {
		t.Errorf("err = %v, want INVALID_INPUT", err)
	}
}

func TestGetPrayerConfig(t *testing.T) {
	prayerTimes := mock.NewMockPrayerTimeUsecaseItf(gomock.NewController(t))
	prayerTimes.EXPECT().GetConfig(gomock.Any()).Return(domain.PrayerConfig{Settings: domain.DefaultPrayerSettings(), Overrides: []domain.PrayerTimeOverride{}}, nil)

	rec, err := serve(NewController(nil, nil, nil, prayerTimes).GetPrayerConfig, http.MethodGet, "/api/public/v1/prayer-time/config", "", "")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":{"settings":`+defaultSettingsJSON+`,"overrides":[]}}` {
		t.Errorf("body = %s", got)
	}
}
