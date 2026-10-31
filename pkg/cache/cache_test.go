package cache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// harness is a cache under test plus a way to move its clock forward
type harness struct {
	cache   Cache
	advance func(time.Duration)
}

// fakeClock is a settable clock for the memory driver
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}

func newMemoryHarness(t *testing.T) (harness, *Memory, *fakeClock) {
	t.Helper()

	clock := &fakeClock{now: time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)}
	m := NewMemory(time.Hour) // janitor driven by hand in tests
	m.now = clock.Now
	t.Cleanup(func() { m.Close() })

	return harness{cache: m, advance: clock.Advance}, m, clock
}

func newRedisHarness(t *testing.T) harness {
	t.Helper()

	server := miniredis.RunT(t)
	r, err := NewRedis(context.Background(), "redis://"+server.Addr())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close() })

	// miniredis only expires keys when told time has passed
	return harness{cache: r, advance: server.FastForward}
}

// TestCache runs the same contract against every driver
func TestCache(t *testing.T) {
	drivers := map[Driver]func(t *testing.T) harness{
		DriverMemory: func(t *testing.T) harness { h, _, _ := newMemoryHarness(t); return h },
		DriverRedis:  newRedisHarness,
	}

	for driver, newHarness := range drivers {
		t.Run(string(driver), func(t *testing.T) {
			t.Run("set and get", func(t *testing.T) {
				h := newHarness(t)
				ctx := context.Background()

				if err := h.cache.Set(ctx, "session:abc", "42", time.Hour); err != nil {
					t.Fatalf("set: %v", err)
				}
				got, err := h.cache.Get(ctx, "session:abc")
				if err != nil || got != "42" {
					t.Errorf("get = %q, %v; want 42", got, err)
				}

				if _, err := h.cache.Get(ctx, "session:missing"); !errors.Is(err, ErrNotFound) {
					t.Errorf("get missing: err = %v, want ErrNotFound", err)
				}
			})

			t.Run("expires after ttl", func(t *testing.T) {
				h := newHarness(t)
				ctx := context.Background()

				if err := h.cache.Set(ctx, "session:abc", "42", time.Minute); err != nil {
					t.Fatalf("set: %v", err)
				}

				h.advance(59 * time.Second)
				if _, err := h.cache.Get(ctx, "session:abc"); err != nil {
					t.Errorf("get before ttl: %v", err)
				}

				h.advance(time.Second)
				if _, err := h.cache.Get(ctx, "session:abc"); !errors.Is(err, ErrNotFound) {
					t.Errorf("get after ttl: err = %v, want ErrNotFound", err)
				}
			})

			t.Run("no ttl never expires", func(t *testing.T) {
				h := newHarness(t)
				ctx := context.Background()

				for _, ttl := range []time.Duration{0, -time.Second} {
					if err := h.cache.Set(ctx, "key", "value", ttl); err != nil {
						t.Fatalf("set with ttl %s: %v", ttl, err)
					}
					h.advance(365 * 24 * time.Hour)
					if got, err := h.cache.Get(ctx, "key"); err != nil || got != "value" {
						t.Errorf("ttl %s: get = %q, %v; want value", ttl, got, err)
					}
				}
			})

			t.Run("set replaces value and ttl", func(t *testing.T) {
				h := newHarness(t)
				ctx := context.Background()

				_ = h.cache.Set(ctx, "key", "old", time.Minute)
				_ = h.cache.Set(ctx, "key", "new", time.Hour)

				h.advance(2 * time.Minute)
				if got, err := h.cache.Get(ctx, "key"); err != nil || got != "new" {
					t.Errorf("get = %q, %v; want new with the longer ttl", got, err)
				}
			})

			t.Run("delete", func(t *testing.T) {
				h := newHarness(t)
				ctx := context.Background()

				_ = h.cache.Set(ctx, "a", "1", time.Hour)
				_ = h.cache.Set(ctx, "b", "2", time.Hour)
				_ = h.cache.Set(ctx, "c", "3", time.Hour)

				if err := h.cache.Delete(ctx, "a", "b", "missing"); err != nil {
					t.Fatalf("delete: %v", err)
				}
				if err := h.cache.Delete(ctx); err != nil {
					t.Fatalf("delete without keys: %v", err)
				}

				for key, want := range map[string]bool{"a": false, "b": false, "c": true} {
					_, err := h.cache.Get(ctx, key)
					if exists := err == nil; exists != want {
						t.Errorf("key %s exists = %v, want %v (err %v)", key, exists, want, err)
					}
				}
			})
		})
	}
}

func TestMemoryJanitorFreesExpiredKeys(t *testing.T) {
	_, m, clock := newMemoryHarness(t)
	ctx := context.Background()

	_ = m.Set(ctx, "short", "1", time.Minute)
	_ = m.Set(ctx, "long", "2", time.Hour)
	_ = m.Set(ctx, "forever", "3", 0)

	clock.Advance(2 * time.Minute)
	m.deleteExpired()

	m.mu.RLock()
	_, short := m.items["short"]
	_, long := m.items["long"]
	_, forever := m.items["forever"]
	m.mu.RUnlock()

	if short || !long || !forever {
		t.Errorf("after sweep: short=%v long=%v forever=%v, want only short freed", short, long, forever)
	}
}

func TestMemoryJanitorRunsAndStops(t *testing.T) {
	m := NewMemory(10 * time.Millisecond)
	ctx := context.Background()

	_ = m.Set(ctx, "key", "value", time.Millisecond)

	deadline := time.Now().Add(time.Second)
	for {
		m.mu.RLock()
		_, ok := m.items["key"]
		m.mu.RUnlock()
		if !ok {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("janitor did not free the expired key")
		}
		time.Sleep(5 * time.Millisecond)
	}

	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if err := m.Close(); err != nil {
		t.Errorf("second close: %v", err)
	}
}

func TestOpen(t *testing.T) {
	ctx := context.Background()

	c, err := Open(ctx, Config{Driver: DriverMemory})
	if err != nil {
		t.Fatalf("open memory: %v", err)
	}
	c.Close()

	server := miniredis.RunT(t)
	c, err = Open(ctx, Config{Driver: DriverRedis, DSN: "redis://" + server.Addr()})
	if err != nil {
		t.Fatalf("open redis: %v", err)
	}
	c.Close()

	if _, err := Open(ctx, Config{Driver: "memcached"}); err == nil {
		t.Error("unsupported driver accepted")
	}
	if _, err := Open(ctx, Config{Driver: DriverRedis, DSN: "not a url"}); err == nil {
		t.Error("invalid redis dsn accepted")
	}

	addr := server.Addr()
	server.Close()
	if _, err := Open(ctx, Config{Driver: DriverRedis, DSN: "redis://" + addr}); err == nil {
		t.Error("unreachable redis accepted")
	}
}
