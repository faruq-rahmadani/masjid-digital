package controller

//go:generate go tool mockgen -source=prayer_time.go -destination=../../../../mock/controller_prayer_time.go -package=mock

import (
	"context"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

type PrayerTimeUsecaseItf interface {
	GetSettings(ctx context.Context) (domain.PrayerSettings, error)
	UpdateSettings(ctx context.Context, req domain.UpdatePrayerSettingsRequest) (domain.PrayerSettings, error)
	ListOverrides(ctx context.Context, r domain.PrayerTimeOverrideRange) ([]domain.PrayerTimeOverride, error)
	CreateOverride(ctx context.Context, req domain.CreatePrayerTimeOverrideRequest) (domain.PrayerTimeOverride, error)
	DeleteOverride(ctx context.Context, id int64) error
	GetConfig(ctx context.Context) (domain.PrayerConfig, error)
}

func (ctrl *Controller) GetPrayerSettings(c echo.Context) error {
	settings, err := ctrl.prayerTime.GetSettings(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: settings})
}

func (ctrl *Controller) UpdatePrayerSettings(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	var req domain.UpdatePrayerSettingsRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	req.UpdatedBy = session.UserID

	settings, err := ctrl.prayerTime.UpdateSettings(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: settings})
}

func (ctrl *Controller) ListPrayerOverrides(c echo.Context) error {
	overrides, err := ctrl.prayerTime.ListOverrides(c.Request().Context(), domain.PrayerTimeOverrideRange{
		From: c.QueryParam("from"),
		To:   c.QueryParam("to"),
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: overrides})
}

func (ctrl *Controller) CreatePrayerOverride(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	var req domain.CreatePrayerTimeOverrideRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	req.CreatedBy = session.UserID

	override, err := ctrl.prayerTime.CreateOverride(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dataResponse{Data: override})
}

func (ctrl *Controller) DeletePrayerOverride(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}

	if err := ctrl.prayerTime.DeleteOverride(c.Request().Context(), id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, messageResponse{Message: "jadwal khusus dihapus"})
}

func (ctrl *Controller) GetPrayerConfig(c echo.Context) error {
	cfg, err := ctrl.prayerTime.GetConfig(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: cfg})
}
