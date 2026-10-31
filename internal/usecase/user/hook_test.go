package user

import (
	"context"
	"errors"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

// expectUpdate sets up a successful update of ahmad to updatedAhmad
func expectUpdate(deps testDeps) {
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil)
	deps.users.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
}

func TestAfterUpdateHooksRunInOrder(t *testing.T) {
	uc, deps := newTestUsecase(t)
	expectUpdate(deps)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(updatedAhmad, nil)

	ctrl := gomock.NewController(t)
	first, second := mock.NewMockAfterUpdateHook(ctrl), mock.NewMockAfterUpdateHook(ctrl)
	gomock.InOrder(
		first.EXPECT().AfterUpdate(gomock.Any(), ahmad, updatedAhmad).Return(nil),
		second.EXPECT().AfterUpdate(gomock.Any(), ahmad, updatedAhmad).Return(nil),
	)
	uc.AddAfterUpdate(first, second)

	if _, err := uc.Update(context.Background(), validUpdateRequest()); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestAfterUpdateHookErrorStopsTheRest(t *testing.T) {
	errBoom := errors.New("boom")

	uc, deps := newTestUsecase(t)
	expectUpdate(deps)

	ctrl := gomock.NewController(t)
	failing, skipped := mock.NewMockAfterUpdateHook(ctrl), mock.NewMockAfterUpdateHook(ctrl)
	failing.EXPECT().AfterUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Return(errBoom)
	// no EXPECT on skipped: it must not run after a failing hook
	uc.AddAfterUpdate(failing, skipped)

	_, err := uc.Update(context.Background(), validUpdateRequest())
	assertCode(t, err, apperror.CodeInternal)
	if !errors.Is(err, errBoom) {
		t.Errorf("err = %v, want it to wrap the hook error", err)
	}
}

func TestRevokeSessionOnDisableHook(t *testing.T) {
	active := domain.User{ID: 7, Status: domain.UserStatusActive}
	disabled := domain.User{ID: 7, Status: domain.UserStatusDisabled}

	tests := []struct {
		name        string
		before      domain.User
		after       domain.User
		wantRevoked bool
	}{
		{name: "active to disabled", before: active, after: disabled, wantRevoked: true},
		{name: "stays active", before: active, after: active},
		{name: "stays disabled", before: disabled, after: disabled},
		{name: "disabled to active", before: disabled, after: active},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions := mock.NewMockSessionRepositoryItf(gomock.NewController(t))
			if tt.wantRevoked {
				sessions.EXPECT().Delete(gomock.Any(), int64(7)).Return(nil)
			}

			if err := NewRevokeSessionOnDisableHook(sessions).AfterUpdate(context.Background(), tt.before, tt.after); err != nil {
				t.Fatalf("hook: %v", err)
			}
		})
	}
}

func TestRevokeSessionOnDisableHookFails(t *testing.T) {
	sessions := mock.NewMockSessionRepositoryItf(gomock.NewController(t))
	sessions.EXPECT().Delete(gomock.Any(), int64(7)).Return(errors.New("redis down"))

	err := NewRevokeSessionOnDisableHook(sessions).AfterUpdate(context.Background(),
		domain.User{ID: 7, Status: domain.UserStatusActive},
		domain.User{ID: 7, Status: domain.UserStatusDisabled},
	)
	if err == nil {
		t.Error("session store failure was ignored")
	}
}

func TestDisablingUserThroughUpdateRevokesSession(t *testing.T) {
	uc, deps := newTestUsecase(t)
	expectUpdate(deps)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil)

	sessions := mock.NewMockSessionRepositoryItf(gomock.NewController(t))
	sessions.EXPECT().Delete(gomock.Any(), int64(1)).Return(nil)
	uc.AddAfterUpdate(NewRevokeSessionOnDisableHook(sessions))

	req := validUpdateRequest()
	req.Status = domain.UserStatusDisabled
	if _, err := uc.Update(context.Background(), req); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestRevokeSessionOnRoleChangeHook(t *testing.T) {
	takmir := domain.User{ID: 7, Role: domain.UserRoleTakmir, Status: domain.UserStatusActive}
	superAdmin := domain.User{ID: 7, Role: domain.UserRoleSuperAdmin, Status: domain.UserStatusActive}
	renamed := takmir
	renamed.Name = "Ahmad Fauzi"

	tests := []struct {
		name        string
		before      domain.User
		after       domain.User
		wantRevoked bool
	}{
		{name: "promoted", before: takmir, after: superAdmin, wantRevoked: true},
		{name: "demoted", before: superAdmin, after: takmir, wantRevoked: true},
		{name: "same role, other fields changed", before: takmir, after: renamed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions := mock.NewMockSessionRepositoryItf(gomock.NewController(t))
			if tt.wantRevoked {
				sessions.EXPECT().Delete(gomock.Any(), int64(7)).Return(nil)
			}

			if err := NewRevokeSessionOnRoleChangeHook(sessions).AfterUpdate(context.Background(), tt.before, tt.after); err != nil {
				t.Fatalf("hook: %v", err)
			}
		})
	}
}

func TestRevokeSessionOnRoleChangeHookFails(t *testing.T) {
	sessions := mock.NewMockSessionRepositoryItf(gomock.NewController(t))
	sessions.EXPECT().Delete(gomock.Any(), int64(7)).Return(errors.New("redis down"))

	err := NewRevokeSessionOnRoleChangeHook(sessions).AfterUpdate(context.Background(),
		domain.User{ID: 7, Role: domain.UserRoleSuperAdmin},
		domain.User{ID: 7, Role: domain.UserRoleTakmir},
	)
	if err == nil {
		t.Error("session store failure was ignored")
	}
}
