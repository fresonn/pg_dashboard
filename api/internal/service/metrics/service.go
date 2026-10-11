// Package metrics implements the centralised collection of PostgreSQL metrics.
//
// Design notes:
//
//   - Collectors are started once with the application context, not per HTTP
//     client, so the number of open charts does not affect the number of SQL
//     queries.
//   - Every collector has its own interval and its own goroutine. A collection
//     that takes longer than the interval only delays the next run of that
//     collector (ticks are dropped by time.Ticker), so runs never overlap and
//     the number of goroutines is bounded by the number of collectors.
//   - A failure inside one collector is logged and never stops the others.
//   - Measurements are published to the hub and appended to the history
//     storage. History lives outside the SSE transport, so a slow or absent
//     client cannot lose data.
package metrics

import (
	"context"
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/logger"
	metricsEntities "dashboard/api/internal/model/metrics"
	"sync"
	"time"
)

// Service owns the metric collectors and the history storage.
type Service struct {
	metricsCfg    config.MetricsVars
	logger        logger.Logger
	pgManager     PgManager
	pg            PostgresRepo
	sharedBuffers SharedBuffers
	hub           Publisher
	storage       Storage
}

// Options are the dependencies of the metrics service.
type Options struct {
	Config        config.AppConfig
	Logger        logger.Logger
	PgManager     PgManager
	PostgresRepo  PostgresRepo
	SharedBuffers SharedBuffers
	Hub           Publisher
	Storage       Storage
}

// New creates the metrics service.
func New(options Options) *Service {
	return &Service{
		metricsCfg:    options.Config.Env.Metrics.Sanitized(),
		logger:        options.Logger,
		pgManager:     options.PgManager,
		pg:            options.PostgresRepo,
		sharedBuffers: options.SharedBuffers,
		hub:           options.Hub,
		storage:       options.Storage,
	}
}

// History returns metric series for the REST API.
//
// Missing filter fields are filled with defaults: the whole known metric set,
// the configured retention window and the configured points limit.
func (s *Service) History(ctx context.Context, filter metricsEntities.HistoryFilter) (metricsEntities.History, error) {

	cfg := s.metricsCfg

	if filter.To.IsZero() {
		filter.To = time.Now().UTC()
	}

	if filter.From.IsZero() {
		filter.From = filter.To.Add(-cfg.HistoryRetention)
	}

	if filter.Limit <= 0 {
		filter.Limit = cfg.HistoryLimit
	}

	filter = filter.Normalize()

	history, err := s.storage.Range(filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "read metrics history failed", "error", err)
		return metricsEntities.History{}, err
	}

	s.logger.DebugContext(ctx, "metrics history read",
		"metrics", len(history.Series),
		"from", filter.From,
		"to", filter.To,
	)

	return history, nil
}

// Run starts the collectors and the history retention loop, then blocks until
// ctx is cancelled and waits for every goroutine to finish.
//
// It is intended to be started once from the application wiring with the
// application lifecycle context.
func (s *Service) Run(ctx context.Context) error {

	collectors := s.collectors()

	var wg sync.WaitGroup

	for _, item := range collectors {
		wg.Add(1)

		go func(item collector) {
			defer wg.Done()
			s.runCollector(ctx, item)
		}(item)
	}

	if s.metricsCfg.HistoryRetention > 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()
			s.runRetention(ctx)
		}()
	}

	s.logger.Info("metrics collectors started",
		"collectors", len(collectors),
		"retention", s.metricsCfg.HistoryRetention,
	)

	<-ctx.Done()

	wg.Wait()

	s.logger.Info("metrics collectors stopped")

	return nil
}

// publish sends measurements to the hub and stores them in the history.
//
// History failures are logged but do not affect live subscribers.
func (s *Service) publish(updates ...metricsEntities.Update) {

	if len(updates) == 0 {
		return
	}

	s.hub.Publish(updates...)

	if err := s.storage.Append(updates); err != nil {
		s.logger.Error("save metrics history failed", "error", err)
	}
}
