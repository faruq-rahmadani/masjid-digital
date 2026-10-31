package authentication

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/faruq-rahmadani/masjid-digital/pkg/token"
	"go.uber.org/mock/gomock"
)

// repository errors, shaped like the real repositories' AppErrors
var (
	errUserNotFound    = apperror.New(apperror.CodeNotFound, domain.ErrUserNotFound, "user tidak ditemukan")
	errDatabase        = apperror.New(apperror.CodeInternal, errors.New("db down"), "gagal mengakses data user")
	errSessionNotFound = apperror.New(apperror.CodeUnauthenticated, domain.ErrSessionNotFound, "sesi tidak valid, silakan login ulang")
	errSessionStore    = apperror.New(apperror.CodeInternal, errors.New("redis down"), "gagal mengakses sesi")
)

var loginTime = time.Date(2026, 10, 31, 3, 0, 0, 0, time.UTC)

// ahmad is the registered user the tests log in as
var ahmad = domain.User{
	ID:           7,
	Name:         "Ahmad",
	Phone:        "+6281234567890",
	Email:        "ahmad@example.com",
	PasswordHash: "stored-hash",
	Role:         domain.UserRoleSuperAdmin,
	Status:       domain.UserStatusActive,
}

var issued = token.Issued{Value: "signed-token", ID: "session-1", ExpiresIn: 24 * time.Hour}

type testDeps struct {
	users    *mock.MockUserRepositoryItf
	hasher   *mock.MockPasswordHasher
	tokens   *mock.MockTokenManagerItf
	sessions *mock.MockSessionRepositoryItf
}

func newTestUsecase(t *testing.T) (*AuthenticationUseCase, testDeps) {
	t.Helper()

	ctrl := gomock.NewController(t)
	deps := testDeps{
		users:    mock.NewMockUserRepositoryItf(ctrl),
		hasher:   mock.NewMockPasswordHasher(ctrl),
		tokens:   mock.NewMockTokenManagerItf(ctrl),
		sessions: mock.NewMockSessionRepositoryItf(ctrl),
	}

	uc := New(deps.users, deps.hasher, deps.tokens, deps.sessions)
	uc.now = func() time.Time { return loginTime }
	return uc, deps
}

func loginRequest(identifier, password string) *domain.LoginRequest {
	return &domain.LoginRequest{Identifier: identifier, Password: password}
}

// expectPasswordChecked sets up finding ahmad by phone and checking his password
func expectPasswordChecked(deps testDeps, matches bool) {
	deps.users.EXPECT().GetByPhone(gomock.Any(), ahmad.Phone).Return(ahmad, nil)
	deps.hasher.EXPECT().Verify("bismillah123", ahmad.PasswordHash).Return(matches, nil)
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expectFind func(testDeps)
	}{
		{name: "by phone", identifier: ahmad.Phone, expectFind: func(d testDeps) {
			d.users.EXPECT().GetByPhone(gomock.Any(), ahmad.Phone).Return(ahmad, nil)
		}},
		{name: "by email", identifier: ahmad.Email, expectFind: func(d testDeps) {
			d.users.EXPECT().GetByEmail(gomock.Any(), ahmad.Email).Return(ahmad, nil)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, deps := newTestUsecase(t)
			tt.expectFind(deps)
			deps.hasher.EXPECT().Verify("bismillah123", ahmad.PasswordHash).Return(true, nil)
			deps.tokens.EXPECT().Issue(ahmad.ID, "super_admin").Return(issued, nil)
			deps.sessions.EXPECT().Create(gomock.Any(), ahmad.ID, issued.ID, issued.ExpiresIn).Return(nil)
			deps.users.EXPECT().UpdateLastLoginAt(gomock.Any(), ahmad.ID, loginTime).Return(nil)

			res, err := uc.Login(context.Background(), loginRequest(tt.identifier, "bismillah123"))
			if err != nil {
				t.Fatalf("login: %v", err)
			}
			if res.Token != "signed-token" || res.ExpireIn != 86400 {
				t.Errorf("response = %+v, want the signed token expiring in 86400s", res)
			}
		})
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	tests := []struct {
		name   string
		req    *domain.LoginRequest
		expect func(testDeps)
	}{
		{name: "wrong password", req: loginRequest(ahmad.Phone, "bismillah123"), expect: func(d testDeps) { expectPasswordChecked(d, false) }},
		{name: "unknown phone", req: loginRequest("+6281299999999", "bismillah123"), expect: func(d testDeps) {
			d.users.EXPECT().GetByPhone(gomock.Any(), "+6281299999999").Return(domain.User{}, errUserNotFound)
		}},
		{name: "unknown email", req: loginRequest("budi@example.com", "bismillah123"), expect: func(d testDeps) {
			d.users.EXPECT().GetByEmail(gomock.Any(), "budi@example.com").Return(domain.User{}, errUserNotFound)
		}},
		// rejected before any lookup
		{name: "password too long", req: loginRequest(ahmad.Phone, strings.Repeat("a", maxPasswordLength+1)), expect: func(testDeps) {}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no token, session or last login EXPECTs: a failed login must not create any
			uc, deps := newTestUsecase(t)
			tt.expect(deps)

			_, err := uc.Login(context.Background(), tt.req)
			assertCode(t, err, apperror.CodeUnauthenticated)

			// unknown users and wrong passwords must look the same to the client,
			// not like the repository's "user tidak ditemukan"
			if appErr, _ := errors.AsType[*apperror.AppError](err); appErr.MessageError != "nomor HP/email atau password salah" {
				t.Errorf("message = %q, want the shared invalid credentials message", appErr.MessageError)
			}
		})
	}
}

func TestLoginDisabledUser(t *testing.T) {
	disabled := ahmad
	disabled.Status = domain.UserStatusDisabled

	t.Run("right password", func(t *testing.T) {
		uc, deps := newTestUsecase(t)
		deps.users.EXPECT().GetByPhone(gomock.Any(), ahmad.Phone).Return(disabled, nil)
		deps.hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(true, nil)

		_, err := uc.Login(context.Background(), loginRequest(ahmad.Phone, "bismillah123"))
		assertCode(t, err, apperror.CodeForbidden)
	})

	t.Run("wrong password does not reveal it", func(t *testing.T) {
		uc, deps := newTestUsecase(t)
		deps.users.EXPECT().GetByPhone(gomock.Any(), ahmad.Phone).Return(disabled, nil)
		deps.hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(false, nil)

		_, err := uc.Login(context.Background(), loginRequest(ahmad.Phone, "wrong-password"))
		assertCode(t, err, apperror.CodeUnauthenticated)
	})
}

func TestLoginRepositoryErrorsPassThrough(t *testing.T) {
	tests := []struct {
		name   string
		expect func(testDeps)
		want   error
	}{
		{name: "find user fails", want: errDatabase, expect: func(d testDeps) {
			d.users.EXPECT().GetByPhone(gomock.Any(), gomock.Any()).Return(domain.User{}, errDatabase)
		}},
		{name: "session store fails", want: errSessionStore, expect: func(d testDeps) {
			expectPasswordChecked(d, true)
			d.tokens.EXPECT().Issue(gomock.Any(), gomock.Any()).Return(issued, nil)
			d.sessions.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errSessionStore)
		}},
		{name: "last login update fails", want: errDatabase, expect: func(d testDeps) {
			expectPasswordChecked(d, true)
			d.tokens.EXPECT().Issue(gomock.Any(), gomock.Any()).Return(issued, nil)
			d.sessions.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			d.users.EXPECT().UpdateLastLoginAt(gomock.Any(), gomock.Any(), gomock.Any()).Return(errDatabase)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, deps := newTestUsecase(t)
			tt.expect(deps)

			if _, err := uc.Login(context.Background(), loginRequest(ahmad.Phone, "bismillah123")); err != tt.want {
				t.Errorf("err = %v, want the repository error returned as is", err)
			}
		})
	}
}

func TestLoginInternalErrors(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name   string
		expect func(testDeps)
	}{
		{name: "corrupt stored hash", expect: func(d testDeps) {
			d.users.EXPECT().GetByPhone(gomock.Any(), gomock.Any()).Return(ahmad, nil)
			d.hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(false, errBoom)
		}},
		{name: "token issuer fails", expect: func(d testDeps) {
			expectPasswordChecked(d, true)
			d.tokens.EXPECT().Issue(gomock.Any(), gomock.Any()).Return(token.Issued{}, errBoom)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, deps := newTestUsecase(t)
			tt.expect(deps)

			_, err := uc.Login(context.Background(), loginRequest(ahmad.Phone, "bismillah123"))
			assertCode(t, err, apperror.CodeInternal)
			if !errors.Is(err, errBoom) {
				t.Errorf("err = %v, want it to wrap the cause", err)
			}
		})
	}
}

var claims = token.Claims{ID: "session-1", UserID: 7, Role: "super_admin"}

func TestAuthenticate(t *testing.T) {
	uc, deps := newTestUsecase(t)
	deps.tokens.EXPECT().Parse("signed-token").Return(claims, nil)
	deps.sessions.EXPECT().GetSessionID(gomock.Any(), int64(7)).Return("session-1", nil)

	session, err := uc.Authenticate(context.Background(), "signed-token")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	want := domain.AuthSession{ID: "session-1", UserID: 7, Role: domain.UserRoleSuperAdmin}
	if *session != want {
		t.Errorf("session = %+v, want %+v", *session, want)
	}
}

func TestAuthenticateRejects(t *testing.T) {
	tests := []struct {
		name   string
		expect func(testDeps)
	}{
		{name: "invalid token", expect: func(d testDeps) {
			d.tokens.EXPECT().Parse(gomock.Any()).Return(token.Claims{}, token.ErrInvalidToken)
		}},
		{name: "logged out", expect: func(d testDeps) {
			d.tokens.EXPECT().Parse(gomock.Any()).Return(claims, nil)
			d.sessions.EXPECT().GetSessionID(gomock.Any(), int64(7)).Return("", errSessionNotFound)
		}},
		{name: "logged in on another device", expect: func(d testDeps) {
			d.tokens.EXPECT().Parse(gomock.Any()).Return(claims, nil)
			d.sessions.EXPECT().GetSessionID(gomock.Any(), int64(7)).Return("session-2", nil)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, deps := newTestUsecase(t)
			tt.expect(deps)

			_, err := uc.Authenticate(context.Background(), "signed-token")
			assertCode(t, err, apperror.CodeUnauthenticated)
			if appErr, _ := errors.AsType[*apperror.AppError](err); appErr.MessageError != "sesi tidak valid, silakan login ulang" {
				t.Errorf("message = %q", appErr.MessageError)
			}
		})
	}
}

func TestAuthenticateSessionStoreFails(t *testing.T) {
	uc, deps := newTestUsecase(t)
	deps.tokens.EXPECT().Parse(gomock.Any()).Return(claims, nil)
	deps.sessions.EXPECT().GetSessionID(gomock.Any(), gomock.Any()).Return("", errSessionStore)

	if _, err := uc.Authenticate(context.Background(), "signed-token"); err != errSessionStore {
		t.Errorf("err = %v, want the session store error returned as is", err)
	}
}

func TestLogout(t *testing.T) {
	uc, deps := newTestUsecase(t)
	deps.sessions.EXPECT().Delete(gomock.Any(), int64(7)).Return(nil)

	if err := uc.Logout(context.Background(), &domain.AuthSession{ID: "session-1", UserID: 7}); err != nil {
		t.Fatalf("logout: %v", err)
	}
}

func TestLogoutFails(t *testing.T) {
	uc, deps := newTestUsecase(t)
	deps.sessions.EXPECT().Delete(gomock.Any(), int64(7)).Return(errSessionStore)

	if err := uc.Logout(context.Background(), &domain.AuthSession{ID: "session-1", UserID: 7}); err != errSessionStore {
		t.Errorf("err = %v, want the session store error returned as is", err)
	}
}

func assertCode(t *testing.T, err error, want apperror.Code) {
	t.Helper()

	appErr, ok := errors.AsType[*apperror.AppError](err)
	if !ok || appErr.Code != want {
		t.Errorf("err = %v, want %s", err, want)
	}
}
