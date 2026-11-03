package screenpresence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/cache"
)

func TestPresence(t *testing.T) {
	c := cache.NewMemory(time.Hour)
	t.Cleanup(func() { c.Close() })
	repo := New(c)
	ctx := context.Background()

	if _, err := repo.Get(ctx, "device-0001"); !errors.Is(err, domain.ErrScreenPresenceNotFound) {
		t.Fatalf("never seen: err = %v, want ErrScreenPresenceNotFound", err)
	}

	seen := domain.ScreenPresence{LastSeenAt: time.Date(2026, 11, 2, 9, 30, 0, 0, time.UTC), AppVersion: "1.4.0"}
	if err := repo.Set(ctx, "device-0001", seen); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := repo.Get(ctx, "device-0001")
	if err != nil || !got.LastSeenAt.Equal(seen.LastSeenAt) || got.AppVersion != "1.4.0" {
		t.Errorf("got %+v, %v; want %+v", got, err, seen)
	}
	if raw, err := c.Get(ctx, "screen:device-0001"); err != nil || raw == "" {
		t.Errorf("cache key screen:device-0001 = %q, %v", raw, err)
	}
}
