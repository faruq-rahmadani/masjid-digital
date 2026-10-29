package http

import (
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/controller"
	"github.com/labstack/echo/v4"
)

func registerRouters(e *echo.Echo, ctrl *controller.Controller) {
	e.GET("/health", healthcheckHandler)
}

func healthcheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "ok"})
}
