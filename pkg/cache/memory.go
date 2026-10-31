package cache

import (
	"context"
	"sync"
	"time"
)

type item struct {
	value     string
	expiresAt time.Time // zero means no expiry
}

func (it item) expired(now time.Time) bool {
	return !it.expiresAt.IsZero() && !now.Before(it.expiresAt)
}

// Memory keeps keys in process memory. Expired keys are never returned by Get,
// and a background janitor frees them so unread keys do not pile up.
type Memory struct {
	mu    sync.RWMutex
	items map[string]item
	now   func() time.Time

	stop      chan struct{}
	closeOnce sync.Once
}

var _ Cache = (*Memory)(nil)

// NewMemory returns an in-memory cache that frees expired keys every janitorInterval
func NewMemory(janitorInterval time.Duration) *Memory {
	m := &Memory{
		items: make(map[string]item),
		now:   time.Now,
		stop:  make(chan struct{}),
	}

	go m.janitor(janitorInterval)

	return m
}

func (m *Memory) Get(_ context.Context, key string) (string, error) {
	m.mu.RLock()
	it, ok := m.items[key]
	m.mu.RUnlock()

	if !ok || it.expired(m.now()) {
		return "", ErrNotFound
	}

	return it.value, nil
}

func (m *Memory) Set(_ context.Context, key, value string, ttl time.Duration) error {
	it := item{value: value}
	if ttl > 0 {
		it.expiresAt = m.now().Add(ttl)
	}

	m.mu.Lock()
	m.items[key] = it
	m.mu.Unlock()

	return nil
}

func (m *Memory) Delete(_ context.Context, keys ...string) error {
	m.mu.Lock()
	for _, key := range keys {
		delete(m.items, key)
	}
	m.mu.Unlock()

	return nil
}

// Close stops the janitor. It is safe to call more than once.
func (m *Memory) Close() error {
	m.closeOnce.Do(func() { close(m.stop) })
	return nil
}

func (m *Memory) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.deleteExpired()
		case <-m.stop:
			return
		}
	}
}

func (m *Memory) deleteExpired() {
	now := m.now()

	m.mu.Lock()
	defer m.mu.Unlock()

	for key, it := range m.items {
		if it.expired(now) {
			delete(m.items, key)
		}
	}
}
