package authentication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/faruq-rahmadani/masjid-digital/pkg/hasher"
)

// maxPasswordLength matches the limit on create, so a huge input is rejected before hashing
const maxPasswordLength = 128

// errInvalidCredentials is the internal cause of every failed login. The client
// gets one message for unknown users and wrong passwords alike, so the response
// does not tell which phone numbers or emails are registered.
var errInvalidCredentials = apperror.New(apperror.CodeUnauthenticated, errors.New("invalid credentials"), "nomor HP/email atau password salah")

type AuthenticationUseCase struct {
	repo     usecase.UserRepositoryItf
	hasher   hasher.PasswordHasher
	tokens   usecase.TokenManagerItf
	sessions usecase.SessionRepositoryItf
	now      func() time.Time
}

func New(repo usecase.UserRepositoryItf, hasher hasher.PasswordHasher, tokens usecase.TokenManagerItf, sessions usecase.SessionRepositoryItf) *AuthenticationUseCase {
	return &AuthenticationUseCase{repo: repo, hasher: hasher, tokens: tokens, sessions: sessions, now: time.Now}
}

// Login checks the phone number or email and password, then issues an access token
func (u *AuthenticationUseCase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	if !req.IsPasswordLengthValid() {
		return nil, errInvalidCredentials
	}

	user, err := u.findUser(ctx, req)
	if errors.Is(err, domain.ErrUserNotFound) {
		// not the repository's "user tidak ditemukan": that would reveal who is registered
		return nil, errInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	ok, err := u.hasher.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, fmt.Errorf("error verifying password of user %d: %w", user.ID, err), "gagal login")
	}
	if !ok {
		return nil, errInvalidCredentials
	}

	// checked after the password, so only the account owner learns it is disabled
	if !user.IsActive() {
		return nil, apperror.New(apperror.CodeForbidden, fmt.Errorf("user %d is %s", user.ID, user.Status), "akun dinonaktifkan")
	}

	issued, err := u.tokens.Issue(user.ID, string(user.Role))
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, err, "gagal login")
	}

	// the token is only accepted while its session exists, so it can be logged out early
	// replaces any earlier session, so logging in here logs the user out elsewhere
	if err := u.sessions.Create(ctx, user.ID, issued.ID, issued.ExpiresIn); err != nil {
		return nil, err
	}

	if err := u.repo.UpdateLastLoginAt(ctx, user.ID, u.now()); err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		Token:    issued.Value,
		ExpireIn: int64(issued.ExpiresIn.Seconds()),
	}, nil
}

// Authenticate returns the session behind an access token. The token must be
// validly signed and unexpired, and must be the user's current session: logging
// out or logging in on another device makes it stop working.
func (u *AuthenticationUseCase) Authenticate(ctx context.Context, tokenString string) (*domain.AuthSession, error) {
	claims, err := u.tokens.Parse(tokenString)
	if err != nil {
		return nil, invalidSession(err)
	}

	// a logged out or expired session already comes back as UNAUTHENTICATED
	currentSessionID, err := u.sessions.GetSessionID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	if currentSessionID != claims.ID {
		return nil, invalidSession(fmt.Errorf("user %d logged in elsewhere: token session %s, current session %s", claims.UserID, claims.ID, currentSessionID))
	}

	return &domain.AuthSession{
		ID:     claims.ID,
		UserID: claims.UserID,
		Role:   domain.UserRole(claims.Role),
	}, nil
}

// Logout ends the session, so its token stops working before it expires
func (u *AuthenticationUseCase) Logout(ctx context.Context, session *domain.AuthSession) error {
	return u.sessions.Delete(ctx, session.UserID)
}

// findUser looks the identifier up as an email when it contains @, otherwise as a
// phone number. The identifier is matched as is: phones in E.164, emails lowercase.
func (u *AuthenticationUseCase) findUser(ctx context.Context, req *domain.LoginRequest) (domain.User, error) {
	method := req.IdentifierMethod()
	switch method {
	case domain.IdentifierMethodEmail:
		return u.repo.GetByEmail(ctx, req.Identifier)

	case domain.IdentifierMethodPhoneNumber:
		return u.repo.GetByPhone(ctx, req.Identifier)

	default:
		// same path as an unknown user, so the client sees the shared invalid credentials error
		return domain.User{}, domain.ErrUserNotFound
	}
}

func invalidSession(err error) *apperror.AppError {
	return apperror.New(apperror.CodeUnauthenticated, err, "sesi tidak valid, silakan login ulang")
}
