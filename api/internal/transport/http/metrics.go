package http

import (
	"context"
	"dashboard/api/gen/openapi"
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/hub"
	"dashboard/api/internal/infra/logger"
	metricsEntities "dashboard/api/internal/model/metrics"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	sseContentType  = "text/event-stream; charset=utf-8"
	sseEventMetrics = "metrics"
)

// MetricsStreamPath is the SSE endpoint. It is registered next to the generated
// OpenAPI router: a strict OpenAPI handler must return a complete body, while
// the metrics stream is endless.
const MetricsStreamPath = "/api/metrics/stream"

// MetricsHub is the subscription side of the metrics hub.
type MetricsHub interface {
	Subscribe(ctx context.Context) *hub.Subscription
}

// MetricsStreamHandler serves GET /api/metrics/stream.
//
// The handler is registered next to the OpenAPI router instead of going through
// the generated strict server: a strict handler must return a complete response
// body, which is impossible for an endless stream.
//
// Responsibilities:
//   - subscribe the client to the hub (the hub is shared by all clients);
//   - coalesce updates that arrive inside the batching window and send only the
//     metrics that actually changed, each with its own measurement timestamp;
//   - keep the connection alive with periodic comments;
//   - end the stream as soon as the request context is cancelled.
type MetricsStreamHandler struct {
	hub    MetricsHub
	logger logger.Logger
	cfg    config.MetricsVars
}

// NewMetricsStreamHandler creates the SSE handler.
func NewMetricsStreamHandler(subscriber MetricsHub, cfg config.MetricsVars, logger logger.Logger) *MetricsStreamHandler {
	return &MetricsStreamHandler{
		hub:    subscriber,
		logger: logger,
		cfg:    cfg.Sanitized(),
	}
}

func (h *MetricsStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	flusher, ok := w.(http.Flusher)
	if !ok {
		h.logger.ErrorContext(r.Context(), "sse: response writer does not support flushing")
		http.Error(w, "streaming is not supported", http.StatusInternalServerError)
		return
	}

	writeSSEHeaders(w)
	w.WriteHeader(http.StatusOK)

	stream := &sseStream{writer: w, flusher: flusher}

	if err := stream.writeRetry(h.cfg.RetryInterval); err != nil {
		h.logger.ErrorContext(r.Context(), "sse: write retry hint failed", "error", err)
		return
	}

	sub := h.hub.Subscribe(r.Context())
	defer sub.Close()

	h.logger.DebugContext(r.Context(), "sse: client subscribed", "subscription", sub.ID())

	// The snapshot closes the gap between the history request the client makes
	// on connect and the live stream: it contains the latest value of every
	// metric known at subscribe time, captured atomically with the
	// subscription. Anything newer arrives on the channel below.
	if err := h.writeSnapshot(stream, sub); err != nil {
		h.logger.ErrorContext(r.Context(), "sse: write snapshot failed", "error", err)
		return
	}

	h.stream(r.Context(), stream, sub)

	h.logger.DebugContext(r.Context(), "sse: client unsubscribed", "subscription", sub.ID())
}

// stream forwards updates to the client until the context is cancelled, the
// client disconnects or a write fails.
func (h *MetricsStreamHandler) stream(ctx context.Context, stream *sseStream, sub *hub.Subscription) {

	var (
		pending metricsEntities.Batch
		timer   *time.Timer
		windowC <-chan time.Time
	)

	heartbeat := time.NewTicker(h.cfg.HeartbeatInterval)
	defer heartbeat.Stop()

	flush := func() bool {
		if len(pending) == 0 {
			return true
		}

		if err := stream.writeMetrics(pending); err != nil {
			h.logger.ErrorContext(ctx, "sse: write metrics event failed", "error", err)
			return false
		}

		pending = nil

		h.reportDropped(ctx, sub)

		return true
	}

	stopWindow := func() {
		if timer != nil {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			timer = nil
		}

		windowC = nil
	}

	defer stopWindow()

	for {
		select {
		case <-ctx.Done():
			return

		case update, ok := <-sub.Updates():
			if !ok {
				return
			}

			if pending == nil {
				pending = make(metricsEntities.Batch, h.cfg.BatchSize)
				// The window starts with the first update: waiting metrics do
				// not delay the ones that are already collected.
				timer = time.NewTimer(h.cfg.BatchWindow)
				windowC = timer.C
			}

			pending.Add(update)

			if len(pending) >= h.cfg.BatchSize {
				if !flush() {
					return
				}

				stopWindow()
			}

		case <-windowC:
			timer = nil
			windowC = nil

			if !flush() {
				return
			}

		case <-heartbeat.C:
			if err := stream.writeComment("heartbeat"); err != nil {
				h.logger.ErrorContext(ctx, "sse: heartbeat failed", "error", err)
				return
			}

			h.reportDropped(ctx, sub)
		}
	}
}

// writeSnapshot sends the latest known values as a regular metrics event.
func (h *MetricsStreamHandler) writeSnapshot(stream *sseStream, sub *hub.Subscription) error {

	snapshot := sub.Snapshot()
	if len(snapshot) == 0 {
		return nil
	}

	batch := make(metricsEntities.Batch, len(snapshot))

	for _, update := range snapshot {
		batch.Add(update)
	}

	return stream.writeMetrics(batch)
}

// reportDropped logs updates the hub had to evict because the client is slower
// than the collectors. Live values are recoverable from the history API, so a
// dropped update is not a data loss for a chart.
func (h *MetricsStreamHandler) reportDropped(ctx context.Context, sub *hub.Subscription) {

	dropped := sub.TakeDropped()
	if dropped == 0 {
		return
	}

	h.logger.WarnContext(ctx, "sse: slow client, live updates were dropped",
		"subscription", sub.ID(),
		"dropped", dropped,
	)
}

func writeSSEHeaders(w http.ResponseWriter) {
	header := w.Header()
	header.Set("Content-Type", sseContentType)
	header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	header.Set("Connection", "keep-alive")
	// Disables proxy buffering (nginx and friends), otherwise events are
	// accumulated instead of being delivered as they happen.
	header.Set("X-Accel-Buffering", "no")
}

// sseStream writes Server-Sent Events and flushes every message.
type sseStream struct {
	writer  http.ResponseWriter
	flusher http.Flusher
}

func (s *sseStream) writeMetrics(batch metricsEntities.Batch) error {

	payload, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("marshal metrics batch: %w", err)
	}

	if _, err := fmt.Fprintf(s.writer, "event: %s\ndata: %s\n\n", sseEventMetrics, payload); err != nil {
		return fmt.Errorf("write metrics event: %w", err)
	}

	s.flusher.Flush()

	return nil
}

func (s *sseStream) writeComment(text string) error {

	if _, err := fmt.Fprintf(s.writer, ": %s\n\n", text); err != nil {
		return fmt.Errorf("write comment: %w", err)
	}

	s.flusher.Flush()

	return nil
}

// writeRetry advertises the reconnect delay the browser should use.
func (s *sseStream) writeRetry(interval time.Duration) error {

	if _, err := fmt.Fprintf(s.writer, "retry: %d\n\n", interval.Milliseconds()); err != nil {
		return fmt.Errorf("write retry: %w", err)
	}

	s.flusher.Flush()

	return nil
}

// MetricsHistory returns stored metric series.
func (h *Handler) MetricsHistory(ctx context.Context, req openapi.MetricsHistoryRequestObject) (openapi.MetricsHistoryResponseObject, error) {

	filter, err := historyFilter(req.Params)
	if err != nil {
		return openapi.MetricsHistory400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	history, err := h.metricsService.History(ctx, filter)
	if err != nil {
		return openapi.MetricsHistory400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return openapi.MetricsHistory200JSONResponse(history), nil
}

// historyFilter converts query parameters into a storage filter.
//
// Unknown metric names are rejected instead of being silently ignored, so a
// typo in a chart configuration is visible immediately.
func historyFilter(params openapi.MetricsHistoryParams) (metricsEntities.HistoryFilter, error) {

	var filter metricsEntities.HistoryFilter

	if params.From != nil {
		filter.From = *params.From
	}

	if params.To != nil {
		filter.To = *params.To
	}

	if params.Limit != nil {
		if *params.Limit <= 0 {
			return filter, errors.New("limit must be a positive number")
		}

		filter.Limit = *params.Limit
	}

	if params.Names != nil {
		for _, raw := range strings.Split(*params.Names, ",") {
			name := metricsEntities.Name(strings.TrimSpace(raw))
			if name == "" {
				continue
			}

			if !metricsEntities.IsKnown(name) {
				return filter, fmt.Errorf("unknown metric %q", name)
			}

			filter.Names = append(filter.Names, name)
		}
	}

	return filter, nil
}
