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

// MosqueProfileRepositoryItf stores the mosque's profile and logo, one of each
// since a deployment serves a single mosque. Updates insert or replace (upsert).
type MosqueProfileRepositoryItf interface {
	// Get returns domain.ErrMosqueProfileNotFound until the profile is first saved
	Get(ctx context.Context) (domain.MosqueProfile, error)
	Update(ctx context.Context, profile domain.MosqueProfile, updatedBy int64) error
	// GetLogo returns domain.ErrMosqueLogoNotFound when no logo is uploaded
	GetLogo(ctx context.Context) (domain.MosqueLogo, error)
	UpdateLogo(ctx context.Context, logo domain.MosqueLogo, updatedBy int64) error
	// DeleteLogo removes the logo; a missing logo is not an error
	DeleteLogo(ctx context.Context) error
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

type PrayerTimeRepositoryItf interface {
	GetSettings(ctx context.Context) (domain.PrayerSettings, error)
	UpdateSettings(ctx context.Context, settings domain.PrayerSettings, updatedBy int64) error
	CreateOverride(ctx context.Context, override domain.PrayerTimeOverride) (domain.PrayerTimeOverride, error)
	FindOverridesBetween(ctx context.Context, from, to string) ([]domain.PrayerTimeOverride, error)
	DeleteOverride(ctx context.Context, id int64) error
}

// IqamahRepositoryItf stores what the TV does around each prayer, as one JSON row in configurations
type IqamahRepositoryItf interface {
	// GetSettings returns domain.ErrIqamahSettingsNotFound when never saved
	GetSettings(ctx context.Context) (domain.IqamahSettings, error)
	// UpdateSettings inserts or replaces the settings; updatedBy 0 is nobody
	UpdateSettings(ctx context.Context, settings domain.IqamahSettings, updatedBy int64) error
}

// FridayDutyRepositoryItf stores the petugas Jumat, one row per Friday
type FridayDutyRepositoryItf interface {
	// Save inserts or replaces the petugas of duty.Date and returns the stored row
	Save(ctx context.Context, duty domain.FridayDuty) (domain.FridayDuty, error)
	// FindBetween returns the petugas Jumat dated from..to inclusive, YYYY-MM-DD
	FindBetween(ctx context.Context, from, to string) ([]domain.FridayDuty, error)
	// Delete returns domain.ErrFridayDutyNotFound when that Friday has no petugas
	Delete(ctx context.Context, date string) error
}
