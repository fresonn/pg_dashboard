package hub

import (
	"context"
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/logger"
	metricsEntities "dashboard/api/internal/model/metrics"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const waitTimeout = 2 * time.Second

func testLogger() logger.Logger {
	return logger.WithScopeLogger(
		logger.New(config.AppConfig{Env: config.EnvironmentVars{Runtime: config.TestRuntime}}),
		"hub-test",
	)
}

func newHub(buffer int) *Hub {
	return New(Options{Buffer: buffer, Logger: testLogger()})
}

func update(name metricsEntities.Name, value int64, at time.Time) metricsEntities.Update {
	return metricsEntities.Update{
		Name:      name,
		Value:     metricsEntities.Int64(value),
		Timestamp: at,
	}
}

func receive(t *testing.T, sub *Subscription) metricsEntities.Update {
	t.Helper()

	select {
	case got, ok := <-sub.Updates():
		require.True(t, ok, "subscription channel closed unexpectedly")
		return got
	case <-time.After(waitTimeout):
		t.Fatal("timed out waiting for an update")
		return metricsEntities.Update{}
	}
}

func TestPublishFanOut(t *testing.T) {

	t.Parallel()

	h := newHub(4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	first := h.Subscribe(ctx)
	second := h.Subscribe(ctx)

	defer first.Close()
	defer second.Close()

	now := time.Now().UTC()

	h.Publish(update(metricsEntities.NameConnections, 26, now))

	for _, sub := range []*Subscription{first, second} {
		got := receive(t, sub)
		assert.Equal(t, metricsEntities.NameConnections, got.Name)
		assert.Equal(t, now, got.Timestamp)
	}

	assert.Equal(t, 2, h.Stats().Subscribers)
	assert.Equal(t, int64(1), h.Stats().Published)
}

func TestPublishRejectsInvalidUpdates(t *testing.T) {

	t.Parallel()

	h := newHub(4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := h.Subscribe(ctx)
	defer sub.Close()

	h.Publish(
		metricsEntities.Update{Name: metricsEntities.NameConnections},                  // no timestamp
		metricsEntities.Update{Value: metricsEntities.Int64(1), Timestamp: time.Now()}, // no name
		metricsEntities.Update{Name: "unknown", Value: metricsEntities.Int64(1), Timestamp: time.Now()},
	)

	stats := h.Stats()
	assert.Equal(t, int64(0), stats.Published)
	assert.Equal(t, int64(3), stats.Invalid)

	select {
	case got := <-sub.Updates():
		t.Fatalf("unexpected update %v", got)
	default:
	}
}

func TestSnapshotContainsLatestValues(t *testing.T) {

	t.Parallel()

	h := newHub(4)

	older := time.Now().UTC()
	newer := older.Add(time.Second)

	h.Publish(update(metricsEntities.NameConnections, 10, older))
	h.Publish(update(metricsEntities.NameConnections, 12, newer))
	h.Publish(update(metricsEntities.NameDatabaseSize, 4096, newer))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := h.Subscribe(ctx)
	defer sub.Close()

	snapshot := sub.Snapshot()
	require.Len(t, snapshot, 2)

	byName := make(map[metricsEntities.Name]metricsEntities.Update, len(snapshot))

	for _, item := range snapshot {
		byName[item.Name] = item
	}

	connections, ok := byName[metricsEntities.NameConnections]
	require.True(t, ok)
	assert.Equal(t, newer, connections.Timestamp)

	intValue, ok := connections.Value.Int()
	require.True(t, ok)
	assert.Equal(t, int64(12), intValue)
}

func TestUnsubscribeStopsDelivery(t *testing.T) {

	t.Parallel()

	h := newHub(4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := h.Subscribe(ctx)

	sub.Close()
	sub.Close() // idempotent

	assert.Equal(t, 0, h.Stats().Subscribers)

	_, ok := <-sub.Updates()
	assert.False(t, ok, "channel must be closed after unsubscribe")

	// Publishing to a released subscription must not panic.
	h.Publish(update(metricsEntities.NameConnections, 1, time.Now().UTC()))

	<-sub.Done()
	assert.Equal(t, int64(1), h.Stats().Published)
}

func TestContextCancelRemovesSubscriber(t *testing.T) {

	t.Parallel()

	h := newHub(4)

	ctx, cancel := context.WithCancel(context.Background())

	sub := h.Subscribe(ctx)

	require.Equal(t, 1, h.Stats().Subscribers)

	cancel()

	require.Eventually(t, func() bool {
		return h.Stats().Subscribers == 0
	}, waitTimeout, 5*time.Millisecond)

	_, ok := <-sub.Updates()
	assert.False(t, ok)
}

// TestSlowSubscriberDropsOldest pins the overflow policy: a full queue evicts
// the oldest sample, so the client always sees the freshest values, and every
// eviction is counted for logging/metrics.
func TestSlowSubscriberDropsOldest(t *testing.T) {

	t.Parallel()

	h := newHub(2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := h.Subscribe(ctx)
	defer sub.Close()

	now := time.Now().UTC()

	for i := range 5 {
		h.Publish(update(metricsEntities.NameConnections, int64(i), now.Add(time.Duration(i)*time.Second)))
	}

	assert.Equal(t, int64(3), sub.Dropped())
	assert.Equal(t, int64(3), h.Stats().Dropped)

	var values []int64

	for {
		select {
		case got := <-sub.Updates():
			intValue, ok := got.Value.Int()
			require.True(t, ok)
			values = append(values, intValue)
			continue
		default:
		}

		break
	}

	assert.Equal(t, []int64{3, 4}, values, "the newest samples must survive")

	assert.Equal(t, int64(3), sub.TakeDropped())
	assert.Equal(t, int64(0), sub.TakeDropped(), "the counter is reset after reading")
}

// TestSlowSubscriberDoesNotBlockOthers checks that a subscriber that never
// reads cannot stall publishing for the rest of the clients.
func TestSlowSubscriberDoesNotBlockOthers(t *testing.T) {

	t.Parallel()

	h := newHub(4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slow := h.Subscribe(ctx)
	defer slow.Close()

	fast := h.Subscribe(ctx)
	defer fast.Close()

	const publishes = 50

	var received atomic.Int64

	go func() {
		for range fast.Updates() {
			received.Add(1)
		}
	}()

	done := make(chan struct{})

	go func() {
		defer close(done)

		now := time.Now().UTC()

		for i := range publishes {
			h.Publish(update(metricsEntities.NameConnections, int64(i), now.Add(time.Duration(i)*time.Second)))
		}
	}()

	select {
	case <-done:
	case <-time.After(waitTimeout):
		t.Fatal("publish blocked on a slow subscriber")
	}

	require.Eventually(t, func() bool {
		// Every published update either reached the subscriber or was evicted
		// by the overflow policy, so the accounting must always add up.
		return received.Load()+fast.Dropped() == publishes
	}, waitTimeout, 5*time.Millisecond)

	assert.Positive(t, received.Load())

	slowBuffered := 0

	for {
		select {
		case <-slow.Updates():
			slowBuffered++
			continue
		default:
		}

		break
	}

	assert.Equal(t, 4, slowBuffered, "the slow subscriber keeps its bounded queue")
	assert.Greater(t, slow.Dropped(), int64(0))
}

// TestSubscriptionCleanup verifies both exit paths of the subscription
// watcher: an explicit Close and a cancelled context. Done() is closed in both
// cases, which is the only condition the watcher waits for, so no goroutine can
// stay blocked and no channel stays registered.
//
// A raw runtime.NumGoroutine comparison is deliberately not used: Go runs test
// packages in parallel, so the process wide count is not attributable to the
// hub.
func TestSubscriptionCleanup(t *testing.T) {

	t.Parallel()

	h := newHub(4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	explicitlyClosed := make([]*Subscription, 0, 10)

	for range 10 {
		explicitlyClosed = append(explicitlyClosed, h.Subscribe(ctx))
	}

	for _, sub := range explicitlyClosed {
		sub.Close()
	}

	for _, sub := range explicitlyClosed {
		assertReleased(t, sub)
	}

	assert.Equal(t, 0, h.Stats().Subscribers)

	cancel()

	cancelled := make([]*Subscription, 0, 10)

	cancelCtx, cancelSubs := context.WithCancel(context.Background())

	for range 10 {
		cancelled = append(cancelled, h.Subscribe(cancelCtx))
	}

	cancelSubs()

	for _, sub := range cancelled {
		assertReleased(t, sub)
	}

	assert.Equal(t, 0, h.Stats().Subscribers)

	// Repeated subscribe/release cycles must not accumulate registrations.
	for range 5 {
		cycle := make([]*Subscription, 0, 20)

		for range 20 {
			cycle = append(cycle, h.Subscribe(ctx))
		}

		for _, sub := range cycle {
			sub.Close()
		}
	}

	assert.Equal(t, 0, h.Stats().Subscribers)
}

func assertReleased(t *testing.T, sub *Subscription) {
	t.Helper()

	select {
	case <-sub.Done():
	case <-time.After(waitTimeout):
		t.Fatal("subscription was not released")
	}

	_, ok := <-sub.Updates()
	assert.False(t, ok, "channel must be closed after release")
}

func TestConcurrentPublishAndSubscribe(t *testing.T) {

	t.Parallel()

	h := newHub(8)

	ctx, cancel := context.WithCancel(context.Background())

	var (
		wg        sync.WaitGroup
		publishes = 200
	)

	for worker := range 4 {
		wg.Add(1)

		go func(worker int) {
			defer wg.Done()

			now := time.Now().UTC()

			for i := range publishes {
				h.Publish(update(metricsEntities.NameConnections, int64(worker*publishes+i), now))
			}
		}(worker)
	}

	for range 8 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			sub := h.Subscribe(ctx)
			time.Sleep(time.Millisecond)
			sub.Close()
		}()
	}

	// Closing subscriptions while publishers are active must never panic.
	wg.Wait()
	cancel()

	require.Eventually(t, func() bool {
		return h.Stats().Subscribers == 0
	}, waitTimeout, 5*time.Millisecond)

	assert.Equal(t, int64(4*publishes), h.Stats().Published)
}

func TestBufferDefault(t *testing.T) {

	t.Parallel()

	h := New(Options{Buffer: 0})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := h.Subscribe(ctx)
	defer sub.Close()

	assert.Equal(t, DefaultBuffer, h.buffer)
	assert.Equal(t, DefaultBuffer, cap(sub.ch))
}
