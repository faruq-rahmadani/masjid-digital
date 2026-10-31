package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/controller"
	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"
)

// newTestRouter serves the real routes. A bearer token equal to a role name logs
// in as that role; usecases succeed with empty results, so a request that passes
// auth and ACL answers 2xx.
func newTestRouter(t *testing.T) *echo.Echo {
	t.Helper()
	ctrl := gomock.NewController(t)

	authenticator := mock.NewMockAuthenticatorItf(ctrl)
	authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, token string) (*domain.AuthSession, error) {
			role := domain.UserRole(token)
			if !role.IsValid() {
				return nil, apperror.New(apperror.CodeUnauthenticated, errors.New("unknown token"), "sesi tidak valid, silakan login ulang")
			}
			return &domain.AuthSession{ID: "session", UserID: 1, Role: role}, nil
		}).AnyTimes()

	users := mock.NewMockUserUsecaseItf(ctrl)
	users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	users.EXPECT().FindAll(gomock.Any()).Return(nil, nil).AnyTimes()
	users.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(domain.User{ID: 1}, nil).AnyTimes()
	users.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.User{ID: 1}, nil).AnyTimes()
	users.EXPECT().ChangePassword(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	users.EXPECT().ResetPassword(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	auth := mock.NewMockAuthUsecaseItf(ctrl)
	auth.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&domain.LoginResponse{}, nil).AnyTimes()
	auth.EXPECT().Logout(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	e := echo.New()
	e.HTTPErrorHandler = newErrorHandler()
	registerRouters(e, controller.NewController(users, auth), middleware.NewAuth(authenticator))
	return e
}

// TestRouteAccess pins who may call each route, so a route added without its
// ACL fails here
func TestRouteAccess(t *testing.T) {
	const (
		public    = "public"
		anyUser   = "any logged in user"
		superOnly = "super admin only"
	)

	routes := []struct {
		method, path, body, access string
	}{
		{http.MethodGet, "/health", "", public},
		{http.MethodPost, "/api/v1/auth/login", `{}`, public},
		{http.MethodPost, "/api/v1/auth/logout", "", anyUser},
		{http.MethodPut, "/api/v1/auth/password", `{}`, anyUser},
		{http.MethodPost, "/api/v1/users", `{}`, superOnly},
		{http.MethodGet, "/api/v1/users", "", superOnly},
		{http.MethodGet, "/api/v1/users/1", "", superOnly},
		{http.MethodPut, "/api/v1/users/1", `{}`, superOnly},
		{http.MethodPut, "/api/v1/users/1/password", `{}`, superOnly},
	}

	callers := []struct {
		name  string
		token string
		role  domain.UserRole
	}{
		{name: "anonymous"},
		{name: "super_admin", token: "super_admin", role: domain.UserRoleSuperAdmin},
		{name: "takmir", token: "takmir", role: domain.UserRoleTakmir},
		{name: "bendahara", token: "bendahara", role: domain.UserRoleBendahara},
	}

	e := newTestRouter(t)

	for _, route := range routes {
		for _, caller := range callers {
			t.Run(route.method+" "+route.path+" as "+caller.name, func(t *testing.T) {
				var want int
				switch {
				case route.access == public:
					want = http.StatusOK
				case caller.token == "":
					want = http.StatusUnauthorized
				case route.access == anyUser || caller.role == domain.UserRoleSuperAdmin:
					want = http.StatusOK
				default:
					want = http.StatusForbidden
				}

				req := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				if caller.token != "" {
					req.Header.Set(echo.HeaderAuthorization, "Bearer "+caller.token)
				}
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)

				// any 2xx means the request reached its handler
				got := rec.Code
				if got >= 200 && got < 300 {
					got = http.StatusOK
				}
				if got != want {
					t.Errorf("status = %d (%s), want %d", rec.Code, strings.TrimSpace(rec.Body.String()), want)
				}
			})
		}
	}
}
