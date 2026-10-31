package http

import (
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/controller"
	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/labstack/echo/v4"
)

func registerRouters(e *echo.Echo, ctrl *controller.Controller, auth *middleware.Auth) {
	e.GET("/health", healthcheckHandler)

	v1 := e.Group("/api/v1")
	v1.POST("/auth/login", ctrl.Login)

	// any logged in user
	v1WithAuth := v1.Group("", auth.RequireAuth)
	v1WithAuth.POST("/auth/logout", ctrl.Logout)
	v1WithAuth.PUT("/auth/password", ctrl.ChangePassword)

	// user management: super admins only
	v1WithAuth.POST("/users", ctrl.CreateUser, middleware.RequireRole(domain.UserRoleSuperAdmin))
	v1WithAuth.GET("/users", ctrl.FindAllUsers, middleware.RequireRole(domain.UserRoleSuperAdmin))
	v1WithAuth.GET("/users/:id", ctrl.FindUserByID, middleware.RequireRole(domain.UserRoleSuperAdmin))
	v1WithAuth.PUT("/users/:id", ctrl.UpdateUser, middleware.RequireRole(domain.UserRoleSuperAdmin))
	v1WithAuth.PUT("/users/:id/password", ctrl.ResetPassword, middleware.RequireRole(domain.UserRoleSuperAdmin))
}

func healthcheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "ok"})
}
