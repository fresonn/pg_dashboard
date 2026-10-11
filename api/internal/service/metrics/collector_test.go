package metrics

import (
	"context"
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/logger"
	"dashboard/api/internal/infra/postgres"
	metricsEntities "dashboard/api/internal/model/metrics"
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
The pipeline is tested without a real PostgreSQL: the connection manager, the
collectors repositories, the shared buffers service, the hub and the history
storage are replaced with recording fakes.
*/

func testLogger() logger.Logger {
	return logger.WithScopeLogger(
		logger.New(config.AppConfig{Env: config.EnvironmentVars{Runtime: config.TestRuntime}}),
		"metrics-test",
	)
}

type fakeManager struct {
	connected atomic.Bool
}

func (m *fakeManager) IsConnected() bool {
	return m.connected.Load()
}

func (m *fakeManager) Status() postgres.ConnectionStatus {
	if m.connected.Load() {
		return postgres.StatusConnected
	}

	return postgres.StatusDisconnected
}

type fakeSharedBuffers struct {
	calls       atomic.Int64
	inFlight    atomic.Int64
	maxInFlight atomic.Int64

	mu    sync.Mutex
	block time.Duration
	err   error
	rate  sharedBuffers.BuffersHitRate
}

func (f *fakeSharedBuffers) ClusterBuffersHitRate(ctx context.Context) (sharedBuffers.BuffersHitRate, error) {

	f.calls.Add(1)

	current := f.inFlight.Add(1)

	for {
		observed := f.maxInFlight.Load()
		if current <= observed || f.maxInFlight.CompareAndSwap(observed, current) {
			break
		}
	}

	defer f.inFlight.Add(-1)

	f.mu.Lock()
	block := f.block
	err := f.err
	rate := f.rate
	f.mu.Unlock()

	if block > 0 {
		select {
		case <-ctx.Done():
			return sharedBuffers.BuffersHitRate{}, ctx.Err()
		case <-time.After(block):
		}
	}

	if err != nil {
		return sharedBuffers.BuffersHitRate{}, err
	}

	if rate.Timestamp.IsZero() {
		rate.Timestamp = time.Now().UTC()
	}

	return rate, nil
}

func (f *fakeSharedBuffers) configure(rate sharedBuffers.BuffersHitRate) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.rate = rate
}

func (f *fakeSharedBuffers) breakWith(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.err = err
}

func (f *fakeSharedBuffers) blockFor(duration time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.block = duration
}

type fakePostgresRepo struct {
	connectionsCalls atomic.Int64
	sizeCalls        atomic.Int64

	mu             sync.Mutex
	connectionsErr error
	sizeErr        error
	connections    metricsEntities.ConnectionsStats
	size           metricsEntities.DatabaseSizeStats
}

func (f *fakePostgresRepo) ClusterConnections(ctx context.Context) (metricsEntities.ConnectionsStats, error) {

	f.connectionsCalls.Add(1)

	f.mu.Lock()
	err := f.connectionsErr
	stats := f.connections
	f.mu.Unlock()

	if err != nil {
		return metricsEntities.ConnectionsStats{}, err
	}

	if stats.Timestamp.IsZero() {
		stats.Timestamp = time.Now().UTC()
	}

	return stats, nil
}

func (f *fakePostgresRepo) CurrentDatabaseSize(ctx context.Context) (metricsEntities.DatabaseSizeStats, error) {

	f.sizeCalls.Add(1)

	f.mu.Lock()
	err := f.sizeErr
	stats := f.size
	f.mu.Unlock()

	if err != nil {
		return metricsEntities.DatabaseSizeStats{}, err
	}

	if stats.Timestamp.IsZero() {
		stats.Timestamp = time.Now().UTC()
	}

	return stats, nil
}

type recordingPublisher struct {
	mu      sync.Mutex
	updates []metricsEntities.Update
}

func (p *recordingPublisher) Publish(updates ...metricsEntities.Update) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.updates = append(p.updates, updates...)
}

func (p *recordingPublisher) snapshot() []metricsEntities.Update {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append([]metricsEntities.Update(nil), p.updates...)
}

func (p *recordingPublisher) countByName(name metricsEntities.Name) int {
	count := 0

	for _, update := range p.snapshot() {
		if update.Name == name {
			count++
		}
	}

	return count
}

func (p *recordingPublisher) firstByName(t *testing.T, name metricsEntities.Name) metricsEntities.Update {
	t.Helper()

	for _, update := range p.snapshot() {
		if update.Name == name {
			return update
		}
	}

	t.Fatalf("metric %q was not published", name)

	return metricsEntities.Update{}
}

type fakeStorage struct {
	mu sync.Mutex

	appended      []metricsEntities.Update
	filters       []metricsEntities.HistoryFilter
	deletedBefore []time.Time
	history       metricsEntities.History

	appendErr error
	rangeErr  error
	deleteErr error
}

func (s *fakeStorage) Append(updates []metricsEntities.Update) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.appended = append(s.appended, updates...)

	return s.appendErr
}

func (s *fakeStorage) Range(filter metricsEntities.HistoryFilter) (metricsEntities.History, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.filters = append(s.filters, filter)

	if s.rangeErr != nil {
		return metricsEntities.History{}, s.rangeErr
	}

	return s.history, nil
}

func (s *fakeStorage) DeleteBefore(before time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deletedBefore = append(s.deletedBefore, before)

	if s.deleteErr != nil {
		return 0, s.deleteErr
	}

	return 0, nil
}

func (s *fakeStorage) appendedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.appended)
}

func (s *fakeStorage) appendedByName(name metricsEntities.Name) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0

	for _, update := range s.appended {
		if update.Name == name {
			count++
		}
	}

	return count
}

func (s *fakeStorage) cleanupRuns() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.deletedBefore)
}

type testEnv struct {
	service   *Service
	manager   *fakeManager
	pg        *fakePostgresRepo
	buffers   *fakeSharedBuffers
	publisher *recordingPublisher
	storage   *fakeStorage
}

func newTestEnv(cfg config.MetricsVars) *testEnv {

	env := &testEnv{
		manager:   &fakeManager{},
		pg:        &fakePostgresRepo{},
		buffers:   &fakeSharedBuffers{},
		publisher: &recordingPublisher{},
		storage:   &fakeStorage{},
	}

	env.manager.connected.Store(true)

	env.pg.connections = metricsEntities.ConnectionsStats{Connections: 26, MaxConnections: 100}
	env.pg.size = metricsEntities.DatabaseSizeStats{Database: "postgres", SizeBytes: 4096}
	env.buffers.rate = sharedBuffers.BuffersHitRate{HitRatio: 98.4, HitsTotal: 1000, ReadsTotal: 10}

	env.service = New(Options{
		Config: config.AppConfig{
			Env: config.EnvironmentVars{
				Runtime: config.TestRuntime,
				Metrics: cfg,
			},
		},
		Logger:        testLogger(),
		PgManager:     env.manager,
		PostgresRepo:  env.pg,
		SharedBuffers: env.buffers,
		Hub:           env.publisher,
		Storage:       env.storage,
	})

	return env
}

// runService runs the pipeline for the given duration and asserts that it stops
// as soon as the context is cancelled.
func runService(t *testing.T, service *Service, duration time.Duration) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		_ = service.Run(ctx)
	}()

	time.Sleep(duration)
	cancel()

	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("metrics service did not stop after context cancellation")
	}
}

func fastConfig() config.MetricsVars {
	cfg := config.DefaultMetricsVars()
	cfg.QueryTimeout = time.Second
	cfg.HistoryCleanupInterval = time.Hour

	return cfg
}

func TestCollectorsUseOwnIntervals(t *testing.T) {

	t.Parallel()

	cfg := fastConfig()
	cfg.BufferCacheHitInterval = 10 * time.Millisecond
	cfg.ConnectionsInterval = 25 * time.Millisecond
	cfg.DatabaseSizeInterval = 60 * time.Millisecond

	env := newTestEnv(cfg)

	runService(t, env.service, 220*time.Millisecond)

	bufferCalls := env.buffers.calls.Load()
	connectionsCalls := env.pg.connectionsCalls.Load()
	sizeCalls := env.pg.sizeCalls.Load()

	assert.GreaterOrEqual(t, bufferCalls, int64(5), "the fastest collector must run often")
	assert.Greater(t, bufferCalls, connectionsCalls)
	assert.Greater(t, connectionsCalls, sizeCalls)
	assert.GreaterOrEqual(t, sizeCalls, int64(1))

	assert.Equal(t, int64(1), env.buffers.maxInFlight.Load(), "runs must never overlap")
}

func TestCollectorFailureDoesNotStopOthers(t *testing.T) {

	t.Parallel()

	cfg := fastConfig()
	cfg.BufferCacheHitInterval = 10 * time.Millisecond
	cfg.ConnectionsInterval = 15 * time.Millisecond
	cfg.DatabaseSizeInterval = time.Hour

	env := newTestEnv(cfg)

	env.pg.mu.Lock()
	env.pg.connectionsErr = errors.New("permission denied for pg_stat_activity")
	env.pg.mu.Unlock()

	runService(t, env.service, 150*time.Millisecond)

	assert.Greater(t, env.pg.connectionsCalls.Load(), int64(1), "the failing collector keeps retrying")
	assert.Greater(t, env.buffers.calls.Load(), int64(3), "the healthy collector keeps running")
	assert.Positive(t, env.publisher.countByName(metricsEntities.NameBufferCacheHit))
	assert.Zero(t, env.publisher.countByName(metricsEntities.NameConnections))
}

func TestCollectorQueryTimeout(t *testing.T) {

	t.Parallel()

	cfg := fastConfig()
	cfg.QueryTimeout = 20 * time.Millisecond
	cfg.BufferCacheHitInterval = 10 * time.Millisecond
	cfg.ConnectionsInterval = time.Hour
	cfg.DatabaseSizeInterval = time.Hour

	env := newTestEnv(cfg)
	env.buffers.blockFor(5 * time.Second)

	runService(t, env.service, 120*time.Millisecond)

	assert.Zero(t, env.publisher.countByName(metricsEntities.NameBufferCacheHit), "a timed out collection publishes nothing")
	assert.Zero(t, env.storage.appendedByName(metricsEntities.NameBufferCacheHit), "a timed out collection stores nothing")
	assert.Greater(t, env.buffers.calls.Load(), int64(1), "the collector continues after a timeout")
}

// TestSlowCollectorDropsTicksInsteadOfOverlapping checks that a collection
// longer than the interval cannot start a second concurrent run.
func TestSlowCollectorDropsTicksInsteadOfOverlapping(t *testing.T) {

	t.Parallel()

	cfg := fastConfig()
	cfg.BufferCacheHitInterval = 5 * time.Millisecond
	cfg.ConnectionsInterval = time.Hour
	cfg.DatabaseSizeInterval = time.Hour

	env := newTestEnv(cfg)
	env.buffers.blockFor(30 * time.Millisecond)

	runService(t, env.service, 150*time.Millisecond)

	assert.Equal(t, int64(1), env.buffers.maxInFlight.Load())
	assert.Less(t, env.buffers.calls.Load(), int64(12), "missed ticks must be dropped, not queued")
}

func TestCollectorsSkipWhileDisconnected(t *testing.T) {

	t.Parallel()

	cfg := fastConfig()
	cfg.BufferCacheHitInterval = 5 * time.Millisecond
	cfg.ConnectionsInterval = 5 * time.Millisecond
	cfg.DatabaseSizeInterval = 5 * time.Millisecond

	env := newTestEnv(cfg)
	env.manager.connected.Store(false)

	runService(t, env.service, 80*time.Millisecond)

	assert.Zero(t, env.buffers.calls.Load())
	assert.Zero(t, env.pg.connectionsCalls.Load())
	assert.Zero(t, env.pg.sizeCalls.Load())
	assert.Empty(t, env.publisher.snapshot())
}

func TestCollectionIsPublishedAndStored(t *testing.T) {

	t.Parallel()

	cfg := fastConfig()
	cfg.BufferCacheHitInterval = 10 * time.Millisecond
	cfg.ConnectionsInterval = 10 * time.Millisecond
	cfg.DatabaseSizeInterval = 10 * time.Millisecond

	env := newTestEnv(cfg)

	runService(t, env.service, 60*time.Millisecond)

	rate := env.publisher.firstByName(t, metricsEntities.NameBufferCacheHit)
	assert.InDelta(t, 98.4, rate.Value.Float(), 0.001)
	assert.False(t, rate.Timestamp.IsZero())

	connections := env.publisher.firstByName(t, metricsEntities.NameConnections)
	intValue, ok := connections.Value.Int()
	require.True(t, ok)
	assert.Equal(t, int64(26), intValue)

	maxConnections := env.publisher.firstByName(t, metricsEntities.NameMaxConnections)
	intValue, ok = maxConnections.Value.Int()
	require.True(t, ok)
	assert.Equal(t, int64(100), intValue)

	size := env.publisher.firstByName(t, metricsEntities.NameDatabaseSize)
	intValue, ok = size.Value.Int()
	require.True(t, ok)
	assert.Equal(t, int64(4096), intValue)

	// History is written by the server, independently of SSE subscribers.
	assert.Greater(t, env.storage.appendedCount(), 0)
}

func TestHistoryFailureDoesNotStopPublishing(t *testing.T) {

	t.Parallel()

	cfg := fastConfig()
	cfg.BufferCacheHitInterval = 10 * time.Millisecond
	cfg.ConnectionsInterval = time.Hour
	cfg.DatabaseSizeInterval = time.Hour

	env := newTestEnv(cfg)
	env.storage.appendErr = errors.New("badger is broken")

	runService(t, env.service, 60*time.Millisecond)

	assert.Greater(t, env.storage.appendedByName(metricsEntities.NameBufferCacheHit), 0, "history was attempted")
	assert.Greater(t, env.publisher.countByName(metricsEntities.NameBufferCacheHit), 1, "live updates keep flowing")
}

func TestRetentionCleanupRunsWithConfiguredWindow(t *testing.T) {

	t.Parallel()

	cfg := fastConfig()
	cfg.BufferCacheHitInterval = time.Hour
	cfg.ConnectionsInterval = time.Hour
	cfg.DatabaseSizeInterval = time.Hour
	cfg.HistoryRetention = 30 * time.Minute
	cfg.HistoryCleanupInterval = 20 * time.Millisecond

	env := newTestEnv(cfg)

	runService(t, env.service, 120*time.Millisecond)

	require.GreaterOrEqual(t, env.storage.cleanupRuns(), 1)

	env.storage.mu.Lock()
	before := env.storage.deletedBefore[0]
	env.storage.mu.Unlock()

	assert.WithinDuration(t, time.Now().UTC().Add(-30*time.Minute), before, 10*time.Second)
}

func TestHistoryAppliesConfiguredDefaults(t *testing.T) {

	t.Parallel()

	cfg := fastConfig()
	cfg.HistoryRetention = 30 * time.Minute
	cfg.HistoryLimit = 123

	env := newTestEnv(cfg)

	stored := metricsEntities.History{
		Series: map[metricsEntities.Name][]metricsEntities.SeriesPoint{
			metricsEntities.NameConnections: {
				{Timestamp: time.Now().UTC(), Value: metricsEntities.Int64(26)},
			},
		},
	}

	env.storage.history = stored

	history, err := env.service.History(context.Background(), metricsEntities.HistoryFilter{})
	require.NoError(t, err)
	assert.Equal(t, stored, history)

	env.storage.mu.Lock()
	require.Len(t, env.storage.filters, 1)
	filter := env.storage.filters[0]
	env.storage.mu.Unlock()

	assert.Equal(t, 123, filter.Limit)
	assert.Equal(t, metricsEntities.Names(), filter.Names)
	assert.WithinDuration(t, filter.To.Add(-30*time.Minute), filter.From, time.Second)
}

func TestHistoryPropagatesStorageError(t *testing.T) {

	t.Parallel()

	env := newTestEnv(fastConfig())
	env.storage.rangeErr = errors.New("range failed")

	_, err := env.service.History(context.Background(), metricsEntities.HistoryFilter{})
	require.Error(t, err)
}

func TestCollectorsRegistry(t *testing.T) {

	t.Parallel()

	env := newTestEnv(fastConfig())

	registered := make([]string, 0, 3)

	for _, item := range env.service.collectors() {
		registered = append(registered, item.name)
	}

	assert.ElementsMatch(t, []string{
		string(metricsEntities.NameBufferCacheHit),
		string(metricsEntities.NameConnections),
		string(metricsEntities.NameDatabaseSize),
	}, registered)
}
