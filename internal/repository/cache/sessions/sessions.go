package sessions

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/faruq-rahmadani/masjid-digital/pkg/cache"
)

var _ usecase.SessionRepositoryItf = (*SessionCacheRepository)(nil)

// SessionCacheRepository keeps one key per user: session:user:<id> -> session ID.
// A new login overwrites the key, which logs the previous device out. The key
// expires with the token, so the cache never holds dead sessions for long.
type SessionCacheRepository struct {
	cache cache.Cache
}

func NewSessionCacheRepository(c cache.Cache) *SessionCacheRepository {
	return &SessionCacheRepository{cache: c}
}

func sessionKey(userID int64) string {
	return "session:user:" + strconv.FormatInt(userID, 10)
}

func (r *SessionCacheRepository) Create(ctx context.Context, userID int64, sessionID string, ttl time.Duration) error {
	if ttl <= 0 {
		// a session without expiry would outlive its token forever
		return apperror.New(apperror.CodeInternal, fmt.Errorf("session ttl must be positive, got %s", ttl), "gagal mengakses sesi")
	}

	if err := r.cache.Set(ctx, sessionKey(userID), sessionID, ttl); err != nil {
		return apperror.New(apperror.CodeInternal, fmt.Errorf("error creating session: %w", err), "gagal mengakses sesi")
	}

	return nil
}

func (r *SessionCacheRepository) GetSessionID(ctx context.Context, userID int64) (string, error) {
	sessionID, err := r.cache.Get(ctx, sessionKey(userID))
	if errors.Is(err, cache.ErrNotFound) {
		// keeps domain.ErrSessionNotFound inside for errors.Is
		return "", apperror.New(apperror.CodeUnauthenticated, domain.ErrSessionNotFound, "sesi tidak valid, silakan login ulang")
	}
	if err != nil {
		return "", apperror.New(apperror.CodeInternal, fmt.Errorf("error getting session: %w", err), "gagal mengakses sesi")
	}

	return sessionID, nil
}

func (r *SessionCacheRepository) Delete(ctx context.Context, userID int64) error {
	if err := r.cache.Delete(ctx, sessionKey(userID)); err != nil {
		return apperror.New(apperror.CodeInternal, fmt.Errorf("error deleting session: %w", err), "gagal mengakses sesi")
	}

	return nil
}
