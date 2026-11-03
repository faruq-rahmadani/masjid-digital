package screenhub

import (
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

func receive(t *testing.T, ch <-chan domain.ScreenEvent) domain.ScreenEventType {
	t.Helper()
	select {
	case ev := <-ch:
		return ev.Type
	default:
		return ""
	}
}

func TestHub(t *testing.T) {
	h := New()
	tv1, close1 := h.Subscribe("device-0001")
	tv2, close2 := h.Subscribe("device-0002")
	defer close2()

	if !h.IsConnected("device-0001") || h.IsConnected("device-9999") {
		t.Fatal("connected flags wrong")
	}

	h.Send("device-0001", domain.ScreenEvent{Type: domain.ScreenEventReload})
	if got := receive(t, tv1); got != domain.ScreenEventReload {
		t.Errorf("tv1 got %q, want reload", got)
	}
	if got := receive(t, tv2); got != "" {
		t.Errorf("tv2 got %q, want nothing", got)
	}

	h.Broadcast(domain.ScreenEvent{Type: domain.ScreenEventRefresh})
	if receive(t, tv1) != domain.ScreenEventRefresh || receive(t, tv2) != domain.ScreenEventRefresh {
		t.Error("broadcast did not reach both")
	}

	close1()
	close1() // closing twice is harmless
	if h.IsConnected("device-0001") {
		t.Error("device-0001 still connected after close")
	}
}

func TestSlowConnectionDoesNotBlock(t *testing.T) {
	h := New()
	_, unsubscribe := h.Subscribe("device-0001")
	defer unsubscribe()

	// nobody reads: sends past the buffer are dropped instead of blocking
	for range bufferSize * 3 {
		h.Broadcast(domain.ScreenEvent{Type: domain.ScreenEventRefresh})
	}
}
