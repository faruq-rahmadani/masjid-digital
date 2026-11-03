package user

import (
	"context"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"go.uber.org/mock/gomock"
)

func TestChangesPublishActivityLogs(t *testing.T) {
	ctx := context.Background()
	uc, deps := newTestUsecaseWithActivityLog(t)
	deps.hasher.EXPECT().Hash(gomock.Any()).Return("hashed", nil).Times(2)

	deps.users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(42), nil)
	deps.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityUserCreated, EntityID: 42, Summary: "menambah pengguna Ahmad Fauzi"})
	if err := uc.Create(ctx, validRequest()); err != nil {
		t.Fatalf("create: %v", err)
	}

	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil)
	deps.users.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(updatedAhmad, nil)
	deps.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityUserUpdated, EntityID: 1, Summary: "mengubah pengguna Ahmad Fauzi"})
	if _, err := uc.Update(ctx, validUpdateRequest()); err != nil {
		t.Fatalf("update: %v", err)
	}

	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil)
	deps.users.EXPECT().UpdatePasswordHash(gomock.Any(), int64(1), "hashed").Return(nil)
	deps.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityUserPasswordReset, EntityID: 1, Summary: "mereset password Ahmad"})
	if err := uc.ResetPassword(ctx, resetPasswordRequest()); err != nil {
		t.Fatalf("reset password: %v", err)
	}
}

func TestFailedChangesPublishNothing(t *testing.T) {
	ctx := context.Background()
	uc, deps := newTestUsecaseWithActivityLog(t) // no Publish expected
	deps.hasher.EXPECT().Hash(gomock.Any()).Return("hashed", nil).AnyTimes()

	deps.users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(0), errAlreadyExists)
	_ = uc.Create(ctx, validRequest())

	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil).Times(2)
	deps.users.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errDatabase)
	deps.users.EXPECT().UpdatePasswordHash(gomock.Any(), gomock.Any(), gomock.Any()).Return(errDatabase)
	_, _ = uc.Update(ctx, validUpdateRequest())
	_ = uc.ResetPassword(ctx, resetPasswordRequest())
}

func TestChangingOwnPasswordIsNotLogged(t *testing.T) {
	uc, deps := newTestUsecaseWithActivityLog(t) // no Publish expected
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil)
	deps.hasher.EXPECT().Verify("bismillah123", "hash").Return(true, nil)
	deps.hasher.EXPECT().Hash("alhamdulillah456").Return("hashed", nil)
	deps.users.EXPECT().UpdatePasswordHash(gomock.Any(), int64(1), "hashed").Return(nil)

	if err := uc.ChangePassword(context.Background(), domain.ChangePasswordRequest{UserID: 1, CurrentPassword: "bismillah123", NewPassword: "alhamdulillah456"}); err != nil {
		t.Fatalf("change password: %v", err)
	}
}
