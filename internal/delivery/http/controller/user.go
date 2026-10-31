package controller

//go:generate go tool mockgen -source=user.go -destination=../../../../mock/controller_user.go -package=mock

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

// userResponse is a user as the API shows it; the password hash never leaves the server
type userResponse struct {
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	Phone       string            `json:"phone"`
	Email       string            `json:"email"`
	Role        domain.UserRole   `json:"role"`
	Status      domain.UserStatus `json:"status"`
	LastLoginAt *time.Time        `json:"last_login_at"` // null until the first login
	CreatedBy   int64             `json:"created_by"`    // 0 when nobody created the user
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newUserResponse(user domain.User) userResponse {
	return userResponse{
		ID:          user.ID,
		Name:        user.Name,
		Phone:       user.Phone,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		CreatedBy:   user.CreatedBy,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// dataResponse wraps a successful payload as {"data": ...}
type dataResponse struct {
	Data any `json:"data"`
}

type messageResponse struct {
	Message string `json:"message"`
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

	res := make([]userResponse, 0, len(users))
	for _, user := range users {
		res = append(res, newUserResponse(user))
	}

	return c.JSON(http.StatusOK, dataResponse{Data: res})
}

// FindUserByID handles GET /users/:id
func (ctrl *Controller) FindUserByID(c echo.Context) error {
	id, err := userIDParam(c)
	if err != nil {
		return err
	}

	user, err := ctrl.user.FindByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: newUserResponse(user)})
}

// UpdateUser handles PUT /users/:id, replacing every editable field
func (ctrl *Controller) UpdateUser(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	id, err := userIDParam(c)
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

	return c.JSON(http.StatusOK, dataResponse{Data: newUserResponse(user)})
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
	id, err := userIDParam(c)
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

// userIDParam reads the :id path parameter as a positive user ID
func userIDParam(c echo.Context) (int64, error) {
	raw := c.Param("id")

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("invalid user id %q", raw), "id user tidak valid")
	}

	return id, nil
}
