package http

import (
	"bytes"
	"context"
	"dashboard/api/gen/openapi"
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/hub"
	"dashboard/api/internal/infra/logger"
	metricsEntities "dashboard/api/internal/model/metrics"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimeout = 5 * time.Second

func testLogger() logger.Logger {
	return logger.WithScopeLogger(
		logger.New(config.AppConfig{Env: config.EnvironmentVars{Runtime: config.TestRuntime}}),
		"transport-test",
	)
}

var errWriteFailed = errors.New("write failed")

// streamRecorder is an http.ResponseWriter with Flush support that captures the
// stream incrementally and can fail on purpose.
type streamRecorder struct {
	mu sync.Mutex

	header  http.Header
	body    bytes.Buffer
	status  int
	writes  int
	flushes int

	// failAfter is the number of successful writes after which Write fails.
	failAfter int
}

func newStreamRecorder() *streamRecorder {
	return &streamRecorder{header: make(http.Header), failAfter: 0}
}

func (r *streamRecorder) Header() http.Header {
	return r.header
}

func (r *streamRecorder) WriteHeader(status int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status == 0 {
		r.status = status
	}
}

func (r *streamRecorder) Write(payload []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.writes++

	if r.failAfter > 0 && r.writes > r.failAfter {
		return 0, errWriteFailed
	}

	if r.status == 0 {
		r.status = http.StatusOK
	}

	return r.body.Write(payload)
}

func (r *streamRecorder) Flush() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.flushes++
}

func (r *streamRecorder) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.body.String()
}

func (r *streamRecorder) statusCode() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.status
}

func (r *streamRecorder) writeCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.writes
}

func (r *streamRecorder) waitFor(t *testing.T, substring string) {
	t.Helper()

	require.Eventually(t, func() bool {
		return strings.Contains(r.String(), substring)
	}, testTimeout, 5*time.Millisecond, "stream never contained %q", substring)
}

// noFlushWriter implements http.ResponseWriter without Flusher.
type noFlushWriter struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newNoFlushWriter() *noFlushWriter {
	return &noFlushWriter{header: make(http.Header)}
}

func (w *noFlushWriter) Header() http.Header { return w.header }

func (w *noFlushWriter) Write(payload []byte) (int, error) { return w.body.Write(payload) }

func (w *noFlushWriter) WriteHeader(status int) { w.status = status }

type streamFixture struct {
	hub      *hub.Hub
	handler  *MetricsStreamHandler
	recorder *streamRecorder

	cancel context.CancelFunc
	done   chan struct{}
}

func newStreamHandler(t *testing.T, configure func(*config.MetricsVars)) (*hub.Hub, *MetricsStreamHandler) {
	t.Helper()

	cfg := config.DefaultMetricsVars()
	cfg.BatchWindow = 40 * time.Millisecond
	cfg.HeartbeatInterval = time.Hour
	cfg.RetryInterval = 3 * time.Second

	if configure != nil {
		configure(&cfg)
	}

	metricsHub := hub.New(hub.Options{Buffer: 16, Logger: testLogger()})

	return metricsHub, NewMetricsStreamHandler(metricsHub, cfg, testLogger())
}

func startStream(t *testing.T, metricsHub *hub.Hub, handler *MetricsStreamHandler) *streamFixture {
	t.Helper()

	recorder := newStreamRecorder()

	ctx, cancel := context.WithCancel(context.Background())

	request := httptest.NewRequest(http.MethodGet, MetricsStreamPath, nil).WithContext(ctx)

	done := make(chan struct{})

	go func() {
		defer close(done)
		handler.ServeHTTP(recorder, request)
	}()

	fixture := &streamFixture{
		hub:      metricsHub,
		handler:  handler,
		recorder: recorder,
		cancel:   cancel,
		done:     done,
	}

	t.Cleanup(func() {
		cancel()

		select {
		case <-done:
		case <-time.After(testTimeout):
			t.Error("stream handler did not stop after the client disconnected")
		}
	})

	// Wait for the subscription: publishing earlier would only fill the
	// snapshot, not the live channel.
	expectedSubscribers := metricsHub.Stats().Subscribers + 1

	require.Eventually(t, func() bool {
		return metricsHub.Stats().Subscribers >= expectedSubscribers
	}, testTimeout, time.Millisecond)

	return fixture
}

func (f *streamFixture) stop(t *testing.T) {
	t.Helper()

	f.cancel()

	select {
	case <-f.done:
	case <-time.After(testTimeout):
		t.Fatal("stream handler did not stop")
	}
}

func integerUpdate(name metricsEntities.Name, value int64, at time.Time) metricsEntities.Update {
	return metricsEntities.Update{Name: name, Value: metricsEntities.Int64(value), Timestamp: at}
}

// metricsBatches extracts the JSON payloads of all "event: metrics" blocks.
func metricsBatches(t *testing.T, raw string) []metricsEntities.Batch {
	t.Helper()

	batches := make([]metricsEntities.Batch, 0)

	for _, block := range normalizeBlocks(raw) {
		lines := strings.Split(block, "\n")

		if len(lines) < 2 || lines[0] != "event: "+sseEventMetrics {
			continue
		}

		data := strings.TrimPrefix(lines[1], "data: ")

		var batch metricsEntities.Batch

		require.NoError(t, json.Unmarshal([]byte(data), &batch), "payload %q is not a metrics batch", data)

		batches = append(batches, batch)
	}

	return batches
}

func normalizeBlocks(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")

	blocks := make([]string, 0)

	for _, block := range strings.Split(raw, "\n\n") {
		if strings.TrimSpace(block) == "" {
			continue
		}

		blocks = append(blocks, block)
	}

	return blocks
}

func TestStreamHeaders(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	fixture := startStream(t, metricsHub, handler)

	fixture.recorder.waitFor(t, "retry: 3000")

	header := fixture.recorder.Header()

	assert.Contains(t, header.Get("Content-Type"), "text/event-stream")
	assert.Equal(t, "no-cache, no-store, must-revalidate", header.Get("Cache-Control"))
	assert.Equal(t, "keep-alive", header.Get("Connection"))
	assert.Equal(t, "no", header.Get("X-Accel-Buffering"))
	assert.Equal(t, http.StatusOK, fixture.recorder.statusCode())
}

func TestStreamEventFormat(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	fixture := startStream(t, metricsHub, handler)

	measuredAt := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)

	metricsHub.Publish(integerUpdate(metricsEntities.NameConnections, 26, measuredAt))

	fixture.recorder.waitFor(t, "event: "+sseEventMetrics)

	raw := fixture.recorder.String()

	require.Contains(t, raw, "event: metrics\ndata: {\"connections\":{\"value\":26,\"timestamp\":\"2026-10-11T18:00:00Z\"}}\n\n")

	batches := metricsBatches(t, raw)
	require.Len(t, batches, 1)

	sample, ok := batches[0][metricsEntities.NameConnections]
	require.True(t, ok)

	intValue, ok := sample.Value.Int()
	require.True(t, ok)
	assert.Equal(t, int64(26), intValue)
	assert.Equal(t, measuredAt, sample.Timestamp)
}

func TestStreamBatchesUpdatesInOneWindow(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	fixture := startStream(t, metricsHub, handler)

	now := time.Now().UTC().Truncate(time.Millisecond)

	metricsHub.Publish(
		integerUpdate(metricsEntities.NameConnections, 26, now),
		metricsEntities.Update{
			Name:      metricsEntities.NameBufferCacheHit,
			Value:     metricsEntities.Float64(98.4),
			Timestamp: now.Add(-time.Second),
		},
	)

	fixture.recorder.waitFor(t, "event: "+sseEventMetrics)

	require.Eventually(t, func() bool {
		return len(metricsBatches(t, fixture.recorder.String())) == 1
	}, testTimeout, 5*time.Millisecond)

	batches := metricsBatches(t, fixture.recorder.String())
	require.Len(t, batches, 1, "updates of one window must be a single event")

	assert.Len(t, batches[0], 2)

	// Every metric keeps its own measurement timestamp.
	assert.Equal(t, now, batches[0][metricsEntities.NameConnections].Timestamp)
	assert.Equal(t, now.Add(-time.Second), batches[0][metricsEntities.NameBufferCacheHit].Timestamp)
}

// TestStreamCoalescesRepeatedMetric documents the batching policy: several
// updates of one metric inside a window collapse into the newest sample.
func TestStreamCoalescesRepeatedMetric(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	fixture := startStream(t, metricsHub, handler)

	now := time.Now().UTC().Truncate(time.Millisecond)

	metricsHub.Publish(integerUpdate(metricsEntities.NameConnections, 10, now))
	metricsHub.Publish(integerUpdate(metricsEntities.NameConnections, 11, now.Add(time.Second)))

	fixture.recorder.waitFor(t, "event: "+sseEventMetrics)

	require.Eventually(t, func() bool {
		batches := metricsBatches(t, fixture.recorder.String())

		return len(batches) == 1 && len(batches[0]) == 1
	}, testTimeout, 5*time.Millisecond)

	batches := metricsBatches(t, fixture.recorder.String())
	require.Len(t, batches, 1)

	sample := batches[0][metricsEntities.NameConnections]
	intValue, ok := sample.Value.Int()
	require.True(t, ok)
	assert.Equal(t, int64(11), intValue, "the newest sample must win")
	assert.Equal(t, now.Add(time.Second), sample.Timestamp)
}

// TestStreamDoesNotRepeatUnchangedMetrics checks that a tick without new values
// does not resend the whole metric set.
func TestStreamDoesNotRepeatUnchangedMetrics(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	fixture := startStream(t, metricsHub, handler)

	metricsHub.Publish(integerUpdate(metricsEntities.NameConnections, 26, time.Now().UTC()))

	fixture.recorder.waitFor(t, "event: "+sseEventMetrics)

	time.Sleep(200 * time.Millisecond)

	assert.Len(t, metricsBatches(t, fixture.recorder.String()), 1)
}

func TestStreamHeartbeat(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, func(cfg *config.MetricsVars) {
		cfg.HeartbeatInterval = 20 * time.Millisecond
	})

	fixture := startStream(t, metricsHub, handler)

	fixture.recorder.waitFor(t, ": heartbeat")

	assert.Empty(t, metricsBatches(t, fixture.recorder.String()), "a heartbeat is a comment, not an event")
}

// TestStreamSnapshotOnSubscribe checks that a fresh client immediately receives
// the latest known value of every metric, which closes the gap between the
// history request and the live stream.
func TestStreamSnapshotOnSubscribe(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	metricsHub.Publish(integerUpdate(metricsEntities.NameConnections, 26, time.Now().UTC()))

	// The subscription happens inside the handler, so the publish above only
	// lands in the snapshot.
	fixture := startStream(t, metricsHub, handler)

	fixture.recorder.waitFor(t, "event: "+sseEventMetrics)

	batches := metricsBatches(t, fixture.recorder.String())
	require.Len(t, batches, 1)

	sample, ok := batches[0][metricsEntities.NameConnections]
	require.True(t, ok)

	intValue, ok := sample.Value.Int()
	require.True(t, ok)
	assert.Equal(t, int64(26), intValue)
}

func TestStreamStopsWhenClientDisconnects(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	fixture := startStream(t, metricsHub, handler)

	fixture.stop(t)

	require.Eventually(t, func() bool {
		return metricsHub.Stats().Subscribers == 0
	}, testTimeout, 5*time.Millisecond, "a disconnected client must be removed from the hub")
}

func TestStreamWriteErrorStopsHandler(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	recorder := newStreamRecorder()
	recorder.failAfter = 1 // the retry hint succeeds, the metrics event fails

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	request := httptest.NewRequest(http.MethodGet, MetricsStreamPath, nil).WithContext(ctx)

	done := make(chan struct{})

	go func() {
		defer close(done)
		handler.ServeHTTP(recorder, request)
	}()

	require.Eventually(t, func() bool {
		return metricsHub.Stats().Subscribers == 1
	}, testTimeout, time.Millisecond)

	metricsHub.Publish(integerUpdate(metricsEntities.NameConnections, 26, time.Now().UTC()))

	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("a write error must end the stream")
	}

	assert.Equal(t, 2, recorder.writeCount())
	assert.Empty(t, metricsBatches(t, recorder.String()))
}

// TestStreamSerializationErrorStopsHandler covers the failure of the payload
// encoding: a non finite value cannot be serialized, so the stream ends instead
// of blocking.
func TestStreamSerializationErrorStopsHandler(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	recorder := newStreamRecorder()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	request := httptest.NewRequest(http.MethodGet, MetricsStreamPath, nil).WithContext(ctx)

	done := make(chan struct{})

	go func() {
		defer close(done)
		handler.ServeHTTP(recorder, request)
	}()

	require.Eventually(t, func() bool {
		return metricsHub.Stats().Subscribers == 1
	}, testTimeout, time.Millisecond)

	metricsHub.Publish(metricsEntities.Update{
		Name:      metricsEntities.NameBufferCacheHit,
		Value:     metricsEntities.Float64(math.NaN()),
		Timestamp: time.Now().UTC(),
	})

	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("a serialization error must end the stream")
	}

	assert.Empty(t, metricsBatches(t, recorder.String()))
}

func TestStreamWithoutFlusher(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	writer := newNoFlushWriter()

	handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, MetricsStreamPath, nil))

	assert.Equal(t, http.StatusInternalServerError, writer.status)
	assert.Contains(t, writer.body.String(), "streaming is not supported")
	assert.Equal(t, 0, metricsHub.Stats().Subscribers, "a failed handshake must not subscribe")
}

func TestStreamMultipleClients(t *testing.T) {

	t.Parallel()

	metricsHub, handler := newStreamHandler(t, nil)

	first := startStream(t, metricsHub, handler)
	second := startStream(t, metricsHub, handler)

	require.Equal(t, 2, metricsHub.Stats().Subscribers)

	metricsHub.Publish(integerUpdate(metricsEntities.NameConnections, 26, time.Now().UTC()))

	for _, fixture := range []*streamFixture{first, second} {
		fixture.recorder.waitFor(t, "event: "+sseEventMetrics)

		batches := metricsBatches(t, fixture.recorder.String())
		require.Len(t, batches, 1)
		assert.Contains(t, batches[0], metricsEntities.NameConnections)
	}
}

// TestStreamOverRealHTTP exercises the handler through a real HTTP server, so
// flushing and connection handling are not simulated.
func TestStreamOverRealHTTP(t *testing.T) {

	metricsHub, handler := newStreamHandler(t, nil)

	router := chi.NewRouter()
	router.Get(MetricsStreamPath, handler.ServeHTTP)

	server := httptest.NewServer(router)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+MetricsStreamPath, nil)
	require.NoError(t, err)

	response, err := server.Client().Do(request)
	require.NoError(t, err)

	defer response.Body.Close()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Contains(t, response.Header.Get("Content-Type"), "text/event-stream")

	require.Eventually(t, func() bool {
		return metricsHub.Stats().Subscribers == 1
	}, testTimeout, time.Millisecond)

	metricsHub.Publish(integerUpdate(metricsEntities.NameConnections, 26, time.Now().UTC()))

	lines := make(chan string, 8)

	go func() {
		defer close(lines)

		buffer := make([]byte, 256)

		for {
			read, err := response.Body.Read(buffer)
			if read > 0 {
				lines <- string(buffer[:read])
			}

			if err != nil {
				return
			}
		}
	}()

	received := ""

	require.Eventually(t, func() bool {
		for {
			select {
			case chunk, ok := <-lines:
				if !ok {
					return strings.Contains(received, "event: metrics")
				}

				received += chunk
			default:
				return strings.Contains(received, "event: metrics")
			}
		}
	}, testTimeout, 5*time.Millisecond)

	assert.Contains(t, received, `"connections":{"value":26`)
}

type fakeMetricsService struct {
	mu sync.Mutex

	filters []metricsEntities.HistoryFilter
	history metricsEntities.History
	err     error
}

func (f *fakeMetricsService) History(ctx context.Context, filter metricsEntities.HistoryFilter) (metricsEntities.History, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.filters = append(f.filters, filter)

	if f.err != nil {
		return metricsEntities.History{}, f.err
	}

	return f.history, nil
}

func (f *fakeMetricsService) lastFilter() metricsEntities.HistoryFilter {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.filters[len(f.filters)-1]
}

func TestMetricsHistoryHandler(t *testing.T) {

	t.Parallel()

	stored := metricsEntities.History{
		Series: map[metricsEntities.Name][]metricsEntities.SeriesPoint{
			metricsEntities.NameConnections: {
				{Timestamp: time.Now().UTC(), Value: metricsEntities.Int64(26)},
			},
		},
	}

	from := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	to := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name      string
		params    openapi.MetricsHistoryParams
		service   *fakeMetricsService
		wantError bool
	}{
		{
			name: "default filter",
			service: &fakeMetricsService{
				history: stored,
			},
			params: openapi.MetricsHistoryParams{},
		},
		{
			name: "explicit filter",
			service: &fakeMetricsService{
				history: stored,
			},
			params: openapi.MetricsHistoryParams{
				Names: stringPointer("connections,buffer_cache_hit"),
				From:  &from,
				To:    &to,
				Limit: intPointer(10),
			},
		},
		{
			name: "unknown metric",
			service: &fakeMetricsService{
				history: stored,
			},
			params:    openapi.MetricsHistoryParams{Names: stringPointer("connections,nope")},
			wantError: true,
		},
		{
			name: "non positive limit",
			service: &fakeMetricsService{
				history: stored,
			},
			params:    openapi.MetricsHistoryParams{Limit: intPointer(0)},
			wantError: true,
		},
		{
			name: "storage error",
			service: &fakeMetricsService{
				err: errors.New("badger is broken"),
			},
			params:    openapi.MetricsHistoryParams{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := New(nil, nil, nil, nil, tt.service)

			response, err := handler.MetricsHistory(context.Background(), openapi.MetricsHistoryRequestObject{
				Params: tt.params,
			})
			require.NoError(t, err)

			if tt.wantError {
				_, ok := response.(openapi.MetricsHistory400JSONResponse)
				assert.True(t, ok, "expected a 400 response, got %T", response)
				return
			}

			ok, isOK := response.(openapi.MetricsHistory200JSONResponse)
			require.True(t, isOK, "expected a 200 response, got %T", response)
			assert.Equal(t, stored, metricsEntities.History(ok))

			// The wire format is what the frontend consumes.
			recorder := httptest.NewRecorder()
			require.NoError(t, response.VisitMetricsHistoryResponse(recorder))

			var decoded metricsEntities.History
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &decoded))
			assert.Equal(t, stored, decoded)
		})
	}
}

func TestMetricsHistoryHandlerFilter(t *testing.T) {

	t.Parallel()

	service := &fakeMetricsService{history: metricsEntities.History{Series: map[metricsEntities.Name][]metricsEntities.SeriesPoint{}}}
	handler := New(nil, nil, nil, nil, service)

	from := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	to := time.Now().UTC().Truncate(time.Second)

	_, err := handler.MetricsHistory(context.Background(), openapi.MetricsHistoryRequestObject{
		Params: openapi.MetricsHistoryParams{
			Names: stringPointer(" connections , buffer_cache_hit ,, "),
			From:  &from,
			To:    &to,
			Limit: intPointer(25),
		},
	})
	require.NoError(t, err)

	filter := service.lastFilter()

	assert.Equal(t, []metricsEntities.Name{metricsEntities.NameConnections, metricsEntities.NameBufferCacheHit}, filter.Names)
	assert.Equal(t, from, filter.From)
	assert.Equal(t, to, filter.To)
	assert.Equal(t, 25, filter.Limit)
}

// TestMetricsStreamPathDoesNotConflictWithOpenAPIRouter documents why the SSE
// endpoint is registered separately: the generated router does not know this
// path, and no generated operation is shadowed by it.
func TestMetricsStreamPathDoesNotConflictWithOpenAPIRouter(t *testing.T) {

	t.Parallel()

	assert.Equal(t, "/api/metrics/stream", MetricsStreamPath)

	router := chi.NewRouter()
	openapi.HandlerWithOptions(openapi.Unimplemented{}, openapi.ChiServerOptions{
		BaseURL:    "/api",
		BaseRouter: router,
	})

	streamRecorder := httptest.NewRecorder()
	router.ServeHTTP(streamRecorder, httptest.NewRequest(http.MethodGet, MetricsStreamPath, nil))
	assert.Equal(t, http.StatusNotFound, streamRecorder.Code, "the SSE path must not be served by the generated router")

	generatedRecorder := httptest.NewRecorder()
	router.ServeHTTP(generatedRecorder, httptest.NewRequest(http.MethodGet, "/api/cluster/status", nil))
	assert.NotEqual(t, http.StatusNotFound, generatedRecorder.Code, "generated routes must keep working")
}

func stringPointer(value string) *string { return &value }

func intPointer(value int) *int { return &value }
