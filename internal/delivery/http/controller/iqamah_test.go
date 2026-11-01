package controller

import (
	"net/http"
	"strings"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"go.uber.org/mock/gomock"
)

const defaultIqamahJSON = `{"iqamah_min":{"subuh":15,"dzuhur":10,"ashar":10,"maghrib":7,"isya":10},"sholat_min":{"subuh":15,"dzuhur":12,"ashar":12,"maghrib":12,"isya":15,"jumat":40},"adzan_display_min":4,"sholat_display":"dim","jumat_before_min":60,"khutbah_min":30,"alarm":{"enabled":true,"at_adzan":true,"before_iqamah":true,"sound":"beep","volume":70}}`

func TestGetIqamahSettings(t *testing.T) {
	iqamah := mock.NewMockIqamahUsecaseItf(gomock.NewController(t))
	iqamah.EXPECT().GetSettings(gomock.Any()).Return(domain.DefaultIqamahSettings(), nil)

	rec, err := serve(authenticated(t, NewController(nil, nil, nil, nil, iqamah, nil).GetIqamahSettings), http.MethodGet, "/api/v1/iqamah/config", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":`+defaultIqamahJSON+`}` {
		t.Errorf("body = %s", got)
	}
}

func TestUpdateIqamahSettingsTakesUpdaterFromSession(t *testing.T) {
	iqamah := mock.NewMockIqamahUsecaseItf(gomock.NewController(t))
	iqamah.EXPECT().UpdateSettings(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, req domain.UpdateIqamahSettingsRequest) (domain.IqamahSettings, error) {
			if req.UpdatedBy != loggedInAs.UserID || req.SholatMin.Jumat != 40 || req.Alarm.Volume != 70 {
				t.Errorf("request = %+v", req)
			}
			return req.IqamahSettings, nil
		})

	rec, err := serve(authenticated(t, NewController(nil, nil, nil, nil, iqamah, nil).UpdateIqamahSettings), http.MethodPut, "/api/v1/iqamah/config", defaultIqamahJSON, "Bearer valid-token")
	if err != nil || rec.Code != http.StatusOK {
		t.Fatalf("handler: %d, %v", rec.Code, err)
	}
}

func TestGetIqamahSettingsIsPublic(t *testing.T) {
	iqamah := mock.NewMockIqamahUsecaseItf(gomock.NewController(t))
	iqamah.EXPECT().GetSettings(gomock.Any()).Return(domain.DefaultIqamahSettings(), nil)

	rec, err := serve(NewController(nil, nil, nil, nil, iqamah, nil).GetIqamahSettings, http.MethodGet, "/api/public/v1/iqamah/config", "", "")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":`+defaultIqamahJSON+`}` {
		t.Errorf("body = %s", got)
	}
}
