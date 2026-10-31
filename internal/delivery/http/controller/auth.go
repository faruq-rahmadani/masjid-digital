package controller

//go:generate go tool mockgen -source=auth.go -destination=../../../../mock/controller_auth.go -package=mock

import (
	"context"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

type AuthUsecaseItf interface {
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error)
	Logout(ctx context.Context, session *domain.AuthSession) error
}

// Login handles POST /auth/login
func (ctrl *Controller) Login(c echo.Context) error {
	var req domain.LoginRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	res, err := ctrl.auth.Login(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// Logout handles POST /auth/logout, ending the session of the token in use
func (ctrl *Controller) Logout(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	if err := ctrl.auth.Logout(c.Request().Context(), session); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, messageResponse{Message: "logout berhasil"})
}
