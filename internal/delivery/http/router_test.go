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

	iqamah := mock.NewMockIqamahUsecaseItf(ctrl)
	iqamah.EXPECT().GetSettings(gomock.Any()).Return(domain.IqamahSettings{}, nil).AnyTimes()
	iqamah.EXPECT().UpdateSettings(gomock.Any(), gomock.Any()).Return(domain.IqamahSettings{}, nil).AnyTimes()

	fridayDuty := mock.NewMockFridayDutyUsecaseItf(ctrl)
	fridayDuty.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	fridayDuty.EXPECT().ListUpcoming(gomock.Any()).Return(nil, nil).AnyTimes()
	fridayDuty.EXPECT().Save(gomock.Any(), gomock.Any()).Return(domain.FridayDuty{}, nil).AnyTimes()
	fridayDuty.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	slide := mock.NewMockSlideUsecaseItf(ctrl)
	slide.EXPECT().List(gomock.Any()).Return(nil, nil).AnyTimes()
	slide.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.Slide{}, nil).AnyTimes()
	slide.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.Slide{}, nil).AnyTimes()
	slide.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	slide.EXPECT().Reorder(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	slide.EXPECT().GetImage(gomock.Any(), gomock.Any()).Return(domain.SlideImage{ContentType: "image/png"}, nil).AnyTimes()
	slide.EXPECT().ListDisplay(gomock.Any()).Return(nil, nil).AnyTimes()

	screen := mock.NewMockScreenUsecaseItf(ctrl)
	screen.EXPECT().List(gomock.Any()).Return(nil, nil).AnyTimes()
	// a closed event channel ends the stream right away
	closedEvents := make(chan domain.ScreenEvent)
	close(closedEvents)
	screen.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(domain.ScreenConnection{Events: closedEvents, Close: func() {}}, nil).AnyTimes()
	screen.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.Screen{}, nil).AnyTimes()
	screen.EXPECT().Block(gomock.Any(), gomock.Any()).Return(domain.Screen{}, nil).AnyTimes()
	screen.EXPECT().Unblock(gomock.Any(), gomock.Any()).Return(domain.Screen{}, nil).AnyTimes()
	screen.EXPECT().SendCommand(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	screen.EXPECT().NotifyContentChanged().AnyTimes()

	activityLog := mock.NewMockActivityLogUsecaseItf(ctrl)
	activityLog.EXPECT().ListRecent(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	finance := mock.NewMockFinanceUsecaseItf(ctrl)
	finance.EXPECT().ListAccounts(gomock.Any()).Return(nil, nil).AnyTimes()
	finance.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(domain.Account{}, nil).AnyTimes()
	finance.EXPECT().UpdateAccount(gomock.Any(), gomock.Any()).Return(domain.Account{}, nil).AnyTimes()
	finance.EXPECT().ArchiveAccount(gomock.Any(), gomock.Any()).Return(domain.Account{}, nil).AnyTimes()
	finance.EXPECT().RestoreAccount(gomock.Any(), gomock.Any()).Return(domain.Account{}, nil).AnyTimes()
	finance.EXPECT().ListPockets(gomock.Any()).Return(nil, nil).AnyTimes()
	finance.EXPECT().CreatePocket(gomock.Any(), gomock.Any()).Return(domain.Pocket{}, nil).AnyTimes()
	finance.EXPECT().UpdatePocket(gomock.Any(), gomock.Any()).Return(domain.Pocket{}, nil).AnyTimes()
	finance.EXPECT().ArchivePocket(gomock.Any(), gomock.Any()).Return(domain.Pocket{}, nil).AnyTimes()
	finance.EXPECT().RestorePocket(gomock.Any(), gomock.Any()).Return(domain.Pocket{}, nil).AnyTimes()
	finance.EXPECT().ListCategories(gomock.Any()).Return(nil, nil).AnyTimes()
	finance.EXPECT().CreateCategory(gomock.Any(), gomock.Any()).Return(domain.Category{}, nil).AnyTimes()
	finance.EXPECT().UpdateCategory(gomock.Any(), gomock.Any()).Return(domain.Category{}, nil).AnyTimes()
	finance.EXPECT().DeleteCategory(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	finance.EXPECT().ArchiveCategory(gomock.Any(), gomock.Any()).Return(domain.Category{}, nil).AnyTimes()
	finance.EXPECT().RestoreCategory(gomock.Any(), gomock.Any()).Return(domain.Category{}, nil).AnyTimes()
	finance.EXPECT().GetBookkeeping(gomock.Any()).Return(domain.Bookkeeping{}, nil).AnyTimes()
	finance.EXPECT().StartBookkeeping(gomock.Any(), gomock.Any()).Return(domain.Bookkeeping{}, nil).AnyTimes()
	finance.EXPECT().Summary(gomock.Any(), gomock.Any()).Return(domain.FinanceSummary{}, nil).AnyTimes()
	finance.EXPECT().ListTransactions(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	finance.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	finance.EXPECT().UpdateTransaction(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	finance.EXPECT().DeleteTransaction(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	finance.EXPECT().ListRevisions(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	finance.EXPECT().GetReceipt(gomock.Any(), gomock.Any()).Return(domain.TransactionPhoto{ContentType: "image/png"}, nil).AnyTimes()
	finance.EXPECT().SaveReceipt(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	finance.EXPECT().MoveFunds(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	finance.EXPECT().FinishPocket(gomock.Any(), gomock.Any()).Return(domain.Pocket{}, nil).AnyTimes()
	finance.EXPECT().ListClosings(gomock.Any()).Return(nil, nil).AnyTimes()
	finance.EXPECT().CloseMonth(gomock.Any(), gomock.Any()).Return(domain.BookClosing{}, nil).AnyTimes()
	finance.EXPECT().GetTVSetting(gomock.Any()).Return(domain.FinanceTVSetting{}, nil).AnyTimes()
	finance.EXPECT().UpdateTVSetting(gomock.Any(), gomock.Any()).Return(domain.FinanceTVSetting{}, nil).AnyTimes()
	finance.EXPECT().GetDisplay(gomock.Any()).Return(domain.FinanceDisplay{}, nil).AnyTimes()

	return newEcho(controller.NewController(users, auth, mosqueProfile, prayerTime, runningText, iqamah, fridayDuty, slide, screen, activityLog, finance), middleware.NewAuth(authenticator), nil)
}

// TestRouteAccess pins who may call each route, so a route added without its
// ACL fails here
func TestRouteAccess(t *testing.T) {
	const (
		public         = "public"
		anyUser        = "any logged in user"
		superOnly      = "super admin only"
		superTakmir    = "super admin or takmir"
		superBendahara = "super admin or bendahara"
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
		{http.MethodGet, "/api/public/v1/iqamah/config", "", public, ""},
		{http.MethodGet, "/api/public/v1/friday-duties", "", public, ""},
		{http.MethodGet, "/api/public/v1/slides", "", public, ""},
		{http.MethodGet, "/api/public/v1/slides/1/image", "", public, ""},
		{http.MethodGet, "/api/public/v1/screens/device-0001/events", "", public, ""},
		{http.MethodPost, "/api/v1/auth/login", `{}`, public, ""},
		{http.MethodPost, "/api/v1/auth/logout", "", anyUser, ""},
		{http.MethodGet, "/api/v1/auth/profile", "", anyUser, ""},
		{http.MethodPut, "/api/v1/auth/password", `{}`, anyUser, ""},
		{http.MethodGet, "/api/v1/activity-logs", "", anyUser, ""},
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
		{http.MethodGet, "/api/v1/iqamah/config", "", anyUser, ""},
		{http.MethodPut, "/api/v1/iqamah/config", `{}`, superTakmir, ""},
		{http.MethodGet, "/api/v1/friday-duties?from=2026-11-01&to=2026-11-30", "", anyUser, ""},
		{http.MethodPut, "/api/v1/friday-duties/2026-11-06", `{}`, superTakmir, ""},
		{http.MethodDelete, "/api/v1/friday-duties/2026-11-06", "", superTakmir, ""},
		{http.MethodGet, "/api/v1/slides", "", anyUser, ""},
		{http.MethodPost, "/api/v1/slides", upload.String(), superTakmir, form.FormDataContentType()},
		{http.MethodPut, "/api/v1/slides/order", `{}`, superTakmir, ""},
		{http.MethodPut, "/api/v1/slides/1", `{}`, superTakmir, ""},
		{http.MethodDelete, "/api/v1/slides/1", "", superTakmir, ""},
		{http.MethodGet, "/api/v1/screens", "", superTakmir, ""},
		{http.MethodPost, "/api/v1/screens/commands", `{}`, superTakmir, ""},
		{http.MethodPut, "/api/v1/screens/1", `{}`, superTakmir, ""},
		{http.MethodPost, "/api/v1/screens/1/block", "", superTakmir, ""},
		{http.MethodDelete, "/api/v1/screens/1/block", "", superTakmir, ""},
		{http.MethodPost, "/api/v1/screens/1/commands", `{}`, superTakmir, ""},
		{http.MethodGet, "/api/v1/finance/accounts", "", anyUser, ""},
		{http.MethodPost, "/api/v1/finance/accounts", `{}`, superBendahara, ""},
		{http.MethodPut, "/api/v1/finance/accounts/1", `{}`, superBendahara, ""},
		{http.MethodPost, "/api/v1/finance/accounts/1/archive", "", superBendahara, ""},
		{http.MethodDelete, "/api/v1/finance/accounts/1/archive", "", superBendahara, ""},
		{http.MethodGet, "/api/v1/finance/pockets", "", anyUser, ""},
		{http.MethodPost, "/api/v1/finance/pockets", `{}`, superBendahara, ""},
		{http.MethodPut, "/api/v1/finance/pockets/1", `{}`, superBendahara, ""},
		{http.MethodPost, "/api/v1/finance/pockets/1/archive", "", superBendahara, ""},
		{http.MethodDelete, "/api/v1/finance/pockets/1/archive", "", superBendahara, ""},
		{http.MethodGet, "/api/v1/finance/categories", "", anyUser, ""},
		{http.MethodPost, "/api/v1/finance/categories", `{}`, superBendahara, ""},
		{http.MethodPut, "/api/v1/finance/categories/1", `{}`, superBendahara, ""},
		{http.MethodDelete, "/api/v1/finance/categories/1", "", superBendahara, ""},
		{http.MethodPost, "/api/v1/finance/categories/1/archive", "", superBendahara, ""},
		{http.MethodDelete, "/api/v1/finance/categories/1/archive", "", superBendahara, ""},
		{http.MethodGet, "/api/v1/finance/bookkeeping", "", anyUser, ""},
		{http.MethodPost, "/api/v1/finance/bookkeeping", `{}`, superBendahara, ""},
		{http.MethodGet, "/api/v1/finance/summary", "", anyUser, ""},
		{http.MethodGet, "/api/v1/finance/transactions?month=2026-11", "", anyUser, ""},
		{http.MethodPost, "/api/v1/finance/transactions", `{}`, superBendahara, ""},
		{http.MethodPut, "/api/v1/finance/transactions/1", `{}`, superBendahara, ""},
		{http.MethodDelete, "/api/v1/finance/transactions/1", `{}`, superBendahara, ""},
		{http.MethodGet, "/api/v1/finance/transactions/1/revisions", "", anyUser, ""},
		{http.MethodGet, "/api/v1/finance/transactions/1/receipt", "", anyUser, ""},
		{http.MethodPut, "/api/v1/finance/transactions/1/receipt", upload.String(), superBendahara, form.FormDataContentType()},
		{http.MethodPost, "/api/v1/finance/transfers", `{}`, superBendahara, ""},
		{http.MethodPut, "/api/v1/finance/transfers/abc", `{}`, superBendahara, ""},
		{http.MethodPost, "/api/v1/finance/pockets/1/finish", `{}`, superBendahara, ""},
		{http.MethodGet, "/api/v1/finance/closings", "", anyUser, ""},
		{http.MethodPost, "/api/v1/finance/closings", `{}`, superBendahara, ""},
		{http.MethodGet, "/api/v1/finance/tv-setting", "", anyUser, ""},
		{http.MethodPut, "/api/v1/finance/tv-setting", `{}`, superBendahara, ""},
		{http.MethodGet, "/api/public/v1/finance/display", "", public, ""},
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
				case route.access == superBendahara && caller.role == domain.UserRoleBendahara:
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
	slide5MB, slideType := upload(5 << 20)
	slide7MB, _ := upload(7 << 20)

	tests := []struct {
		name, method, path, body, contentType string
		want                                  int
	}{
		{"JSON over the default limit", http.MethodPut, "/api/v1/mosque-profile", `{"name":"` + strings.Repeat("a", 2<<20) + `"}`, echo.MIMEApplicationJSON, http.StatusRequestEntityTooLarge},
		{"logo of 2 MB", http.MethodPut, "/api/v1/mosque-profile/logo", logo2MB, logoType, http.StatusOK},
		{"logo over the upload limit", http.MethodPut, "/api/v1/mosque-profile/logo", logo4MB, logoType, http.StatusRequestEntityTooLarge},
		{"slide of 5 MB", http.MethodPost, "/api/v1/slides", slide5MB, slideType, http.StatusCreated},
		{"slide over the upload limit", http.MethodPost, "/api/v1/slides", slide7MB, slideType, http.StatusRequestEntityTooLarge},
	}

	e := newTestRouter(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
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

// TestCORS answers a preflight only for the configured web clients
func TestCORS(t *testing.T) {
	ctrl := controller.NewController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	e := newEcho(ctrl, middleware.NewAuth(nil), []string{"https://admin.example.app"})

	preflight := func(origin string) string {
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/finance/summary", nil)
		req.Header.Set(echo.HeaderOrigin, origin)
		req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Header().Get(echo.HeaderAccessControlAllowOrigin)
	}

	if got := preflight("https://admin.example.app"); got != "https://admin.example.app" {
		t.Errorf("allowed origin got %q", got)
	}
	if got := preflight("https://evil.example"); got != "" {
		t.Errorf("unknown origin got %q", got)
	}
}
