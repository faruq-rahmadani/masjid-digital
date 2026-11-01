package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"
)

// loggedInAs is the session behind "Bearer valid-token" in these tests
var loggedInAs = &domain.AuthSession{ID: "abc", UserID: 7, Role: domain.UserRoleSuperAdmin}

// authenticated wraps handler in RequireAuth with "Bearer valid-token" accepted as
// loggedInAs, the way the router registers protected routes
func authenticated(t *testing.T, handler echo.HandlerFunc) echo.HandlerFunc {
	t.Helper()

	auth := mock.NewMockAuthenticatorItf(gomock.NewController(t))
	auth.EXPECT().Authenticate(gomock.Any(), "valid-token").Return(loggedInAs, nil).AnyTimes()
	auth.EXPECT().Authenticate(gomock.Any(), gomock.Not("valid-token")).
		Return(nil, apperror.New(apperror.CodeUnauthenticated, errors.New("unknown token"), "sesi tidak valid, silakan login ulang")).AnyTimes()

	return middleware.NewAuth(auth).RequireAuth(handler)
}

// serve runs handler for a request with an optional JSON body, Authorization
// header and path parameters given as name, value pairs
func serve(handler echo.HandlerFunc, method, path, body, authorization string, pathParams ...string) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	if authorization != "" {
		req.Header.Set(echo.HeaderAuthorization, authorization)
	}

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	var names, values []string
	for i := 0; i+1 < len(pathParams); i += 2 {
		names, values = append(names, pathParams[i]), append(values, pathParams[i+1])
	}
	c.SetParamNames(names...)
	c.SetParamValues(values...)

	return rec, handler(c)
}

func TestLogin(t *testing.T) {
	auth := mock.NewMockAuthUsecaseItf(gomock.NewController(t))
	auth.EXPECT().
		Login(gomock.Any(), &domain.LoginRequest{Identifier: "081234567890", Password: "bismillah123"}).
		Return(&domain.LoginResponse{Token: "signed-token", ExpireIn: 86400}, nil)

	rec, err := serve(NewController(nil, auth, nil, nil, nil).Login, http.MethodPost, "/api/v1/auth/login", `{"identifier":"081234567890","password":"bismillah123"}`, "")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"token":"signed-token","expire_in":86400}` {
		t.Errorf("body = %s", got)
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	// no EXPECT: a malformed body must not reach the usecase
	auth := mock.NewMockAuthUsecaseItf(gomock.NewController(t))

	_, err := serve(NewController(nil, auth, nil, nil, nil).Login, http.MethodPost, "/api/v1/auth/login", `{"identifier":`, "")
	assertCode(t, err, apperror.CodeInvalidInput)
}

func TestLoginUsecaseError(t *testing.T) {
	want := apperror.New(apperror.CodeUnauthenticated, errors.New("invalid credentials"), "nomor HP/email atau password salah")

	auth := mock.NewMockAuthUsecaseItf(gomock.NewController(t))
	auth.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, want)

	_, err := serve(NewController(nil, auth, nil, nil, nil).Login, http.MethodPost, "/api/v1/auth/login", `{"identifier":"081234567890","password":"wrong"}`, "")
	if err != want {
		t.Errorf("err = %v, want the usecase error returned as is", err)
	}
}

func TestLogout(t *testing.T) {
	auth := mock.NewMockAuthUsecaseItf(gomock.NewController(t))
	auth.EXPECT().Logout(gomock.Any(), loggedInAs).Return(nil)

	rec, err := serve(authenticated(t, NewController(nil, auth, nil, nil, nil).Logout), http.MethodPost, "/api/v1/auth/logout", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"message":"logout berhasil"}` {
		t.Errorf("response = %d %s", rec.Code, rec.Body.String())
	}
}

func TestLogoutUnauthenticated(t *testing.T) {
	// no EXPECT: without a valid session the handler must not run
	auth := mock.NewMockAuthUsecaseItf(gomock.NewController(t))

	_, err := serve(authenticated(t, NewController(nil, auth, nil, nil, nil).Logout), http.MethodPost, "/api/v1/auth/logout", "", "Bearer unknown-token")
	assertCode(t, err, apperror.CodeUnauthenticated)
}

func TestLogoutUsecaseError(t *testing.T) {
	want := apperror.New(apperror.CodeInternal, errors.New("redis down"), "gagal logout")

	auth := mock.NewMockAuthUsecaseItf(gomock.NewController(t))
	auth.EXPECT().Logout(gomock.Any(), loggedInAs).Return(want)

	_, err := serve(authenticated(t, NewController(nil, auth, nil, nil, nil).Logout), http.MethodPost, "/api/v1/auth/logout", "", "Bearer valid-token")
	if err != want {
		t.Errorf("err = %v, want the usecase error returned as is", err)
	}
}

func assertCode(t *testing.T, err error, want apperror.Code) {
	t.Helper()

	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != want {
		t.Errorf("err = %v, want %s", err, want)
	}
}
