package http

import (
	"bytes"
	"errors"
	"mime/multipart"
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

// newTestRouter serves the real routes and middlewares. A bearer token equal to a role name logs
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

	mosqueProfile := mock.NewMockMosqueProfileUsecaseItf(ctrl)
	mosqueProfile.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{}, nil).AnyTimes()
	mosqueProfile.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.MosqueProfile{}, nil).AnyTimes()
	mosqueProfile.EXPECT().GetLogo(gomock.Any()).Return(domain.MosqueLogo{}, nil).AnyTimes()
	mosqueProfile.EXPECT().UpdateLogo(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mosqueProfile.EXPECT().DeleteLogo(gomock.Any()).Return(nil).AnyTimes()

	prayerTime := mock.NewMockPrayerTimeUsecaseItf(ctrl)
	prayerTime.EXPECT().GetSettings(gomock.Any()).Return(domain.PrayerSettings{}, nil).AnyTimes()
	prayerTime.EXPECT().UpdateSettings(gomock.Any(), gomock.Any()).Return(domain.PrayerSettings{}, nil).AnyTimes()
	prayerTime.EXPECT().ListOverrides(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	prayerTime.EXPECT().CreateOverride(gomock.Any(), gomock.Any()).Return(domain.PrayerTimeOverride{}, nil).AnyTimes()
	prayerTime.EXPECT().DeleteOverride(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	prayerTime.EXPECT().GetConfig(gomock.Any()).Return(domain.PrayerConfig{}, nil).AnyTimes()

	runningText := mock.NewMockRunningTextUsecaseItf(ctrl)
	runningText.EXPECT().List(gomock.Any()).Return(nil, nil).AnyTimes()
	runningText.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.RunningText{}, nil).AnyTimes()
	runningText.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.RunningText{}, nil).AnyTimes()
	runningText.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	runningText.EXPECT().Reorder(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	runningText.EXPECT().GetSetting(gomock.Any()).Return(domain.RunningTextSetting{}, nil).AnyTimes()
	runningText.EXPECT().UpdateSetting(gomock.Any(), gomock.Any()).Return(domain.RunningTextSetting{}, nil).AnyTimes()
	runningText.EXPECT().GetDisplay(gomock.Any()).Return(domain.RunningTextDisplay{}, nil).AnyTimes()

	return newEcho(controller.NewController(users, auth, mosqueProfile, prayerTime, runningText), middleware.NewAuth(authenticator))
}

// TestRouteAccess pins who may call each route, so a route added without its
// ACL fails here
func TestRouteAccess(t *testing.T) {
	const (
		public      = "public"
		anyUser     = "any logged in user"
		superOnly   = "super admin only"
		superTakmir = "super admin or takmir"
	)

	// a logo upload, so the route reaches its handler
	var upload bytes.Buffer
	form := multipart.NewWriter(&upload)
	part, _ := form.CreateFormFile("file", "logo.png")
	part.Write([]byte("\x89PNG"))
	form.Close()

	routes := []struct {
		method, path, body, access string
		contentType                string // JSON when empty
	}{
		{http.MethodGet, "/health", "", public, ""},
		{http.MethodGet, "/api/public/v1/time", "", public, ""},
		{http.MethodGet, "/api/public/v1/mosque-profile", "", public, ""},
		{http.MethodGet, "/api/public/v1/mosque-profile/logo", "", public, ""},
		{http.MethodGet, "/api/public/v1/prayer-time/config", "", public, ""},
		{http.MethodGet, "/api/public/v1/running-texts", "", public, ""},
		{http.MethodPost, "/api/v1/auth/login", `{}`, public, ""},
		{http.MethodPost, "/api/v1/auth/logout", "", anyUser, ""},
		{http.MethodGet, "/api/v1/auth/profile", "", anyUser, ""},
		{http.MethodPut, "/api/v1/auth/password", `{}`, anyUser, ""},
		{http.MethodPost, "/api/v1/users", `{}`, superOnly, ""},
		{http.MethodGet, "/api/v1/users", "", superOnly, ""},
		{http.MethodGet, "/api/v1/users/1", "", superOnly, ""},
		{http.MethodPut, "/api/v1/users/1", `{}`, superOnly, ""},
		{http.MethodPut, "/api/v1/users/1/password", `{}`, superOnly, ""},
		{http.MethodGet, "/api/v1/mosque-profile", "", anyUser, ""},
		{http.MethodPut, "/api/v1/mosque-profile", `{}`, superTakmir, ""},
		{http.MethodGet, "/api/v1/mosque-profile/logo", "", anyUser, ""},
		{http.MethodPut, "/api/v1/mosque-profile/logo", upload.String(), superTakmir, form.FormDataContentType()},
		{http.MethodDelete, "/api/v1/mosque-profile/logo", "", superTakmir, ""},
		{http.MethodGet, "/api/v1/prayer-time/config", "", anyUser, ""},
		{http.MethodPut, "/api/v1/prayer-time/config", `{}`, superTakmir, ""},
		{http.MethodGet, "/api/v1/prayer-time/overrides?from=2026-11-01&to=2026-11-30", "", anyUser, ""},
		{http.MethodPost, "/api/v1/prayer-time/overrides", `{}`, superTakmir, ""},
		{http.MethodDelete, "/api/v1/prayer-time/overrides/1", "", superTakmir, ""},
		{http.MethodGet, "/api/v1/running-texts", "", anyUser, ""},
		{http.MethodPost, "/api/v1/running-texts", `{}`, superTakmir, ""},
		{http.MethodPut, "/api/v1/running-texts/order", `{}`, superTakmir, ""},
		{http.MethodPut, "/api/v1/running-texts/1", `{}`, superTakmir, ""},
		{http.MethodDelete, "/api/v1/running-texts/1", "", superTakmir, ""},
		{http.MethodGet, "/api/v1/running-text/setting", "", anyUser, ""},
		{http.MethodPut, "/api/v1/running-text/setting", `{}`, superTakmir, ""},
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
				case route.access == superTakmir && caller.role == domain.UserRoleTakmir:
					want = http.StatusOK
				default:
					want = http.StatusForbidden
				}

				req := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
				contentType := route.contentType
				if contentType == "" {
					contentType = echo.MIMEApplicationJSON
				}
				req.Header.Set(echo.HeaderContentType, contentType)
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

// TestBodyLimits checks uploads get a larger body limit than every other route
func TestBodyLimits(t *testing.T) {
	upload := func(size int) (string, string) {
		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		part, _ := form.CreateFormFile("file", "logo.png")
		part.Write(make([]byte, size))
		form.Close()
		return body.String(), form.FormDataContentType()
	}

	logo2MB, logoType := upload(2 << 20)
	logo4MB, _ := upload(4 << 20)

	tests := []struct {
		name, path, body, contentType string
		want                          int
	}{
		{"JSON over the default limit", "/api/v1/mosque-profile", `{"name":"` + strings.Repeat("a", 2<<20) + `"}`, echo.MIMEApplicationJSON, http.StatusRequestEntityTooLarge},
		{"logo of 2 MB", "/api/v1/mosque-profile/logo", logo2MB, logoType, http.StatusOK},
		{"logo over the upload limit", "/api/v1/mosque-profile/logo", logo4MB, logoType, http.StatusRequestEntityTooLarge},
	}

	e := newTestRouter(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, tt.contentType)
			req.Header.Set(echo.HeaderAuthorization, "Bearer takmir")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Errorf("status = %d (%s), want %d", rec.Code, strings.TrimSpace(rec.Body.String()), tt.want)
			}
		})
	}
}
