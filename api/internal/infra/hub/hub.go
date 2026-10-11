// Package hub implements the in-process fan-out of metric measurements.
//
// Collectors publish measurements to the hub, the SSE transport subscribes to
// them. The hub is deliberately independent from HTTP and from PostgreSQL: it
// only knows the metrics model, so it can be tested in isolation.
//
// The hub is a single-process component. If the application is scaled
// horizontally, every instance collects and streams its own metrics.
package hub

import (
	"context"
	"dashboard/api/internal/infra/logger"
	"dashboard/api/internal/model/metrics"
	"sync"
	"sync/atomic"
)

// DefaultBuffer is the per-subscriber queue size used when Options.Buffer is
// not set.
const DefaultBuffer = 64

// Options configures the hub.
type Options struct {
	// Buffer is the number of measurements kept for a single subscriber.
	Buffer int
	// Logger is optional, but strongly recommended: dropped updates for slow
	// subscribers are reported there.
	Logger logger.Logger
}

// Hub broadcasts measurements to all active subscribers.
type Hub struct {
	logger logger.Logger
	buffer int

	// mu guards subscribers, latest and nextID. Publish acquires the write
	// lock for the whole fan-out; every send is non-blocking, so the lock is
	// never held for a noticeable time and a slow client cannot stall others.
	mu          sync.Mutex
	subscribers map[int64]*Subscription
	latest      map[metrics.Name]metrics.Update
	nextID      int64

	published atomic.Int64
	dropped   atomic.Int64
	invalid   atomic.Int64
}

// New creates an empty hub.
func New(options Options) *Hub {
	buffer := options.Buffer
	if buffer < 1 {
		buffer = DefaultBuffer
	}

	return &Hub{
		logger:      options.Logger,
		buffer:      buffer,
		subscribers: make(map[int64]*Subscription),
		latest:      make(map[metrics.Name]metrics.Update),
	}
}

// Publish sends measurements to every subscriber.
//
// Overflow policy: a subscriber has a bounded queue; when it is full the oldest
// queued sample is evicted to keep the freshest data (live charts only care
// about the latest value) and the eviction is counted and logged. Missed points
// are expected to be recovered from the history storage, not from the stream.
//
// Invalid measurements are rejected and counted; a publish never blocks and
// never panics.
func (h *Hub) Publish(updates ...metrics.Update) {
	if len(updates) == 0 {
		return
	}

	valid := make([]metrics.Update, 0, len(updates))

	for _, update := range updates {
		if err := update.Validate(); err != nil {
			h.invalid.Add(1)
			h.warn("rejected metric update", "metric", string(update.Name), "error", err)
			continue
		}

		if !metrics.IsKnown(update.Name) {
			h.invalid.Add(1)
			h.warn("rejected unknown metric", "metric", string(update.Name))
			continue
		}

		valid = append(valid, update)
	}

	if len(valid) == 0 {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, update := range valid {
		h.latest[update.Name] = update
	}

	h.published.Add(int64(len(valid)))

	for _, sub := range h.subscribers {
		for _, update := range valid {
			sub.push(update)
		}
	}
}

// Subscribe registers a new subscriber and returns its subscription.
//
// The subscription carries a snapshot of the latest known value of every metric
// observed so far, captured atomically with the registration: everything
// published before the call is in the snapshot, everything published after it
// arrives on the channel, so a client can combine the snapshot with history
// without gaps.
//
// The subscription is closed when ctx is cancelled or when Close is called.
func (h *Hub) Subscribe(ctx context.Context) *Subscription {
	h.mu.Lock()

	snapshot := make([]metrics.Update, 0, len(h.latest))

	for _, name := range metrics.Names() {
		if update, ok := h.latest[name]; ok {
			snapshot = append(snapshot, update)
		}
	}

	h.nextID++

	sub := &Subscription{
		id:       h.nextID,
		hub:      h,
		ch:       make(chan metrics.Update, h.buffer),
		closed:   make(chan struct{}),
		snapshot: snapshot,
	}

	h.subscribers[sub.id] = sub

	h.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
			sub.Close()
		case <-sub.closed:
		}
	}()

	return sub
}

// Stats returns a point-in-time snapshot of hub counters.
func (h *Hub) Stats() Stats {
	h.mu.Lock()
	subscribers := len(h.subscribers)
	h.mu.Unlock()

	return Stats{
		Subscribers: subscribers,
		Published:   h.published.Load(),
		Dropped:     h.dropped.Load(),
		Invalid:     h.invalid.Load(),
	}
}

// unsubscribe removes the subscriber and closes its channels.
//
// The channels are closed while the write lock is held: Publish sends under the
// same lock, so a send on a closed channel (and a corresponding panic) is
// impossible.
func (h *Hub) unsubscribe(id int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sub, ok := h.subscribers[id]
	if !ok {
		return
	}

	delete(h.subscribers, id)

	close(sub.closed)
	close(sub.ch)
}

func (h *Hub) warn(msg string, args ...any) {
	if h.logger == nil {
		return
	}

	h.logger.Warn(msg, args...)
}

// Stats is a point-in-time view of hub activity.
type Stats struct {
	Subscribers int
	Published   int64
	Dropped     int64
	Invalid     int64
}

// Subscription is a single subscriber of the hub.
//
// The struct is safe for concurrent use: Close is idempotent and may be called
// from a different goroutine than the one reading Updates.
type Subscription struct {
	id  int64
	hub *Hub
	ch  chan metrics.Update

	closed   chan struct{}
	snapshot []metrics.Update

	// sendMu serializes publishers pushing into this subscription, which makes
	// the drop accounting exact when several collectors publish at once.
	sendMu  sync.Mutex
	once    sync.Once
	dropped atomic.Int64
}

// ID returns the subscription identifier, useful for logs.
func (s *Subscription) ID() int64 {
	return s.id
}

// Updates returns the receive-only channel with new measurements.
func (s *Subscription) Updates() <-chan metrics.Update {
	return s.ch
}

// Snapshot returns the latest value of every metric known at subscribe time.
func (s *Subscription) Snapshot() []metrics.Update {
	return s.snapshot
}

// Dropped returns the total number of updates evicted for this subscription
// because the client was too slow.
func (s *Subscription) Dropped() int64 {
	return s.dropped.Load()
}

// TakeDropped returns the number of evicted updates since the previous call and
// resets the counter.
func (s *Subscription) TakeDropped() int64 {
	return s.dropped.Swap(0)
}

// Done returns a channel closed when the subscription is released.
func (s *Subscription) Done() <-chan struct{} {
	return s.closed
}

// Close unregisters the subscription. It is safe to call it multiple times.
func (s *Subscription) Close() {
	s.once.Do(func() {
		s.hub.unsubscribe(s.id)
	})
}

// push enqueues a measurement applying the drop-oldest overflow policy.
//
// The caller holds the hub write lock, so the subscription cannot be closed
// while push runs.
func (s *Subscription) push(update metrics.Update) {
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	select {
	case s.ch <- update:
		return
	default:
	}

	// Queue is full: evict the oldest sample to keep the freshest one.
	select {
	case evicted := <-s.ch:
		s.dropped.Add(1)
		s.hub.dropped.Add(1)
		s.hub.warn("slow subscriber, dropped metric update",
			"subscription", s.id,
			"metric", string(evicted.Name),
		)
	default:
	}

	select {
	case s.ch <- update:
	default:
		s.dropped.Add(1)
		s.hub.dropped.Add(1)
		s.hub.warn("slow subscriber, metric update dropped",
			"subscription", s.id,
			"metric", string(update.Name),
		)
	}
}
