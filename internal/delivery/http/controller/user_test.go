package controller

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

var (
	createdAt = time.Date(2026, 10, 31, 3, 0, 0, 0, time.UTC)

	ahmad = domain.User{
		ID:           1,
		Name:         "Ahmad",
		Phone:        "+6281234567890",
		Email:        "ahmad@example.com",
		PasswordHash: "$argon2id$secret",
		Role:         domain.UserRoleTakmir,
		Status:       domain.UserStatusActive,
		CreatedBy:    7,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	}

	ahmadJSON = `{"id":1,"name":"Ahmad","phone":"+6281234567890","email":"ahmad@example.com","role":"takmir","status":"active",` +
		`"last_login_at":null,"created_by":7,"created_at":"2026-10-31T03:00:00Z","updated_at":"2026-10-31T03:00:00Z"}`
)

func TestCreateUser(t *testing.T) {
	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().Create(gomock.Any(), domain.CreateUserRequest{
		Name:      "Ahmad",
		Phone:     "081234567890",
		Email:     "ahmad@example.com",
		Password:  "bismillah123",
		CreatedBy: loggedInAs.UserID,
	}).Return(nil)

	rec, err := serve(authenticated(t, NewController(users, nil).CreateUser), http.MethodPost, "/api/v1/users",
		`{"name":"Ahmad","phone":"081234567890","email":"ahmad@example.com","password":"bismillah123"}`, "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
}

func TestCreateUserPassesRole(t *testing.T) {
	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().Create(gomock.Any(), gomock.Cond(func(req domain.CreateUserRequest) bool {
		return req.Role == domain.UserRoleSuperAdmin
	})).Return(nil)

	if _, err := serve(authenticated(t, NewController(users, nil).CreateUser), http.MethodPost, "/api/v1/users",
		`{"name":"Ahmad","phone":"081234567890","password":"bismillah123","role":"super_admin"}`, "Bearer valid-token"); err != nil {
		t.Fatalf("handler: %v", err)
	}
}

func TestCreateUserRecordsCreatorFromSession(t *testing.T) {
	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	// created_by in the body must be ignored; the creator is whoever is logged in
	users.EXPECT().Create(gomock.Any(), gomock.Cond(func(req domain.CreateUserRequest) bool {
		return req.CreatedBy == loggedInAs.UserID
	})).Return(nil)

	if _, err := serve(authenticated(t, NewController(users, nil).CreateUser), http.MethodPost, "/api/v1/users",
		`{"name":"Ahmad","phone":"081234567890","password":"bismillah123","created_by":1,"CreatedBy":1}`, "Bearer valid-token"); err != nil {
		t.Fatalf("handler: %v", err)
	}
}

func TestCreateUserRejects(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		authorization string
		wantCode      apperror.Code
	}{
		{name: "not logged in", body: `{"name":"Ahmad"}`, authorization: "Bearer unknown-token", wantCode: apperror.CodeUnauthenticated},
		{name: "invalid JSON", body: `{"name":`, authorization: "Bearer valid-token", wantCode: apperror.CodeInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no EXPECT: the request must not reach the usecase
			users := mock.NewMockUserUsecaseItf(gomock.NewController(t))

			_, err := serve(authenticated(t, NewController(users, nil).CreateUser), http.MethodPost, "/api/v1/users", tt.body, tt.authorization)
			assertCode(t, err, tt.wantCode)
		})
	}
}

func TestCreateUserUsecaseError(t *testing.T) {
	want := apperror.New(apperror.CodeAlreadyExists, domain.ErrUserAlreadyExists, "nomor HP atau email sudah terdaftar")

	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(want)

	_, err := serve(authenticated(t, NewController(users, nil).CreateUser), http.MethodPost, "/api/v1/users",
		`{"name":"Ahmad","phone":"081234567890","password":"bismillah123"}`, "Bearer valid-token")
	if err != want {
		t.Errorf("err = %v, want the usecase error returned as is", err)
	}
}

func TestFindAllUsers(t *testing.T) {
	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().FindAll(gomock.Any()).Return([]domain.User{ahmad}, nil)

	rec, err := serve(authenticated(t, NewController(users, nil).FindAllUsers), http.MethodGet, "/api/v1/users", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":[`+ahmadJSON+`]}` {
		t.Errorf("body = %s", got)
	}
}

func TestFindAllUsersEmpty(t *testing.T) {
	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().FindAll(gomock.Any()).Return(nil, nil)

	rec, err := serve(authenticated(t, NewController(users, nil).FindAllUsers), http.MethodGet, "/api/v1/users", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	// an empty list, not null, so the frontend can iterate it directly
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":[]}` {
		t.Errorf("body = %s, want an empty list", got)
	}
}

func TestFindUserByID(t *testing.T) {
	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().FindByID(gomock.Any(), int64(1)).Return(ahmad, nil)

	rec, err := serve(authenticated(t, NewController(users, nil).FindUserByID), http.MethodGet, "/api/v1/users/1", "", "Bearer valid-token", "id", "1")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	body := strings.TrimSpace(rec.Body.String())
	if body != `{"data":`+ahmadJSON+`}` {
		t.Errorf("body = %s", body)
	}
	if strings.Contains(body, "argon2") || strings.Contains(body, "password") {
		t.Error("response leaks the password hash")
	}
}

func TestFindUserByIDInvalidID(t *testing.T) {
	for _, id := range []string{"abc", "0", "-1", "1.5", ""} {
		t.Run(id, func(t *testing.T) {
			// no EXPECT: an invalid id must not reach the usecase
			users := mock.NewMockUserUsecaseItf(gomock.NewController(t))

			_, err := serve(authenticated(t, NewController(users, nil).FindUserByID), http.MethodGet, "/api/v1/users/"+id, "", "Bearer valid-token", "id", id)
			assertCode(t, err, apperror.CodeInvalidInput)
		})
	}
}

func TestFindUserByIDNotFound(t *testing.T) {
	want := apperror.New(apperror.CodeNotFound, domain.ErrUserNotFound, "user tidak ditemukan")

	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().FindByID(gomock.Any(), int64(99)).Return(domain.User{}, want)

	_, err := serve(authenticated(t, NewController(users, nil).FindUserByID), http.MethodGet, "/api/v1/users/99", "", "Bearer valid-token", "id", "99")
	if err != want {
		t.Errorf("err = %v, want the usecase error returned as is", err)
	}
}

func TestUpdateUser(t *testing.T) {
	updated := ahmad
	updated.Name = "Ahmad Fauzi"

	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().Update(gomock.Any(), domain.UpdateUserRequest{
		ID:        1,
		UpdatedBy: loggedInAs.UserID,
		Name:      "Ahmad Fauzi",
		Phone:     "+6281234567890",
		Email:     "ahmad@example.com",
		Role:      domain.UserRoleTakmir,
		Status:    domain.UserStatusActive,
	}).Return(updated, nil)

	rec, err := serve(authenticated(t, NewController(users, nil).UpdateUser), http.MethodPut, "/api/v1/users/1",
		`{"name":"Ahmad Fauzi","phone":"+6281234567890","email":"ahmad@example.com","role":"takmir","status":"active"}`, "Bearer valid-token", "id", "1")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":`+strings.Replace(ahmadJSON, `"Ahmad"`, `"Ahmad Fauzi"`, 1)+`}` {
		t.Errorf("body = %s", got)
	}
}

func TestUpdateUserIDFromPath(t *testing.T) {
	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	// an id in the body must not redirect the update to another user, nor fake who made it
	users.EXPECT().Update(gomock.Any(), gomock.Cond(func(req domain.UpdateUserRequest) bool {
		return req.ID == 1 && req.UpdatedBy == loggedInAs.UserID
	})).Return(ahmad, nil)

	if _, err := serve(authenticated(t, NewController(users, nil).UpdateUser), http.MethodPut, "/api/v1/users/1",
		`{"id":2,"ID":2,"UpdatedBy":1,"name":"Ahmad","phone":"+6281234567890","role":"takmir","status":"active"}`, "Bearer valid-token", "id", "1"); err != nil {
		t.Fatalf("handler: %v", err)
	}
}

func TestUpdateUserRejects(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		body     string
		wantCode apperror.Code
	}{
		{name: "invalid id", id: "abc", body: `{"name":"Ahmad"}`, wantCode: apperror.CodeInvalidInput},
		{name: "invalid JSON", id: "1", body: `{"name":`, wantCode: apperror.CodeInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no EXPECT: the request must not reach the usecase
			users := mock.NewMockUserUsecaseItf(gomock.NewController(t))

			_, err := serve(authenticated(t, NewController(users, nil).UpdateUser), http.MethodPut, "/api/v1/users/"+tt.id, tt.body, "Bearer valid-token", "id", tt.id)
			assertCode(t, err, tt.wantCode)
		})
	}
}

func TestUpdateUserUsecaseError(t *testing.T) {
	want := apperror.New(apperror.CodeAlreadyExists, domain.ErrUserAlreadyExists, "nomor HP atau email sudah terdaftar")

	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.User{}, want)

	_, err := serve(authenticated(t, NewController(users, nil).UpdateUser), http.MethodPut, "/api/v1/users/1",
		`{"name":"Ahmad","phone":"+6281234567890","role":"takmir","status":"active"}`, "Bearer valid-token", "id", "1")
	if err != want {
		t.Errorf("err = %v, want the usecase error returned as is", err)
	}
}

func TestChangePassword(t *testing.T) {
	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().ChangePassword(gomock.Any(), domain.ChangePasswordRequest{
		UserID:          loggedInAs.UserID,
		CurrentPassword: "bismillah123",
		NewPassword:     "alhamdulillah456",
	}).Return(nil)

	rec, err := serve(authenticated(t, NewController(users, nil).ChangePassword), http.MethodPut, "/api/v1/auth/password",
		`{"current_password":"bismillah123","new_password":"alhamdulillah456"}`, "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"message":"password berhasil diganti"}` {
		t.Errorf("response = %d %s", rec.Code, rec.Body.String())
	}
}

func TestChangePasswordOnlyOwnAccount(t *testing.T) {
	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	// a user id in the body must not change someone else's password
	users.EXPECT().ChangePassword(gomock.Any(), gomock.Cond(func(req domain.ChangePasswordRequest) bool {
		return req.UserID == loggedInAs.UserID
	})).Return(nil)

	if _, err := serve(authenticated(t, NewController(users, nil).ChangePassword), http.MethodPut, "/api/v1/auth/password",
		`{"user_id":1,"UserID":1,"current_password":"bismillah123","new_password":"alhamdulillah456"}`, "Bearer valid-token"); err != nil {
		t.Fatalf("handler: %v", err)
	}
}

func TestChangePasswordRejects(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		authorization string
		wantCode      apperror.Code
	}{
		{name: "not logged in", body: `{}`, authorization: "Bearer unknown-token", wantCode: apperror.CodeUnauthenticated},
		{name: "invalid JSON", body: `{"current_password":`, authorization: "Bearer valid-token", wantCode: apperror.CodeInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no EXPECT: the request must not reach the usecase
			users := mock.NewMockUserUsecaseItf(gomock.NewController(t))

			_, err := serve(authenticated(t, NewController(users, nil).ChangePassword), http.MethodPut, "/api/v1/auth/password", tt.body, tt.authorization)
			assertCode(t, err, tt.wantCode)
		})
	}
}

func TestChangePasswordUsecaseError(t *testing.T) {
	want := apperror.New(apperror.CodeInvalidInput, errors.New("wrong current password"), "password lama salah")

	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().ChangePassword(gomock.Any(), gomock.Any()).Return(want)

	_, err := serve(authenticated(t, NewController(users, nil).ChangePassword), http.MethodPut, "/api/v1/auth/password",
		`{"current_password":"wrong","new_password":"alhamdulillah456"}`, "Bearer valid-token")
	if err != want {
		t.Errorf("err = %v, want the usecase error returned as is", err)
	}
}

func TestResetPassword(t *testing.T) {
	users := mock.NewMockUserUsecaseItf(gomock.NewController(t))
	users.EXPECT().ResetPassword(gomock.Any(), domain.ResetPasswordRequest{UserID: 3, NewPassword: "alhamdulillah456"}).Return(nil)

	rec, err := serve(authenticated(t, NewController(users, nil).ResetPassword), http.MethodPut, "/api/v1/users/3/password",
		`{"new_password":"alhamdulillah456","user_id":9}`, "Bearer valid-token", "id", "3")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"message":"password berhasil direset"}` {
		t.Errorf("response = %d %s", rec.Code, rec.Body.String())
	}
}

func TestResetPasswordRejects(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		body     string
		wantCode apperror.Code
	}{
		{name: "invalid id", id: "abc", body: `{"new_password":"alhamdulillah456"}`, wantCode: apperror.CodeInvalidInput},
		{name: "invalid JSON", id: "3", body: `{"new_password":`, wantCode: apperror.CodeInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no EXPECT: the request must not reach the usecase
			users := mock.NewMockUserUsecaseItf(gomock.NewController(t))

			_, err := serve(authenticated(t, NewController(users, nil).ResetPassword), http.MethodPut, "/api/v1/users/"+tt.id+"/password", tt.body, "Bearer valid-token", "id", tt.id)
			assertCode(t, err, tt.wantCode)
		})
	}
}
