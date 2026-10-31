package sessions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/cache"
)

func newTestRepository(t *testing.T) (*SessionCacheRepository, cache.Cache) {
	t.Helper()

	c := cache.NewMemory(time.Hour)
	t.Cleanup(func() { c.Close() })

	return NewSessionCacheRepository(c), c
}

func TestSessionLifecycle(t *testing.T) {
	repo, c := newTestRepository(t)
	ctx := context.Background()

	if err := repo.Create(ctx, 42, "abc", time.Hour); err != nil {
		t.Fatalf("create: %v", err)
	}

	if raw, err := c.Get(ctx, "session:user:42"); err != nil || raw != "abc" {
		t.Errorf("cache key session:user:42 = %q, %v; want abc", raw, err)
	}

	sessionID, err := repo.GetSessionID(ctx, 42)
	if err != nil || sessionID != "abc" {
		t.Errorf("get = %q, %v; want abc", sessionID, err)
	}

	if err := repo.Delete(ctx, 42); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetSessionID(ctx, 42); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Errorf("get after delete: err = %v, want ErrSessionNotFound", err)
	}

	if err := repo.Delete(ctx, 42); err != nil {
		t.Errorf("deleting a missing session: %v", err)
	}
}

func TestNewSessionReplacesPrevious(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	_ = repo.Create(ctx, 42, "phone", time.Hour)
	_ = repo.Create(ctx, 42, "laptop", time.Hour)
	_ = repo.Create(ctx, 43, "other-user", time.Hour)

	if sessionID, _ := repo.GetSessionID(ctx, 42); sessionID != "laptop" {
		t.Errorf("session of user 42 = %q, want the latest login (laptop)", sessionID)
	}
	if sessionID, _ := repo.GetSessionID(ctx, 43); sessionID != "other-user" {
		t.Errorf("session of user 43 = %q, want it untouched", sessionID)
	}
}

func TestSessionErrors(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	if _, err := repo.GetSessionID(ctx, 42); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Errorf("missing session: err = %v, want ErrSessionNotFound", err)
	}

	if err := repo.Create(ctx, 42, "abc", 0); err == nil {
		t.Error("session without ttl accepted")
	}
}
