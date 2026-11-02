package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestNotifyScreens(t *testing.T) {
	tests := []struct {
		name, method, path string
		status             int
		err                error
		want               bool
	}{
		{"content saved", http.MethodPut, "/api/v1/slides/:id", http.StatusOK, nil, true},
		{"content created", http.MethodPost, "/api/v1/running-texts", http.StatusCreated, nil, true},
		{"content read", http.MethodGet, "/api/v1/slides", http.StatusOK, nil, false},
		{"save failed", http.MethodPut, "/api/v1/slides/:id", http.StatusOK, errors.New("invalid"), false},
		{"not shown on the TV", http.MethodPost, "/api/v1/users", http.StatusCreated, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notified := false
			handler := NotifyScreens(func() { notified = true }, "/api/v1/slides", "/api/v1/running-text")(func(c echo.Context) error {
				if tt.err != nil {
					return tt.err
				}
				return c.NoContent(tt.status)
			})

			c := echo.New().NewContext(httptest.NewRequest(tt.method, "/", nil), httptest.NewRecorder())
			c.SetPath(tt.path)
			_ = handler(c)

			if notified != tt.want {
				t.Errorf("notified = %v, want %v", notified, tt.want)
			}
		})
	}
}
