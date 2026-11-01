package controller

//go:generate go tool mockgen -source=running_text.go -destination=../../../../mock/controller_running_text.go -package=mock

import (
	"context"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

type RunningTextUsecaseItf interface {
	List(ctx context.Context) ([]domain.RunningText, error)
	Create(ctx context.Context, req domain.CreateRunningTextRequest) (domain.RunningText, error)
	Update(ctx context.Context, req domain.UpdateRunningTextRequest) (domain.RunningText, error)
	Delete(ctx context.Context, id int64) error
	Reorder(ctx context.Context, req domain.ReorderRunningTextsRequest) ([]domain.RunningText, error)
	GetSetting(ctx context.Context) (domain.RunningTextSetting, error)
	UpdateSetting(ctx context.Context, req domain.UpdateRunningTextSettingRequest) (domain.RunningTextSetting, error)
	GetDisplay(ctx context.Context) (domain.RunningTextDisplay, error)
}

func (ctrl *Controller) ListRunningTexts(c echo.Context) error {
	texts, err := ctrl.runningText.List(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: texts})
}

func (ctrl *Controller) CreateRunningText(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	var req domain.CreateRunningTextRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	req.CreatedBy = session.UserID

	text, err := ctrl.runningText.Create(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dataResponse{Data: text})
}

func (ctrl *Controller) UpdateRunningText(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}

	var req domain.UpdateRunningTextRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	req.ID = id

	text, err := ctrl.runningText.Update(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: text})
}

func (ctrl *Controller) DeleteRunningText(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}

	if err := ctrl.runningText.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, messageResponse{Message: "running text dihapus"})
}

func (ctrl *Controller) ReorderRunningTexts(c echo.Context) error {
	var req domain.ReorderRunningTextsRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	texts, err := ctrl.runningText.Reorder(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: texts})
}

func (ctrl *Controller) GetRunningTextSetting(c echo.Context) error {
	setting, err := ctrl.runningText.GetSetting(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: setting})
}

func (ctrl *Controller) UpdateRunningTextSetting(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	var req domain.UpdateRunningTextSettingRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	req.UpdatedBy = session.UserID

	setting, err := ctrl.runningText.UpdateSetting(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: setting})
}

func (ctrl *Controller) GetRunningTextDisplay(c echo.Context) error {
	display, err := ctrl.runningText.GetDisplay(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: display})
}
