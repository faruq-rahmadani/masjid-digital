package screenhub

import (
	"sync"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
)

var _ usecase.ScreenHubItf = (*hub)(nil)

// bufferSize is how many events a slow connection may fall behind before new ones are dropped
// for it. Events are small and rare (a refresh, a command), so a full buffer means a dead TCP
// connection that the next keep-alive will close.
const bufferSize = 16

// hub holds this API instance's live TV connections in memory. One instance serves the mosque;
// running several would need a shared hub (e.g. Redis pub/sub) behind the same interface.
type hub struct {
	mu   sync.RWMutex
	subs map[string]map[chan domain.ScreenEvent]struct{} // device ID -> its open connections
}

func New() *hub {
	return &hub{subs: map[string]map[chan domain.ScreenEvent]struct{}{}}
}

func (h *hub) Subscribe(deviceID string) (<-chan domain.ScreenEvent, func()) {
	ch := make(chan domain.ScreenEvent, bufferSize)

	h.mu.Lock()
	if h.subs[deviceID] == nil {
		h.subs[deviceID] = map[chan domain.ScreenEvent]struct{}{}
	}
	h.subs[deviceID][ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.subs[deviceID], ch)
			if len(h.subs[deviceID]) == 0 {
				delete(h.subs, deviceID)
			}
			h.mu.Unlock()
		})
	}
	return ch, unsubscribe
}

func (h *hub) Send(deviceID string, event domain.ScreenEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[deviceID] {
		deliver(ch, event)
	}
}

func (h *hub) Broadcast(event domain.ScreenEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conns := range h.subs {
		for ch := range conns {
			deliver(ch, event)
		}
	}
}

func (h *hub) IsConnected(deviceID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs[deviceID]) > 0
}

// deliver never blocks: a connection that stopped reading loses the event instead of stalling
// every other TV
func deliver(ch chan domain.ScreenEvent, event domain.ScreenEvent) {
	select {
	case ch <- event:
	default:
	}
}
