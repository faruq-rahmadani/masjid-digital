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

type RunningTextRepositoryItf interface {
	// FindAll returns every running text in display order
	FindAll(ctx context.Context) ([]domain.RunningText, error)
	// Create inserts an active running text after the last one
	Create(ctx context.Context, text domain.RunningText) (domain.RunningText, error)
	// Update saves text and active, or returns domain.ErrRunningTextNotFound
	Update(ctx context.Context, text domain.RunningText) (domain.RunningText, error)
	// Delete returns domain.ErrRunningTextNotFound when no running text has the ID
	Delete(ctx context.Context, id int64) error
	// Reorder sets each ID's position to its index in ids
	Reorder(ctx context.Context, ids []int64) error
	// GetSetting returns domain.ErrRunningTextSettingNotFound until the setting is first saved
	GetSetting(ctx context.Context) (domain.RunningTextSetting, error)
	UpdateSetting(ctx context.Context, setting domain.RunningTextSetting, updatedBy int64) error
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

type SlideRepositoryItf interface {
	// FindAll returns every slide in display order, without images
	FindAll(ctx context.Context) ([]domain.Slide, error)
	// Create inserts an active slide with its image after the last one
	Create(ctx context.Context, slide domain.Slide, image domain.SlideImage) (domain.Slide, error)
	// Update saves title, caption, duration, show caption and active, or returns domain.ErrSlideNotFound
	Update(ctx context.Context, slide domain.Slide) (domain.Slide, error)
	// GetImage returns domain.ErrSlideNotFound when no slide has the ID
	GetImage(ctx context.Context, id int64) (domain.SlideImage, error)
	// Delete returns domain.ErrSlideNotFound when no slide has the ID
	Delete(ctx context.Context, id int64) error
	// Reorder sets each ID's position to its index in ids
	Reorder(ctx context.Context, ids []int64) error
}

// ScreenRepositoryItf stores the TVs, one row per device. Rows are written when a device first
// connects and when an admin changes one, never per heartbeat.
type ScreenRepositoryItf interface {
	// FindAll returns every screen, blocked ones too, oldest first
	FindAll(ctx context.Context) ([]domain.Screen, error)
	// FindByID returns domain.ErrScreenNotFound when no screen has the ID
	FindByID(ctx context.Context, id int64) (domain.Screen, error)
	// FindByDeviceID returns domain.ErrScreenNotFound for an unknown device
	FindByDeviceID(ctx context.Context, deviceID string) (domain.Screen, error)
	// Register inserts the device unless it already has a row, and returns its row either way
	Register(ctx context.Context, screen domain.Screen) (domain.Screen, error)
	// Update saves the name, forced mode and until, and blocked_at, or returns domain.ErrScreenNotFound
	Update(ctx context.Context, screen domain.Screen) (domain.Screen, error)
}

// ScreenPresenceRepositoryItf keeps when each device was last seen and on which app version,
// in the cache (Redis in production) rather than the database
type ScreenPresenceRepositoryItf interface {
	// Get returns domain.ErrScreenPresenceNotFound for a device not seen since the cache was emptied
	Get(ctx context.Context, deviceID string) (domain.ScreenPresence, error)
	Set(ctx context.Context, deviceID string, presence domain.ScreenPresence) error
}

// ScreenHubItf holds the live TV connections of this API instance
type ScreenHubItf interface {
	// Subscribe registers one connection of the device; call unsubscribe when it ends
	Subscribe(deviceID string) (activityLog <-chan domain.ScreenEvent, unsubscribe func())
	// Send delivers the event to every connection of the device; a missing device is ignored
	Send(deviceID string, event domain.ScreenEvent)
	// Broadcast delivers the event to every connection
	Broadcast(event domain.ScreenEvent)
	IsConnected(deviceID string) bool
}

type ActivityLogRepositoryItf interface {
	Create(ctx context.Context, log domain.ActivityLog) error
	// FindRecent returns the latest limit logs, newest first, with their actor's name
	FindRecent(ctx context.Context, limit int) ([]domain.ActivityLog, error)
}

// ActivityLogPublisherItf saves activity logs in the background. Publish never blocks and never
// fails: a full queue drops the log, so call it after the change is saved and ignore the outcome.
type ActivityLogPublisherItf interface {
	// Publish fills an empty ActorID from ctx (see domain.WithActor) and an empty CreatedAt with
	// now, then queues the log
	Publish(ctx context.Context, activityLog domain.ActivityLog)
}
