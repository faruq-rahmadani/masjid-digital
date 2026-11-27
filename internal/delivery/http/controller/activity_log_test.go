package controller

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"go.uber.org/mock/gomock"
)

func TestListActivityLogsPassesLimit(t *testing.T) {
	logs := mock.NewMockActivityLogUsecaseItf(gomock.NewController(t))
	at := time.Date(2026, 11, 3, 1, 0, 0, 0, time.UTC)
	logs.EXPECT().ListRecent(gomock.Any(), domain.ListActivityLogsRequest{Limit: "4"}).
		Return([]domain.ActivityLog{{ID: 1, ActorID: 7, ActorName: "Hasan Basri", Action: domain.ActivityIqamahSettingsUpdated, Summary: "mengubah pengaturan iqamah", CreatedAt: at}}, nil)

	ctrl := NewController(nil, nil, nil, nil, nil, nil, nil, nil, nil, logs, nil)
	rec, err := serve(authenticated(t, ctrl.ListActivityLogs), http.MethodGet, "/api/v1/activity-logs?limit=4", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	want := `{"data":[{"id":1,"actor_id":7,"actor_name":"Hasan Basri","action":"iqamah_settings.updated","entity_id":0,"summary":"mengubah pengaturan iqamah","created_at":"2026-11-03T01:00:00Z"}]}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s\nwant %s", got, want)
	}
}
