package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

func changePasswordRequest() domain.ChangePasswordRequest {
	return domain.ChangePasswordRequest{UserID: 1, CurrentPassword: "bismillah123", NewPassword: "alhamdulillah456"}
}

func TestChangePassword(t *testing.T) {
	uc, deps := newTestUsecase(t)
	gomock.InOrder(
		deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil),
		deps.hasher.EXPECT().Verify("bismillah123", ahmad.PasswordHash).Return(true, nil),
		deps.hasher.EXPECT().Hash("alhamdulillah456").Return("new-hash", nil),
		deps.users.EXPECT().UpdatePasswordHash(gomock.Any(), int64(1), "new-hash").Return(nil),
	)

	if err := uc.ChangePassword(context.Background(), changePasswordRequest()); err != nil {
		t.Fatalf("change password: %v", err)
	}
}

func TestChangePasswordWrongCurrentPassword(t *testing.T) {
	// no Hash or UpdatePasswordHash EXPECT: nothing is saved
	uc, deps := newTestUsecase(t)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil)
	deps.hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(false, nil)

	err := uc.ChangePassword(context.Background(), changePasswordRequest())
	assertCode(t, err, apperror.CodeInvalidInput)
}

func TestChangePasswordSameAsCurrent(t *testing.T) {
	uc, deps := newTestUsecase(t)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil)
	deps.hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(true, nil)

	req := changePasswordRequest()
	req.NewPassword = req.CurrentPassword

	assertCode(t, uc.ChangePassword(context.Background(), req), apperror.CodeInvalidInput)
}

func TestChangePasswordInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*domain.ChangePasswordRequest)
	}{
		{name: "missing current password", modify: func(r *domain.ChangePasswordRequest) { r.CurrentPassword = "" }},
		{name: "new password too short", modify: func(r *domain.ChangePasswordRequest) { r.NewPassword = "1234567" }},
		{name: "new password too long", modify: func(r *domain.ChangePasswordRequest) { r.NewPassword = strings.Repeat("a", 129) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no EXPECT: invalid input must not reach the repository or hasher
			uc, _ := newTestUsecase(t)

			req := changePasswordRequest()
			tt.modify(&req)

			assertCode(t, uc.ChangePassword(context.Background(), req), apperror.CodeInvalidInput)
		})
	}
}

func TestChangePasswordErrors(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name     string
		expect   func(testDeps)
		wantErr  error // checked with ==, the repository error passes through as is
		wantCode apperror.Code
	}{
		{name: "user not found", wantErr: errNotFound, expect: func(d testDeps) {
			d.users.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(domain.User{}, errNotFound)
		}},
		{name: "corrupt stored hash", wantCode: apperror.CodeInternal, expect: func(d testDeps) {
			d.users.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(ahmad, nil)
			d.hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(false, errBoom)
		}},
		{name: "hashing fails", wantCode: apperror.CodeInternal, expect: func(d testDeps) {
			d.users.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(ahmad, nil)
			d.hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(true, nil)
			d.hasher.EXPECT().Hash(gomock.Any()).Return("", errBoom)
		}},
		{name: "saving fails", wantErr: errDatabase, expect: func(d testDeps) {
			d.users.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(ahmad, nil)
			d.hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(true, nil)
			d.hasher.EXPECT().Hash(gomock.Any()).Return("new-hash", nil)
			d.users.EXPECT().UpdatePasswordHash(gomock.Any(), gomock.Any(), gomock.Any()).Return(errDatabase)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, deps := newTestUsecase(t)
			tt.expect(deps)

			err := uc.ChangePassword(context.Background(), changePasswordRequest())
			if tt.wantErr != nil && err != tt.wantErr {
				t.Errorf("err = %v, want %v as is", err, tt.wantErr)
			}
			if tt.wantCode != "" {
				assertCode(t, err, tt.wantCode)
			}
		})
	}
}

func resetPasswordRequest() domain.ResetPasswordRequest {
	return domain.ResetPasswordRequest{UserID: 1, NewPassword: "alhamdulillah456"}
}

func TestResetPassword(t *testing.T) {
	uc, deps := newTestUsecase(t)
	hook := mock.NewMockAfterPasswordResetHook(gomock.NewController(t))
	uc.AddAfterPasswordReset(hook)

	gomock.InOrder(
		deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil),
		deps.hasher.EXPECT().Hash("alhamdulillah456").Return("new-hash", nil),
		deps.users.EXPECT().UpdatePasswordHash(gomock.Any(), int64(1), "new-hash").Return(nil),
		hook.EXPECT().AfterPasswordReset(gomock.Any(), ahmad).Return(nil),
	)

	if err := uc.ResetPassword(context.Background(), resetPasswordRequest()); err != nil {
		t.Fatalf("reset password: %v", err)
	}
}

func TestResetPasswordInvalidInput(t *testing.T) {
	// no EXPECT: invalid input must not reach the repository or hasher
	uc, _ := newTestUsecase(t)

	req := resetPasswordRequest()
	req.NewPassword = "short"

	assertCode(t, uc.ResetPassword(context.Background(), req), apperror.CodeInvalidInput)
}

func TestResetPasswordUnknownUser(t *testing.T) {
	// no Hash, UpdatePasswordHash or hook EXPECT: nothing happens for a missing user
	uc, deps := newTestUsecase(t)
	uc.AddAfterPasswordReset(mock.NewMockAfterPasswordResetHook(gomock.NewController(t)))
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(domain.User{}, errNotFound)

	if err := uc.ResetPassword(context.Background(), resetPasswordRequest()); err != errNotFound {
		t.Errorf("err = %v, want the repository error returned as is", err)
	}
}

func TestResetPasswordSaveFailsSkipsHooks(t *testing.T) {
	// no hook EXPECT: hooks only run once the password is saved
	uc, deps := newTestUsecase(t)
	uc.AddAfterPasswordReset(mock.NewMockAfterPasswordResetHook(gomock.NewController(t)))
	deps.users.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(ahmad, nil)
	deps.hasher.EXPECT().Hash(gomock.Any()).Return("new-hash", nil)
	deps.users.EXPECT().UpdatePasswordHash(gomock.Any(), gomock.Any(), gomock.Any()).Return(errDatabase)

	if err := uc.ResetPassword(context.Background(), resetPasswordRequest()); err != errDatabase {
		t.Errorf("err = %v, want the repository error returned as is", err)
	}
}

func TestResetPasswordHookFails(t *testing.T) {
	errBoom := errors.New("redis down")

	uc, deps := newTestUsecase(t)
	hook := mock.NewMockAfterPasswordResetHook(gomock.NewController(t))
	uc.AddAfterPasswordReset(hook)
	deps.users.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(ahmad, nil)
	deps.hasher.EXPECT().Hash(gomock.Any()).Return("new-hash", nil)
	deps.users.EXPECT().UpdatePasswordHash(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	hook.EXPECT().AfterPasswordReset(gomock.Any(), gomock.Any()).Return(errBoom)

	err := uc.ResetPassword(context.Background(), resetPasswordRequest())
	assertCode(t, err, apperror.CodeInternal)
	if !errors.Is(err, errBoom) {
		t.Errorf("err = %v, want it to wrap the hook error", err)
	}
}

func TestRevokeSessionOnPasswordResetHook(t *testing.T) {
	sessions := mock.NewMockSessionRepositoryItf(gomock.NewController(t))
	sessions.EXPECT().Delete(gomock.Any(), int64(1)).Return(nil)

	if err := NewRevokeSessionOnPasswordResetHook(sessions).AfterPasswordReset(context.Background(), ahmad); err != nil {
		t.Fatalf("hook: %v", err)
	}
}

func TestRevokeSessionOnPasswordResetHookFails(t *testing.T) {
	sessions := mock.NewMockSessionRepositoryItf(gomock.NewController(t))
	sessions.EXPECT().Delete(gomock.Any(), int64(1)).Return(errors.New("redis down"))

	if err := NewRevokeSessionOnPasswordResetHook(sessions).AfterPasswordReset(context.Background(), ahmad); err == nil {
		t.Error("session store failure was ignored")
	}
}
