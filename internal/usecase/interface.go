package usecase

//go:generate go tool mockgen -source=interface.go -destination=../../mock/usecase.go -package=mock

import (
	"context"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/token"
)

// Repositories return *apperror.AppError ready for the client. NOT_FOUND and
// ALREADY_EXISTS errors keep the domain sentinel inside, so usecases can
// errors.Is them (e.g. domain.ErrUserNotFound) when they need another answer.
type UserRepositoryItf interface {
	// Create inserts a user and returns its ID, or domain.ErrUserAlreadyExists
	Create(ctx context.Context, user domain.User) (int64, error)
	// GetByPhone and GetByEmail return domain.ErrUserNotFound when no user matches
	GetByPhone(ctx context.Context, phone string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateLastLoginAt(ctx context.Context, id int64, at time.Time) error
	FindAll(ctx context.Context) ([]domain.User, error)
	// FindByID returns domain.ErrUserNotFound when no user has the ID
	FindByID(ctx context.Context, id int64) (domain.User, error)
	// Update saves name, phone, email, role and status. It returns
	// domain.ErrUserNotFound or domain.ErrUserAlreadyExists.
	Update(ctx context.Context, user domain.User) error
	UpdatePasswordHash(ctx context.Context, id int64, passwordHash string) error
}

// TokenManagerItf signs and verifies access tokens for logged in users
type TokenManagerItf interface {
	Issue(userID int64, role string) (token.Issued, error)
	// Parse returns token.ErrInvalidToken for a malformed, forged or expired token
	Parse(tokenString string) (token.Claims, error)
}

// SessionRepositoryItf keeps one logged in session per user. Creating a session
// replaces the user's previous one, so a user is logged in on one device at a time.
type SessionRepositoryItf interface {
	Create(ctx context.Context, userID int64, sessionID string, ttl time.Duration) error
	// GetSessionID returns domain.ErrSessionNotFound once the user logged out or the session expired
	GetSessionID(ctx context.Context, userID int64) (string, error)
	Delete(ctx context.Context, userID int64) error
}
