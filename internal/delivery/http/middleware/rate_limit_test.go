package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

func TestRateLimit(t *testing.T) {
	e := echo.New()
	handler := RateLimit(time.Hour, 2, ByClientIP)(func(c echo.Context) error {
		return c.NoContent(http.StatusCreated)
	})

	call := func(ip string) error {
		req := httptest.NewRequest(http.MethodPost, "/api/public/v1/inquiries", nil)
		req.RemoteAddr = ip + ":1234"
		return serveLimited(e, handler, req)
	}

	for i := range 2 {
		if err := call("203.0.113.7"); err != nil {
			t.Fatalf("request %d: %v, want allowed within the burst", i+1, err)
		}
	}

	err := call("203.0.113.7")
	if appErr, ok := errors.AsType[*apperror.AppError](err); !ok || appErr.Code != apperror.CodeTooManyRequests {
		t.Errorf("request 3: error = %v, want TOO_MANY_REQUESTS", err)
	}

	if err := call("198.51.100.9"); err != nil {
		t.Errorf("another IP: %v, want allowed", err)
	}
}

func TestRateLimitEveryone(t *testing.T) {
	e := echo.New()
	handler := RateLimit(time.Hour, 1, Everyone)(func(c echo.Context) error {
		return c.NoContent(http.StatusCreated)
	})

	call := func(ip string) error {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.RemoteAddr = ip + ":1234"
		return serveLimited(e, handler, req)
	}

	if err := call("203.0.113.7"); err != nil {
		t.Fatalf("first request: %v", err)
	}
	if err := call("198.51.100.9"); err == nil {
		t.Error("second IP: allowed, want the shared limit to deny it")
	}
}

// serveLimited runs handler and returns its error, including the one the rate limiter hands
// to the error handler (c.Error) instead of returning it
func serveLimited(e *echo.Echo, handler echo.HandlerFunc, req *http.Request) error {
	var handled error
	e.HTTPErrorHandler = func(err error, _ echo.Context) { handled = err }
	if err := handler(e.NewContext(req, httptest.NewRecorder())); err != nil {
		return err
	}
	return handled
}
