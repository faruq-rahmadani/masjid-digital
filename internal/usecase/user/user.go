package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/faruq-rahmadani/masjid-digital/pkg/hasher"
)

type UserUsecase struct {
	users              usecase.UserRepositoryItf
	hasher             hasher.PasswordHasher
	afterUpdate        []AfterUpdateHook
	afterPasswordReset []AfterPasswordResetHook
}

func NewUserUsecase(users usecase.UserRepositoryItf, hasher hasher.PasswordHasher) *UserUsecase {
	return &UserUsecase{users: users, hasher: hasher}
}

// Create validates req and stores a new user, with req.CreatedBy as its creator
func (u *UserUsecase) Create(ctx context.Context, req domain.CreateUserRequest) error {
	if err := req.ValidateAndNormalize(); err != nil {
		return err
	}

	passwordHash, err := u.hasher.Hash(req.Password)
	if err != nil {
		return apperror.New(apperror.CodeInternal, fmt.Errorf("error hashing password: %w", err), "gagal membuat user")
	}

	_, err = u.users.Create(ctx, req.ToUser(passwordHash))
	return err
}

// FindAll returns every user
func (u *UserUsecase) FindAll(ctx context.Context) ([]domain.User, error) {
	return u.users.FindAll(ctx)
}

// FindByID returns the user with the given ID
func (u *UserUsecase) FindByID(ctx context.Context, id int64) (domain.User, error) {
	return u.users.FindByID(ctx, id)
}

// Update replaces the editable fields of the user req.ID, then runs the
// AfterUpdate hooks (e.g. logging the user out when disabled)
func (u *UserUsecase) Update(ctx context.Context, req domain.UpdateUserRequest) (domain.User, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.User{}, err
	}

	before, err := u.FindByID(ctx, req.ID)
	if err != nil {
		return domain.User{}, err
	}

	after := req.ApplyTo(before)

	// otherwise the only super admin could demote themselves and lock everyone out of user management
	if req.IsUpdateLoggedInUser() && after.Role != before.Role {
		return domain.User{}, apperror.New(apperror.CodeForbidden, fmt.Errorf("user %d tried to change their own role", req.ID), "tidak bisa mengubah role sendiri")
	}

	if err := u.users.Update(ctx, after); err != nil {
		return domain.User{}, err
	}

	if err := u.runAfterUpdate(ctx, before, after); err != nil {
		return domain.User{}, apperror.New(apperror.CodeInternal, err, "user tersimpan, tapi gagal menjalankan proses lanjutan")
	}

	// read back so the response carries the new updated_at
	return u.FindByID(ctx, req.ID)
}

// ChangePassword lets a logged in user change their own password. The current
// password must be right; the session in use stays logged in.
func (u *UserUsecase) ChangePassword(ctx context.Context, req domain.ChangePasswordRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	user, err := u.users.FindByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	ok, err := u.hasher.Verify(req.CurrentPassword, user.PasswordHash)
	if err != nil {
		return apperror.New(apperror.CodeInternal, fmt.Errorf("error verifying password of user %d: %w", user.ID, err), "gagal mengganti password")
	}
	if !ok {
		// INVALID_INPUT, not UNAUTHENTICATED: the session is fine, and a 401 would
		// make the frontend log the user out over a typo
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("wrong current password for user %d", user.ID), "password lama salah")
	}

	if req.IsPasswordSame() {
		return apperror.New(apperror.CodeInvalidInput, errors.New("new password equals the current one"), "password baru harus berbeda dari password lama")
	}

	return u.setPassword(ctx, user.ID, req.NewPassword)
}

// ResetPassword sets another user's password without the old one (e.g. they
// forgot it), then runs the AfterPasswordReset hooks (e.g. logging them out)
func (u *UserUsecase) ResetPassword(ctx context.Context, req domain.ResetPasswordRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	user, err := u.users.FindByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	if err := u.setPassword(ctx, user.ID, req.NewPassword); err != nil {
		return err
	}

	if err := u.runAfterPasswordReset(ctx, user); err != nil {
		return apperror.New(apperror.CodeInternal, err, "password tersimpan, tapi gagal menjalankan proses lanjutan")
	}

	return nil
}

// setPassword hashes and stores a new password for the user
func (u *UserUsecase) setPassword(ctx context.Context, userID int64, password string) error {
	passwordHash, err := u.hasher.Hash(password)
	if err != nil {
		return apperror.New(apperror.CodeInternal, fmt.Errorf("error hashing password: %w", err), "gagal mengganti password")
	}

	return u.users.UpdatePasswordHash(ctx, userID, passwordHash)
}
