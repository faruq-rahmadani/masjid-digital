package controller

import (
	"net/http"
	"strings"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"go.uber.org/mock/gomock"
)

func TestSaveFridayDutyTakesDateFromPath(t *testing.T) {
	duties := mock.NewMockFridayDutyUsecaseItf(gomock.NewController(t))
	duties.EXPECT().Save(gomock.Any(), domain.SaveFridayDutyRequest{
		UpdatedBy: loggedInAs.UserID, Date: "2026-11-06", Khatib: "Ustadz Abdullah", Theme: "Amanah",
	}).Return(domain.FridayDuty{Date: "2026-11-06", Khatib: "Ustadz Abdullah", Theme: "Amanah"}, nil)

	body := `{"khatib":"Ustadz Abdullah","theme":"Amanah"}`
	rec, err := serve(authenticated(t, NewController(nil, nil, nil, nil, nil, duties).SaveFridayDuty), http.MethodPut, "/api/v1/friday-duties/2026-11-06", body, "Bearer valid-token", "date", "2026-11-06")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"khatib":"Ustadz Abdullah"`) {
		t.Errorf("body = %s", got)
	}
}

func TestDeleteFridayDuty(t *testing.T) {
	duties := mock.NewMockFridayDutyUsecaseItf(gomock.NewController(t))
	duties.EXPECT().Delete(gomock.Any(), "2026-11-06").Return(nil)

	rec, err := serve(authenticated(t, NewController(nil, nil, nil, nil, nil, duties).DeleteFridayDuty), http.MethodDelete, "/api/v1/friday-duties/2026-11-06", "", "Bearer valid-token", "date", "2026-11-06")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"message":"petugas Jumat dihapus"}` {
		t.Errorf("body = %s", got)
	}
}

func TestListFridayDutiesPassesTheRange(t *testing.T) {
	duties := mock.NewMockFridayDutyUsecaseItf(gomock.NewController(t))
	duties.EXPECT().List(gomock.Any(), domain.FridayDutyRange{From: "2026-11-01", To: "2026-11-30"}).Return([]domain.FridayDuty{}, nil)

	rec, err := serve(authenticated(t, NewController(nil, nil, nil, nil, nil, duties).ListFridayDuties), http.MethodGet, "/api/v1/friday-duties?from=2026-11-01&to=2026-11-30", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":[]}` {
		t.Errorf("body = %s", got)
	}
}

func TestListUpcomingFridayDutiesIsPublic(t *testing.T) {
	duties := mock.NewMockFridayDutyUsecaseItf(gomock.NewController(t))
	duties.EXPECT().ListUpcoming(gomock.Any()).Return([]domain.FridayDuty{{Date: "2026-11-06", Khatib: "Ustadz Abdullah"}}, nil)

	rec, err := serve(NewController(nil, nil, nil, nil, nil, duties).ListUpcomingFridayDuties, http.MethodGet, "/api/public/v1/friday-duties", "", "")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"date":"2026-11-06"`) {
		t.Errorf("body = %s", got)
	}
}
