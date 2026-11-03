package activitylogbus

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
)

var _ usecase.ActivityLogPublisherItf = (*Bus)(nil)

const (
	// queueSize is how many logs may wait for the worker before new ones are dropped. They come
	// from admins saving forms, a handful a minute, so a full queue means a stuck handler.
	queueSize = 256
	// handlerTimeout bounds one handler's work on one log (e.g. a database insert)
	handlerTimeout = 5 * time.Second
)

// Handler receives each published log. It runs on the bus's worker, one log at a time, with its
// own context: the request that published the log may be long gone.
type Handler func(ctx context.Context, activityLog domain.ActivityLog)

// Bus is an in-memory queue in front of the activity log: Publish queues the log and returns at
// once, and one worker goroutine hands each log to every handler in order. Nothing is persisted,
// so logs still queued when the process dies are lost; the activity log is best effort.
// Running several API instances is fine: each one saves its own logs.
type Bus struct {
	queue    chan domain.ActivityLog
	handlers []Handler
	done     chan struct{} // closed when the worker has finished
	now      func() time.Time

	mu     sync.RWMutex // guards closed, so a Publish racing Close never sends on a closed channel
	closed bool
}

// New starts the worker with the handlers, fixed for the bus's life. Call Close on shutdown,
// before the database the handlers write to is closed.
func New(handlers ...Handler) *Bus {
	b := &Bus{
		queue:    make(chan domain.ActivityLog, queueSize),
		handlers: handlers,
		done:     make(chan struct{}),
		now:      time.Now,
	}
	go b.run()
	return b
}

func (b *Bus) Publish(ctx context.Context, activityLog domain.ActivityLog) {
	if activityLog.ActorID == 0 {
		activityLog.ActorID = domain.ActorFrom(ctx)
	}
	if activityLog.CreatedAt.IsZero() {
		activityLog.CreatedAt = b.now().UTC()
	}

	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		log.Printf("activity log bus closed, dropping %s", activityLog.Action)
		return
	}
	select {
	case b.queue <- activityLog:
	default:
		log.Printf("activity log bus full, dropping %s", activityLog.Action)
	}
}

// Close stops taking logs and waits until the worker has handed out the queued ones
func (b *Bus) Close() {
	b.mu.Lock()
	if !b.closed {
		b.closed = true
		close(b.queue)
	}
	b.mu.Unlock()
	<-b.done
}

func (b *Bus) run() {
	defer close(b.done)
	for activityLog := range b.queue {
		for _, h := range b.handlers {
			dispatch(h, activityLog)
		}
	}
}

// dispatch runs one handler; a panic is logged so the other handlers and later logs still run
func dispatch(h Handler, activityLog domain.ActivityLog) {
	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("activity log handler panicked on %s: %v", activityLog.Action, r)
		}
	}()
	h(ctx, activityLog)
}
