package controller

//go:generate go tool mockgen -source=friday_duty.go -destination=../../../../mock/controller_friday_duty.go -package=mock

import (
	"context"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

type FridayDutyUsecaseItf interface {
	List(ctx context.Context, r domain.FridayDutyRange) ([]domain.FridayDuty, error)
	ListUpcoming(ctx context.Context) ([]domain.FridayDuty, error)
	Save(ctx context.Context, req domain.SaveFridayDutyRequest) (domain.FridayDuty, error)
	Delete(ctx context.Context, date string) error
}

// ListFridayDuties handles GET /friday-duties?from=YYYY-MM-DD&to=YYYY-MM-DD
func (ctrl *Controller) ListFridayDuties(c echo.Context) error {
	duties, err := ctrl.fridayDuty.List(c.Request().Context(), domain.FridayDutyRange{
		From: c.QueryParam("from"),
		To:   c.QueryParam("to"),
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: duties})
}

// ListUpcomingFridayDuties handles the public GET /friday-duties: the coming month's petugas
// Jumat, so the TV still shows them offline
func (ctrl *Controller) ListUpcomingFridayDuties(c echo.Context) error {
	duties, err := ctrl.fridayDuty.ListUpcoming(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: duties})
}

// SaveFridayDuty handles PUT /friday-duties/:date, setting that Friday's petugas
func (ctrl *Controller) SaveFridayDuty(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	var req domain.SaveFridayDutyRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	// set after Bind so whatever the client sent is overwritten
	req.UpdatedBy = session.UserID
	req.Date = c.Param("date")

	duty, err := ctrl.fridayDuty.Save(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: duty})
}

// DeleteFridayDuty handles DELETE /friday-duties/:date
func (ctrl *Controller) DeleteFridayDuty(c echo.Context) error {
	if err := ctrl.fridayDuty.Delete(c.Request().Context(), c.Param("date")); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, messageResponse{Message: "petugas Jumat dihapus"})
}
