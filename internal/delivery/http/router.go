package http

import (
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/controller"
	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

const (
	// defaultBodyLimit caps every request body except uploads
	defaultBodyLimit = "1M"
	// uploadBodyLimit fits a logo of domain.MaxMosqueLogoSize plus the multipart framing
	uploadBodyLimit = "3M"
	// slideUploadBodyLimit fits an image of domain.MaxSlideImageSize plus the multipart framing
	slideUploadBodyLimit = "6M"
)

// largeBodyRoutes are the routes (as registered, with the /api/v1 prefix) that
// set their own body limit, so the default one skips them
var largeBodyRoutes = map[string]bool{
	"/api/v1/mosque-profile/logo": true,
	"/api/v1/slides":              true,
}

// screenContentRoutes are the route prefixes whose changes show on the TV
var screenContentRoutes = []string{
	"/api/v1/mosque-profile",
	"/api/v1/prayer-time",
	"/api/v1/iqamah",
	"/api/v1/friday-duties",
	"/api/v1/running-text", // running-texts and running-text/setting
	"/api/v1/slides",
}

func registerRouters(e *echo.Echo, ctrl *controller.Controller, auth *middleware.Auth) {
	e.GET("/health", healthcheckHandler)

	// public: read only, no login, for screens anyone can see (the TV display)
	public := e.Group("/api/public/v1")
	public.GET("/time", ctrl.GetServerTime)
	public.GET("/mosque-profile", ctrl.GetMosqueProfile)
	public.GET("/mosque-profile/logo", ctrl.GetMosqueLogo)
	public.GET("/prayer-time/config", ctrl.GetPrayerConfig)
	public.GET("/running-texts", ctrl.GetRunningTextDisplay)
	public.GET("/iqamah/config", ctrl.GetIqamahSettings)
	public.GET("/friday-duties", ctrl.ListUpcomingFridayDuties)
	public.GET("/slides", ctrl.GetSlideDisplay)
	public.GET("/slides/:id/image", ctrl.GetSlideImage)               // any slide, so the admin list's <img> needs no token
	public.GET("/screens/:device_id/events", ctrl.StreamScreenEvents) // the TV's live connection (SSE)

	v1 := e.Group("/api/v1")
	v1.POST("/auth/login", ctrl.Login)

	// any logged in user
	v1WithAuth := v1.Group("", auth.RequireAuth)
	// a saved change to anything the TV shows reaches every TV within seconds
	v1WithAuth.Use(middleware.NotifyScreens(ctrl.NotifyScreensContentChanged, screenContentRoutes...))
	v1WithAuth.POST("/auth/logout", ctrl.Logout)
	v1WithAuth.GET("/auth/profile", ctrl.GetProfile)
	v1WithAuth.PUT("/auth/password", ctrl.ChangePassword)

	// activity log: who changed what, read by anyone logged in
	v1WithAuth.GET("/activity-logs", ctrl.ListActivityLogs)

	// user management: super admins only
	v1WithAuth.POST("/users", ctrl.CreateUser, middleware.RequireRole(domain.UserRoleSuperAdmin))
	v1WithAuth.GET("/users", ctrl.FindAllUsers, middleware.RequireRole(domain.UserRoleSuperAdmin))
	v1WithAuth.GET("/users/:id", ctrl.FindUserByID, middleware.RequireRole(domain.UserRoleSuperAdmin))
	v1WithAuth.PUT("/users/:id", ctrl.UpdateUser, middleware.RequireRole(domain.UserRoleSuperAdmin))
	v1WithAuth.PUT("/users/:id/password", ctrl.ResetPassword, middleware.RequireRole(domain.UserRoleSuperAdmin))

	// mosque profile: read by anyone logged in, changed by super admins and takmir
	v1WithAuth.GET("/mosque-profile", ctrl.GetMosqueProfile)
	v1WithAuth.PUT("/mosque-profile", ctrl.UpdateMosqueProfile, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.GET("/mosque-profile/logo", ctrl.GetMosqueLogo)
	v1WithAuth.PUT("/mosque-profile/logo", ctrl.UpdateMosqueLogo, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir), echomiddleware.BodyLimit(uploadBodyLimit))
	v1WithAuth.DELETE("/mosque-profile/logo", ctrl.DeleteMosqueLogo, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))

	// prayer times: computed on each device; the settings and date overrides
	v1WithAuth.GET("/prayer-time/config", ctrl.GetPrayerSettings)
	v1WithAuth.PUT("/prayer-time/config", ctrl.UpdatePrayerSettings, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.GET("/prayer-time/overrides", ctrl.ListPrayerOverrides)
	v1WithAuth.POST("/prayer-time/overrides", ctrl.CreatePrayerOverride, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.DELETE("/prayer-time/overrides/:id", ctrl.DeletePrayerOverride, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))

	// running text
	v1WithAuth.GET("/running-texts", ctrl.ListRunningTexts)
	v1WithAuth.POST("/running-texts", ctrl.CreateRunningText, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.PUT("/running-texts/order", ctrl.ReorderRunningTexts, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.PUT("/running-texts/:id", ctrl.UpdateRunningText, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.DELETE("/running-texts/:id", ctrl.DeleteRunningText, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.GET("/running-text/setting", ctrl.GetRunningTextSetting)
	v1WithAuth.PUT("/running-text/setting", ctrl.UpdateRunningTextSetting, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))

	// iqamah, mode sholat and alarm: read by anyone logged in, changed by super admins and takmir
	v1WithAuth.GET("/iqamah/config", ctrl.GetIqamahSettings)
	v1WithAuth.PUT("/iqamah/config", ctrl.UpdateIqamahSettings, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))

	// petugas Jumat: read by anyone logged in, changed by super admins and takmir
	v1WithAuth.GET("/friday-duties", ctrl.ListFridayDuties)
	v1WithAuth.PUT("/friday-duties/:date", ctrl.SaveFridayDuty, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.DELETE("/friday-duties/:date", ctrl.DeleteFridayDuty, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	// slides: the same on every screen
	v1WithAuth.GET("/slides", ctrl.ListSlides)
	v1WithAuth.POST("/slides", ctrl.CreateSlide, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir), echomiddleware.BodyLimit(slideUploadBodyLimit))
	v1WithAuth.PUT("/slides/order", ctrl.ReorderSlides, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.PUT("/slides/:id", ctrl.UpdateSlide, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.DELETE("/slides/:id", ctrl.DeleteSlide, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))

	// screens (layar TV): super admins and takmir
	v1WithAuth.GET("/screens", ctrl.ListScreens, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.POST("/screens/commands", ctrl.SendScreenCommand, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.PUT("/screens/:id", ctrl.UpdateScreen, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.POST("/screens/:id/block", ctrl.BlockScreen, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.DELETE("/screens/:id/block", ctrl.UnblockScreen, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.POST("/screens/:id/commands", ctrl.SendScreenCommand, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))

	// keuangan (E7): read by anyone logged in, changed by super admins and bendahara
	financeEditor := middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleBendahara)
	finance := v1WithAuth.Group("/finance")
	finance.GET("/accounts", ctrl.ListAccounts)
	finance.POST("/accounts", ctrl.CreateAccount, financeEditor)
	finance.PUT("/accounts/:id", ctrl.UpdateAccount, financeEditor)
	finance.POST("/accounts/:id/archive", ctrl.ArchiveAccount, financeEditor)
	finance.DELETE("/accounts/:id/archive", ctrl.RestoreAccount, financeEditor)
	finance.GET("/pockets", ctrl.ListPockets)
	finance.POST("/pockets", ctrl.CreatePocket, financeEditor)
	finance.PUT("/pockets/:id", ctrl.UpdatePocket, financeEditor)
	finance.POST("/pockets/:id/archive", ctrl.ArchivePocket, financeEditor)
	finance.DELETE("/pockets/:id/archive", ctrl.RestorePocket, financeEditor)
	finance.GET("/categories", ctrl.ListCategories)
	finance.POST("/categories", ctrl.CreateCategory, financeEditor)
	finance.PUT("/categories/:id", ctrl.UpdateCategory, financeEditor)
	finance.DELETE("/categories/:id", ctrl.DeleteCategory, financeEditor)
	finance.POST("/categories/:id/archive", ctrl.ArchiveCategory, financeEditor)
	finance.DELETE("/categories/:id/archive", ctrl.RestoreCategory, financeEditor)
	finance.GET("/bookkeeping", ctrl.GetBookkeeping)
	finance.POST("/bookkeeping", ctrl.StartBookkeeping, financeEditor)
}

func healthcheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "ok"})
}
