package screenpresence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/faruq-rahmadani/masjid-digital/pkg/cache"
)

var _ usecase.ScreenPresenceRepositoryItf = (*repo)(nil)

// presenceTTL keeps a device's last seen for a month after it was last written, long enough for
// "Aktif 3 minggu lalu" and short enough that the cache never fills with gone TVs
const presenceTTL = 30 * 24 * time.Hour

// repo keeps one key per device: screen:<device_id> -> {"last_seen_at", "app_version"} as JSON.
// With the Redis driver it survives an API restart; with the memory driver screens just show no
// last seen until they reconnect.
type repo struct {
	cache cache.Cache
}

func New(c cache.Cache) *repo {
	return &repo{cache: c}
}

func presenceKey(deviceID string) string {
	return "screen:" + deviceID
}

func (r *repo) Get(ctx context.Context, deviceID string) (domain.ScreenPresence, error) {
	raw, err := r.cache.Get(ctx, presenceKey(deviceID))
	if errors.Is(err, cache.ErrNotFound) {
		return domain.ScreenPresence{}, apperror.New(apperror.CodeNotFound, domain.ErrScreenPresenceNotFound, "layar belum pernah terlihat")
	}
	if err != nil {
		return domain.ScreenPresence{}, errInternal(fmt.Errorf("error getting screen presence: %w", err))
	}

	var presence domain.ScreenPresence
	if err := json.Unmarshal([]byte(raw), &presence); err != nil {
		return domain.ScreenPresence{}, errInternal(fmt.Errorf("error decoding screen presence: %w", err))
	}
	return presence, nil
}

func (r *repo) Set(ctx context.Context, deviceID string, presence domain.ScreenPresence) error {
	encoded, err := json.Marshal(presence)
	if err != nil {
		return errInternal(fmt.Errorf("error encoding screen presence: %w", err))
	}
	if err := r.cache.Set(ctx, presenceKey(deviceID), string(encoded), presenceTTL); err != nil {
		return errInternal(fmt.Errorf("error saving screen presence: %w", err))
	}
	return nil
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses status layar")
}
