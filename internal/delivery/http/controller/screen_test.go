package controller

import (
	"net/http"
	"strings"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"
)

func screenController(t *testing.T) (*Controller, *mock.MockScreenUsecaseItf) {
	t.Helper()
	screens := mock.NewMockScreenUsecaseItf(gomock.NewController(t))
	return NewController(nil, nil, nil, nil, nil, nil, nil, nil, screens, nil), screens
}

func TestStreamScreenEvents(t *testing.T) {
	ctrl, screens := screenController(t)
	events := make(chan domain.ScreenEvent, 2)
	events <- domain.ScreenEvent{Type: domain.ScreenEventRefresh}
	events <- domain.ScreenEvent{Type: domain.ScreenEventBlocked} // ends the stream
	closed := false

	screens.EXPECT().Connect(gomock.Any(), domain.ConnectScreenRequest{DeviceID: "device-0001", NameHint: "teras", AppVersion: "1.4.0", UserAgent: "TV"}).
		Return(domain.ScreenConnection{
			State:  domain.ScreenState{Name: "TV Teras", ForcedMode: domain.ScreenModeAuto},
			Events: events,
			Close:  func() { closed = true },
		}, nil)

	rec, err := serve(func(c echo.Context) error {
		c.Request().Header.Set("User-Agent", "TV")
		return ctrl.StreamScreenEvents(c)
	}, http.MethodGet, "/api/public/v1/screens/device-0001/events?name=teras&v=1.4.0", "", "", "device_id", "device-0001")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}

	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("content type = %q", got)
	}
	want := "retry: 5000\nevent: state\ndata: {\"name\":\"TV Teras\",\"forced_mode\":\"auto\",\"forced_until\":null}\n\n" +
		"event: refresh\ndata: null\n\n" +
		"event: blocked\ndata: null\n\n"
	if got := rec.Body.String(); got != want {
		t.Errorf("body =\n%s\nwant\n%s", got, want)
	}
	if !closed {
		t.Error("connection not closed")
	}
}

func TestStreamScreenEventsBlocked(t *testing.T) {
	ctrl, screens := screenController(t)
	screens.EXPECT().Connect(gomock.Any(), gomock.Any()).
		Return(domain.ScreenConnection{}, apperror.New(apperror.CodeForbidden, domain.ErrScreenBlocked, "layar ini diblokir"))

	rec, err := serve(ctrl.StreamScreenEvents, http.MethodGet, "/api/public/v1/screens/device-0001/events", "", "", "device_id", "device-0001")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	// 200 with a long retry, so the browser checks back now and then instead of giving up
	if rec.Code != http.StatusOK || rec.Body.String() != "retry: 300000\nevent: blocked\ndata: null\n\n" {
		t.Errorf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestUpdateScreen(t *testing.T) {
	ctrl, screens := screenController(t)
	screens.EXPECT().Update(gomock.Any(), domain.UpdateScreenRequest{ID: 3, Name: "TV Teras", ForcedMode: domain.ScreenModeDark}).
		Return(domain.Screen{ID: 3, Name: "TV Teras"}, nil)

	rec, err := serve(authenticated(t, ctrl.UpdateScreen), http.MethodPut, "/api/v1/screens/3", `{"name":"TV Teras","forced_mode":"dark"}`, "Bearer valid-token", "id", "3")
	if err != nil || !strings.Contains(rec.Body.String(), `"name":"TV Teras"`) {
		t.Errorf("got %s, %v", rec.Body.String(), err)
	}
}

func TestSendScreenCommand(t *testing.T) {
	ctrl, screens := screenController(t)
	screens.EXPECT().SendCommand(gomock.Any(), domain.SendScreenCommandRequest{ScreenID: 3, Command: domain.ScreenCommandReload}).Return(nil)
	screens.EXPECT().SendCommand(gomock.Any(), domain.SendScreenCommandRequest{Command: domain.ScreenCommandTestSound}).Return(nil)

	if _, err := serve(authenticated(t, ctrl.SendScreenCommand), http.MethodPost, "/api/v1/screens/3/commands", `{"command":"reload"}`, "Bearer valid-token", "id", "3"); err != nil {
		t.Errorf("one screen: %v", err)
	}
	if _, err := serve(authenticated(t, ctrl.SendScreenCommand), http.MethodPost, "/api/v1/screens/commands", `{"command":"test_sound"}`, "Bearer valid-token"); err != nil {
		t.Errorf("all screens: %v", err)
	}
}
