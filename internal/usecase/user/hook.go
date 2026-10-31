package user

//go:generate go tool mockgen -source=hook.go -destination=../../../mock/user_hook.go -package=mock

import (
	"context"
	"fmt"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
)

// AfterUpdateHook runs after a user is saved by Update, with the user as it was
// and as it is now. Hooks run in the order they were added; the first error
// stops the rest and is returned by Update, although the user is already saved.
type AfterUpdateHook interface {
	AfterUpdate(ctx context.Context, before, after domain.User) error
}

// AddAfterUpdate registers hooks to run after every successful Update
func (u *UserUsecase) AddAfterUpdate(hooks ...AfterUpdateHook) {
	u.afterUpdate = append(u.afterUpdate, hooks...)
}

func (u *UserUsecase) runAfterUpdate(ctx context.Context, before, after domain.User) error {
	for _, hook := range u.afterUpdate {
		if err := hook.AfterUpdate(ctx, before, after); err != nil {
			return err
		}
	}
	return nil
}

// RevokeSessionOnDisableHook logs a user out as soon as they are disabled, so
// their token stops working right away instead of when it expires
type RevokeSessionOnDisableHook struct {
	sessions usecase.SessionRepositoryItf
}

func NewRevokeSessionOnDisableHook(sessions usecase.SessionRepositoryItf) *RevokeSessionOnDisableHook {
	return &RevokeSessionOnDisableHook{sessions: sessions}
}

func (h *RevokeSessionOnDisableHook) AfterUpdate(ctx context.Context, before, after domain.User) error {
	if before.IsActive() && !after.IsActive() {
		if err := h.sessions.Delete(ctx, after.ID); err != nil {
			return fmt.Errorf("error revoking session of disabled user %d: %w", after.ID, err)
		}
	}
	return nil
}

// RevokeSessionOnRoleChangeHook logs a user out when their role changes. The
// role travels inside the access token, so without this a demoted user would
// keep their old permissions until the token expires.
type RevokeSessionOnRoleChangeHook struct {
	sessions usecase.SessionRepositoryItf
}

func NewRevokeSessionOnRoleChangeHook(sessions usecase.SessionRepositoryItf) *RevokeSessionOnRoleChangeHook {
	return &RevokeSessionOnRoleChangeHook{sessions: sessions}
}

func (h *RevokeSessionOnRoleChangeHook) AfterUpdate(ctx context.Context, before, after domain.User) error {
	if before.Role != after.Role {
		if err := h.sessions.Delete(ctx, after.ID); err != nil {
			return fmt.Errorf("error revoking session of user %d after role change: %w", after.ID, err)
		}
	}
	return nil
}

// AfterPasswordResetHook runs after an admin resets a user's password with ResetPassword.
// Like AfterUpdateHook, the first error stops the rest and the password is already saved.
type AfterPasswordResetHook interface {
	AfterPasswordReset(ctx context.Context, user domain.User) error
}

// AddAfterPasswordReset registers hooks to run after every successful ResetPassword
func (u *UserUsecase) AddAfterPasswordReset(hooks ...AfterPasswordResetHook) {
	u.afterPasswordReset = append(u.afterPasswordReset, hooks...)
}

func (u *UserUsecase) runAfterPasswordReset(ctx context.Context, user domain.User) error {
	for _, hook := range u.afterPasswordReset {
		if err := hook.AfterPasswordReset(ctx, user); err != nil {
			return err
		}
	}
	return nil
}

// RevokeSessionOnPasswordResetHook logs a user out when an admin resets their
// password, so whoever holds their old session has to log in with the new one
type RevokeSessionOnPasswordResetHook struct {
	sessions usecase.SessionRepositoryItf
}

func NewRevokeSessionOnPasswordResetHook(sessions usecase.SessionRepositoryItf) *RevokeSessionOnPasswordResetHook {
	return &RevokeSessionOnPasswordResetHook{sessions: sessions}
}

func (h *RevokeSessionOnPasswordResetHook) AfterPasswordReset(ctx context.Context, user domain.User) error {
	if err := h.sessions.Delete(ctx, user.ID); err != nil {
		return fmt.Errorf("error revoking session of user %d after password reset: %w", user.ID, err)
	}
	return nil
}
