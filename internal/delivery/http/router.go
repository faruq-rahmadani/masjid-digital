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
)

// largeBodyRoutes are the routes (as registered, with the /api/v1 prefix) that
// set their own body limit, so the default one skips them
var largeBodyRoutes = map[string]bool{
	"/api/v1/mosque-profile/logo": true,
}

func registerRouters(e *echo.Echo, ctrl *controller.Controller, auth *middleware.Auth) {
	e.GET("/health", healthcheckHandler)

	// public: read only, no login, for screens anyone can see (the TV display)
	public := e.Group("/api/public/v1")
	public.GET("/time", ctrl.GetServerTime)
	public.GET("/mosque-profile", ctrl.GetMosqueProfile)
	public.GET("/mosque-profile/logo", ctrl.GetMosqueLogo)
	public.GET("/prayer-time/config", ctrl.GetPrayerConfig)
	public.GET("/iqamah/config", ctrl.GetIqamahSettings)
	public.GET("/friday-duties", ctrl.ListUpcomingFridayDuties)

	v1 := e.Group("/api/v1")
	v1.POST("/auth/login", ctrl.Login)

	// any logged in user
	v1WithAuth := v1.Group("", auth.RequireAuth)
	v1WithAuth.POST("/auth/logout", ctrl.Logout)
	v1WithAuth.GET("/auth/profile", ctrl.GetProfile)
	v1WithAuth.PUT("/auth/password", ctrl.ChangePassword)

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

	// iqamah, mode sholat and alarm: read by anyone logged in, changed by super admins and takmir
	v1WithAuth.GET("/iqamah/config", ctrl.GetIqamahSettings)
	v1WithAuth.PUT("/iqamah/config", ctrl.UpdateIqamahSettings, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))

	// petugas Jumat: read by anyone logged in, changed by super admins and takmir
	v1WithAuth.GET("/friday-duties", ctrl.ListFridayDuties)
	v1WithAuth.PUT("/friday-duties/:date", ctrl.SaveFridayDuty, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
	v1WithAuth.DELETE("/friday-duties/:date", ctrl.DeleteFridayDuty, middleware.RequireRole(domain.UserRoleSuperAdmin, domain.UserRoleTakmir))
}

func healthcheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "ok"})
}
