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

// repository errors, shaped like the real repository's AppErrors
var (
	errNotFound      = apperror.New(apperror.CodeNotFound, domain.ErrUserNotFound, "user tidak ditemukan")
	errAlreadyExists = apperror.New(apperror.CodeAlreadyExists, domain.ErrUserAlreadyExists, "nomor HP atau email sudah terdaftar")
	errDatabase      = apperror.New(apperror.CodeInternal, errors.New("db down"), "gagal mengakses data user")
)

type testDeps struct {
	users  *mock.MockUserRepositoryItf
	hasher *mock.MockPasswordHasher
}

func newTestUsecase(t *testing.T) (*UserUsecase, testDeps) {
	t.Helper()

	ctrl := gomock.NewController(t)
	deps := testDeps{
		users:  mock.NewMockUserRepositoryItf(ctrl),
		hasher: mock.NewMockPasswordHasher(ctrl),
	}
	return NewUserUsecase(deps.users, deps.hasher), deps
}

func validRequest() domain.CreateUserRequest {
	return domain.CreateUserRequest{
		Name:      "  Ahmad Fauzi ",
		Phone:     "0812-3456-7890",
		Email:     " Ahmad@Example.com",
		Password:  "bismillah123",
		Role:      domain.UserRoleTakmir,
		CreatedBy: 7,
	}
}

func TestCreate(t *testing.T) {
	uc, deps := newTestUsecase(t)
	deps.hasher.EXPECT().Hash("bismillah123").Return("hashed", nil)
	deps.users.EXPECT().Create(gomock.Any(), domain.User{
		Name:         "Ahmad Fauzi",
		Phone:        "+6281234567890",
		Email:        "ahmad@example.com",
		PasswordHash: "hashed",
		Role:         domain.UserRoleTakmir,
		Status:       domain.UserStatusActive,
		CreatedBy:    7,
	}).Return(int64(42), nil)

	if err := uc.Create(context.Background(), validRequest()); err != nil {
		t.Fatalf("create: %v", err)
	}
}

func TestCreateOptionalFields(t *testing.T) {
	uc, deps := newTestUsecase(t)
	deps.hasher.EXPECT().Hash(gomock.Any()).Return("hashed", nil)
	deps.users.EXPECT().Create(gomock.Any(), gomock.Cond(func(user domain.User) bool {
		return user.Email == "" && user.Role == domain.UserRoleSuperAdmin
	})).Return(int64(42), nil)

	req := validRequest()
	req.Email = ""
	req.Role = domain.UserRoleSuperAdmin

	if err := uc.Create(context.Background(), req); err != nil {
		t.Fatalf("create: %v", err)
	}
}

func TestCreateInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*domain.CreateUserRequest)
	}{
		{name: "empty name", modify: func(r *domain.CreateUserRequest) { r.Name = "   " }},
		{name: "name too long", modify: func(r *domain.CreateUserRequest) { r.Name = strings.Repeat("a", 101) }},
		{name: "invalid phone", modify: func(r *domain.CreateUserRequest) { r.Phone = "12345" }},
		{name: "invalid email", modify: func(r *domain.CreateUserRequest) { r.Email = "ahmad@" }},
		{name: "password too short", modify: func(r *domain.CreateUserRequest) { r.Password = "1234567" }},
		{name: "password too long", modify: func(r *domain.CreateUserRequest) { r.Password = strings.Repeat("a", 129) }},
		{name: "missing role", modify: func(r *domain.CreateUserRequest) { r.Role = "" }},
		{name: "unknown role", modify: func(r *domain.CreateUserRequest) { r.Role = "imam" }},
		{name: "old user role", modify: func(r *domain.CreateUserRequest) { r.Role = "user" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no EXPECT: invalid input must not be hashed or stored
			uc, _ := newTestUsecase(t)

			req := validRequest()
			tt.modify(&req)

			assertCode(t, uc.Create(context.Background(), req), apperror.CodeInvalidInput)
		})
	}
}

func TestCreateRepositoryErrorsPassThrough(t *testing.T) {
	for _, repoErr := range []*apperror.AppError{errAlreadyExists, errDatabase} {
		t.Run(string(repoErr.Code), func(t *testing.T) {
			uc, deps := newTestUsecase(t)
			deps.hasher.EXPECT().Hash(gomock.Any()).Return("hashed", nil)
			deps.users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(0), repoErr)

			if err := uc.Create(context.Background(), validRequest()); err != repoErr {
				t.Errorf("err = %v, want the repository error returned as is", err)
			}
		})
	}
}

func TestCreateHasherFails(t *testing.T) {
	errBoom := errors.New("boom")

	uc, deps := newTestUsecase(t)
	deps.hasher.EXPECT().Hash(gomock.Any()).Return("", errBoom)

	err := uc.Create(context.Background(), validRequest())
	assertCode(t, err, apperror.CodeInternal)
	if !errors.Is(err, errBoom) {
		t.Errorf("err = %v, want it to wrap the cause", err)
	}
}

func TestFindAll(t *testing.T) {
	users := []domain.User{{ID: 1, Name: "Ahmad"}, {ID: 2, Name: "Budi"}}

	uc, deps := newTestUsecase(t)
	deps.users.EXPECT().FindAll(gomock.Any()).Return(users, nil)

	got, err := uc.FindAll(context.Background())
	if err != nil || len(got) != 2 || got[1].Name != "Budi" {
		t.Errorf("find all = %+v, %v", got, err)
	}
}

func TestFindAllRepositoryError(t *testing.T) {
	uc, deps := newTestUsecase(t)
	deps.users.EXPECT().FindAll(gomock.Any()).Return(nil, errDatabase)

	if _, err := uc.FindAll(context.Background()); err != errDatabase {
		t.Errorf("err = %v, want the repository error returned as is", err)
	}
}

func TestFindByID(t *testing.T) {
	uc, deps := newTestUsecase(t)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(2)).Return(domain.User{ID: 2, Name: "Budi"}, nil)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(99)).Return(domain.User{}, errNotFound)

	if got, err := uc.FindByID(context.Background(), 2); err != nil || got.Name != "Budi" {
		t.Errorf("find by id = %+v, %v; want Budi", got, err)
	}
	if _, err := uc.FindByID(context.Background(), 99); err != errNotFound {
		t.Errorf("err = %v, want the repository error returned as is", err)
	}
}

// ahmad is user 1 as stored before an update
var ahmad = domain.User{
	ID:           1,
	Name:         "Ahmad",
	Phone:        "+6281234567890",
	Email:        "ahmad@example.com",
	PasswordHash: "hash",
	Role:         domain.UserRoleTakmir,
	Status:       domain.UserStatusActive,
	CreatedBy:    2,
}

func validUpdateRequest() domain.UpdateUserRequest {
	return domain.UpdateUserRequest{
		ID:     1,
		Name:   "  Ahmad Fauzi ",
		Phone:  "0812-9999-0001",
		Email:  " Fauzi@Example.com",
		Role:   domain.UserRoleSuperAdmin,
		Status: domain.UserStatusActive,
	}
}

// updatedAhmad is ahmad after validUpdateRequest; password hash and creator are untouched
var updatedAhmad = domain.User{
	ID:           1,
	Name:         "Ahmad Fauzi",
	Phone:        "+6281299990001",
	Email:        "fauzi@example.com",
	PasswordHash: "hash",
	Role:         domain.UserRoleSuperAdmin,
	Status:       domain.UserStatusActive,
	CreatedBy:    2,
}

func TestUpdate(t *testing.T) {
	uc, deps := newTestUsecase(t)
	gomock.InOrder(
		deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil),
		deps.users.EXPECT().Update(gomock.Any(), updatedAhmad).Return(nil),
		// read back for the new updated_at
		deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(updatedAhmad, nil),
	)

	got, err := uc.Update(context.Background(), validUpdateRequest())
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got != updatedAhmad {
		t.Errorf("returned %+v, want %+v", got, updatedAhmad)
	}
}

func TestUpdateClearsEmail(t *testing.T) {
	uc, deps := newTestUsecase(t)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil).Times(2)
	deps.users.EXPECT().Update(gomock.Any(), gomock.Cond(func(user domain.User) bool { return user.Email == "" })).Return(nil)

	req := validUpdateRequest()
	req.Email = ""
	if _, err := uc.Update(context.Background(), req); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestUpdateInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*domain.UpdateUserRequest)
	}{
		{name: "empty name", modify: func(r *domain.UpdateUserRequest) { r.Name = " " }},
		{name: "invalid phone", modify: func(r *domain.UpdateUserRequest) { r.Phone = "123" }},
		{name: "invalid email", modify: func(r *domain.UpdateUserRequest) { r.Email = "ahmad@" }},
		{name: "empty role", modify: func(r *domain.UpdateUserRequest) { r.Role = "" }},
		{name: "unknown role", modify: func(r *domain.UpdateUserRequest) { r.Role = "imam" }},
		{name: "empty status", modify: func(r *domain.UpdateUserRequest) { r.Status = "" }},
		{name: "unknown status", modify: func(r *domain.UpdateUserRequest) { r.Status = "deleted" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no EXPECT: invalid input must not reach the repository
			uc, _ := newTestUsecase(t)

			req := validUpdateRequest()
			tt.modify(&req)

			_, err := uc.Update(context.Background(), req)
			assertCode(t, err, apperror.CodeInvalidInput)
		})
	}
}

func TestUpdateUnknownUser(t *testing.T) {
	// no Update EXPECT: nothing is saved for a missing user
	uc, deps := newTestUsecase(t)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(domain.User{}, errNotFound)

	if _, err := uc.Update(context.Background(), validUpdateRequest()); err != errNotFound {
		t.Errorf("err = %v, want the repository error returned as is", err)
	}
}

func TestUpdateOwnRole(t *testing.T) {
	// no Update EXPECT: changing your own role must not be saved
	uc, deps := newTestUsecase(t)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil)

	req := validUpdateRequest()
	req.UpdatedBy = 1

	_, err := uc.Update(context.Background(), req)
	assertCode(t, err, apperror.CodeForbidden)
}

func TestUpdateSelfKeepingRole(t *testing.T) {
	// editing your own profile is fine as long as the role stays
	uc, deps := newTestUsecase(t)
	deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil).Times(2)
	deps.users.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	req := validUpdateRequest()
	req.UpdatedBy = 1
	req.Role = ahmad.Role

	if _, err := uc.Update(context.Background(), req); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestUpdateRepositoryErrorsPassThrough(t *testing.T) {
	for _, repoErr := range []*apperror.AppError{errNotFound, errAlreadyExists, errDatabase} {
		t.Run(string(repoErr.Code), func(t *testing.T) {
			uc, deps := newTestUsecase(t)
			deps.users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil)
			deps.users.EXPECT().Update(gomock.Any(), gomock.Any()).Return(repoErr)

			if _, err := uc.Update(context.Background(), validUpdateRequest()); err != repoErr {
				t.Errorf("err = %v, want the repository error returned as is", err)
			}
		})
	}
}

func assertCode(t *testing.T, err error, want apperror.Code) {
	t.Helper()

	appErr, ok := errors.AsType[*apperror.AppError](err)
	if !ok || appErr.Code != want {
		t.Errorf("err = %v, want %s", err, want)
	}
}
