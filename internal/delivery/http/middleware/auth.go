package middleware

//go:generate go tool mockgen -source=auth.go -destination=../../../../mock/middleware_auth.go -package=mock

import (
	"context"
	"errors"
	"strings"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

// sessionContextKey is where RequireAuth stores the *domain.AuthSession in the echo context
const sessionContextKey = "auth_session"

type AuthenticatorItf interface {
	// Authenticate returns the session behind an access token, or an UNAUTHENTICATED AppError
	Authenticate(ctx context.Context, token string) (*domain.AuthSession, error)
}

type Auth struct {
	auth AuthenticatorItf
}

func NewAuth(auth AuthenticatorItf) *Auth {
	return &Auth{auth: auth}
}

// RequireAuth rejects requests without a valid "Authorization: Bearer <token>"
// whose session is still logged in, and stores the session for the handlers.
// Add it per route: a Group with middleware would also catch unknown paths and
// answer them 401 instead of 404.
func (m *Auth) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString, ok := bearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated, errors.New("missing bearer token"), "silakan login terlebih dahulu")
		}

		session, err := m.auth.Authenticate(c.Request().Context(), tokenString)
		if err != nil {
			return err
		}

		c.Set(sessionContextKey, session)
		return next(c)
	}
}

// Session returns the session RequireAuth stored for this request
func Session(c echo.Context) (*domain.AuthSession, error) {
	session, ok := c.Get(sessionContextKey).(*domain.AuthSession)
	if !ok || session == nil {
		// a protected handler was registered without RequireAuth
		return nil, apperror.New(apperror.CodeInternal, errors.New("no auth session in context; route is missing RequireAuth"), "gagal memeriksa sesi")
	}
	return session, nil
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header
func bearerToken(header string) (string, bool) {
	scheme, tokenString, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}

	tokenString = strings.TrimSpace(tokenString)
	return tokenString, tokenString != ""
}
