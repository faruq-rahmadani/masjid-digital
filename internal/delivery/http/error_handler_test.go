package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

func TestErrorHandler(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "app error",
			err:        apperror.New(apperror.CodeAlreadyExists, errors.New("phone taken"), "nomor HP sudah terdaftar"),
			wantStatus: http.StatusConflict,
			wantBody:   `{"code":"ALREADY_EXISTS","message":"nomor HP sudah terdaftar"}`,
		},
		{
			name:       "plain error hides internals",
			err:        errors.New("connection refused"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"code":"INTERNAL","message":"Internal Server Error"}`,
		},
		{
			name:       "unmapped code",
			err:        apperror.New("SOMETHING_NEW", nil, "oops"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"code":"SOMETHING_NEW","message":"oops"}`,
		},
		{
			name:       "app error wrapping an echo error",
			err:        apperror.New(apperror.CodeInvalidInput, echo.NewHTTPError(http.StatusBadRequest, "unexpected EOF"), "format request tidak valid"),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"code":"INVALID_INPUT","message":"format request tidak valid"}`,
		},
		{
			name:       "echo not found",
			err:        echo.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   `{"code":"NOT_FOUND","message":"Not Found"}`,
		},
		{
			name:       "echo method not allowed",
			err:        echo.ErrMethodNotAllowed,
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   `{"code":"INVALID_INPUT","message":"Method Not Allowed"}`,
		},
		{
			name:       "echo body too large",
			err:        echo.ErrStatusRequestEntityTooLarge,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody:   `{"code":"INVALID_INPUT","message":"ukuran request terlalu besar"}`,
		},
		{
			name:       "echo internal error hides its message",
			err:        echo.NewHTTPError(http.StatusInternalServerError, "db password is wrong"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"code":"INTERNAL","message":"Internal Server Error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			rec := httptest.NewRecorder()
			c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

			newErrorHandler()(tt.err, c)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
		})
	}
}
