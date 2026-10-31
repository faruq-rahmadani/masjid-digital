package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"
)

// serve runs a protected handler behind RequireAuth, returning the session it saw
func serve(auth AuthenticatorItf, authorization string) (*domain.AuthSession, error) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authorization != "" {
		req.Header.Set(echo.HeaderAuthorization, authorization)
	}
	c := echo.New().NewContext(req, httptest.NewRecorder())

	var seen *domain.AuthSession
	handler := func(c echo.Context) error {
		session, err := Session(c)
		seen = session
		return err
	}

	err := NewAuth(auth).RequireAuth(handler)(c)
	return seen, err
}

func TestRequireAuth(t *testing.T) {
	session := &domain.AuthSession{ID: "abc", UserID: 7, Role: domain.UserRoleTakmir}

	for _, header := range []string{"Bearer valid-token", "bearer valid-token", "Bearer   valid-token  "} {
		t.Run(header, func(t *testing.T) {
			auth := mock.NewMockAuthenticatorItf(gomock.NewController(t))
			auth.EXPECT().Authenticate(gomock.Any(), "valid-token").Return(session, nil)

			seen, err := serve(auth, header)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if seen != session {
				t.Errorf("handler saw %+v, want the authenticated session", seen)
			}
		})
	}
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	tests := map[string]string{
		"no header":         "",
		"basic auth":        "Basic dXNlcjpwYXNz",
		"bearer only":       "Bearer",
		"bearer with space": "Bearer   ",
		"token only":        "valid-token",
	}

	for name, header := range tests {
		t.Run(name, func(t *testing.T) {
			// no EXPECT: the token must be rejected before reaching the authenticator
			auth := mock.NewMockAuthenticatorItf(gomock.NewController(t))

			seen, err := serve(auth, header)
			assertCode(t, err, apperror.CodeUnauthenticated)
			if seen != nil {
				t.Error("handler ran without a valid session")
			}
		})
	}
}

func TestRequireAuthRejectsInvalidSession(t *testing.T) {
	invalid := apperror.New(apperror.CodeUnauthenticated, errors.New("unknown token"), "sesi tidak valid, silakan login ulang")

	auth := mock.NewMockAuthenticatorItf(gomock.NewController(t))
	auth.EXPECT().Authenticate(gomock.Any(), "unknown-token").Return(nil, invalid)

	seen, err := serve(auth, "Bearer unknown-token")
	if err != invalid {
		t.Errorf("err = %v, want the authenticator error returned as is", err)
	}
	if seen != nil {
		t.Error("handler ran without a valid session")
	}
}

func TestSessionWithoutRequireAuth(t *testing.T) {
	c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())

	_, err := Session(c)
	assertCode(t, err, apperror.CodeInternal)
}

func assertCode(t *testing.T, err error, want apperror.Code) {
	t.Helper()

	appErr, ok := errors.AsType[*apperror.AppError](err)
	if !ok || appErr.Code != want {
		t.Errorf("err = %v, want %s", err, want)
	}
}
