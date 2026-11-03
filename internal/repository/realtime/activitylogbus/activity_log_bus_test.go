package activitylogbus

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

// recorder is a handler keeping what it got
type recorder struct {
	mu   sync.Mutex
	logs []domain.ActivityLog
}

func (r *recorder) handle(_ context.Context, activityLog domain.ActivityLog) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, activityLog)
}

func (r *recorder) got() []domain.ActivityLog {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]domain.ActivityLog(nil), r.logs...)
}

func TestPublishFillsActorAndTimeThenDeliversInOrder(t *testing.T) {
	var rec recorder
	b := New(rec.handle)
	now := time.Date(2026, 11, 3, 8, 0, 0, 0, time.FixedZone("WIB", 7*3600))
	b.now = func() time.Time { return now }

	ctx := domain.WithActor(context.Background(), 7)
	b.Publish(ctx, domain.ActivityLog{Action: domain.ActivitySlideCreated, Summary: "menambah slide"})
	b.Publish(ctx, domain.ActivityLog{Action: domain.ActivitySlideDeleted, ActorID: 9}) // an actor set by the caller wins
	b.Close()                                                                           // Close drains the queue

	got := rec.got()
	if len(got) != 2 {
		t.Fatalf("got %d logs, want 2", len(got))
	}
	if got[0].Action != domain.ActivitySlideCreated || got[0].ActorID != 7 || !got[0].CreatedAt.Equal(now) || got[0].CreatedAt.Location() != time.UTC {
		t.Errorf("first = %+v", got[0])
	}
	if got[1].Action != domain.ActivitySlideDeleted || got[1].ActorID != 9 {
		t.Errorf("second = %+v", got[1])
	}
}

func TestEveryHandlerRunsEvenAfterAPanic(t *testing.T) {
	var rec recorder
	b := New(func(context.Context, domain.ActivityLog) { panic("boom") }, rec.handle)

	b.Publish(context.Background(), domain.ActivityLog{Action: domain.ActivitySlideCreated})
	b.Publish(context.Background(), domain.ActivityLog{Action: domain.ActivitySlideUpdated})
	b.Close()

	if got := rec.got(); len(got) != 2 {
		t.Errorf("got %d logs, want 2", len(got))
	}
}

func TestPublishNeverBlocks(t *testing.T) {
	release := make(chan struct{})
	b := New(func(context.Context, domain.ActivityLog) { <-release }) // a stuck subscriber

	finished := make(chan struct{})
	go func() {
		for range queueSize * 2 {
			b.Publish(context.Background(), domain.ActivityLog{Action: domain.ActivitySlideCreated})
		}
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("Publish blocked on a full queue")
	}
	close(release)
	b.Close()
}

func TestPublishAfterCloseIsDropped(t *testing.T) {
	var rec recorder
	b := New(rec.handle)
	b.Close()
	b.Close() // twice is fine

	b.Publish(context.Background(), domain.ActivityLog{Action: domain.ActivitySlideCreated})

	if got := rec.got(); len(got) != 0 {
		t.Errorf("got %d logs after close, want 0", len(got))
	}
}
