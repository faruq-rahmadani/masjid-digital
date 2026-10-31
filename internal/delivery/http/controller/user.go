package controller

//go:generate go tool mockgen -source=user.go -destination=../../../../mock/controller_user.go -package=mock

import (
	"context"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

type UserUsecaseItf interface {
	Create(ctx context.Context, req domain.CreateUserRequest) error
	FindAll(ctx context.Context) ([]domain.User, error)
	FindByID(ctx context.Context, id int64) (domain.User, error)
	Update(ctx context.Context, req domain.UpdateUserRequest) (domain.User, error)
	ChangePassword(ctx context.Context, req domain.ChangePasswordRequest) error
	ResetPassword(ctx context.Context, req domain.ResetPasswordRequest) error
}

// CreateUser handles POST /users
func (ctrl *Controller) CreateUser(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	var req domain.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	// set after Bind so whatever the client sent is overwritten: the creator is always the session's user
	req.CreatedBy = session.UserID

	if err := ctrl.user.Create(c.Request().Context(), req); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, messageResponse{Message: "user berhasil dibuat"})
}

// FindAllUsers handles GET /users
func (ctrl *Controller) FindAllUsers(c echo.Context) error {
	users, err := ctrl.user.FindAll(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: users})
}

// FindUserByID handles GET /users/:id
func (ctrl *Controller) FindUserByID(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}

	user, err := ctrl.user.FindByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: user})
}

// UpdateUser handles PUT /users/:id, replacing every editable field
func (ctrl *Controller) UpdateUser(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	id, err := paramID(c)
	if err != nil {
		return err
	}

	var req domain.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	// set after Bind: the user being updated is always the one in the URL, by the session's user
	req.ID = id
	req.UpdatedBy = session.UserID

	user, err := ctrl.user.Update(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: user})
}

// GetProfile handles GET /auth/profile: the logged in user's own account
func (ctrl *Controller) GetProfile(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	user, err := ctrl.user.FindByID(c.Request().Context(), session.UserID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: user})
}

// ChangePassword handles PUT /auth/password: the logged in user changes their own password
func (ctrl *Controller) ChangePassword(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	var req domain.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	// set after Bind: users can only change their own password here
	req.UserID = session.UserID

	if err := ctrl.user.ChangePassword(c.Request().Context(), req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, messageResponse{Message: "password berhasil diganti"})
}

// ResetPassword handles PUT /users/:id/password: setting another user's password
// without the old one. Until ACL exists, any logged in user can call it.
func (ctrl *Controller) ResetPassword(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}

	var req domain.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	// set after Bind: the user being reset is always the one in the URL
	req.UserID = id

	if err := ctrl.user.ResetPassword(c.Request().Context(), req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, messageResponse{Message: "password berhasil direset"})
}
