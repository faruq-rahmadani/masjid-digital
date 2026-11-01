package controller

//go:generate go tool mockgen -source=iqamah.go -destination=../../../../mock/controller_iqamah.go -package=mock

import (
	"context"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

type IqamahUsecaseItf interface {
	GetSettings(ctx context.Context) (domain.IqamahSettings, error)
	UpdateSettings(ctx context.Context, req domain.UpdateIqamahSettingsRequest) (domain.IqamahSettings, error)
}

// GetIqamahSettings handles GET /iqamah/config, public for the TV and for logged in users; the
// defaults until first saved
func (ctrl *Controller) GetIqamahSettings(c echo.Context) error {
	settings, err := ctrl.iqamah.GetSettings(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: settings})
}

// UpdateIqamahSettings handles PUT /iqamah/config, replacing every setting
func (ctrl *Controller) UpdateIqamahSettings(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	var req domain.UpdateIqamahSettingsRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	// set after Bind so whatever the client sent is overwritten
	req.UpdatedBy = session.UserID

	settings, err := ctrl.iqamah.UpdateSettings(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: settings})
}
