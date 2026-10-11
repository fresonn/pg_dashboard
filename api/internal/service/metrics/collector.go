package metrics

import (
	"context"
	"dashboard/api/internal/config"
	metricsEntities "dashboard/api/internal/model/metrics"
	"time"
)

// collector is a single metric producer.
//
// A collector returns all indicators that a query provides, so several metrics
// may be collected with one SQL round trip.
type collector struct {
	name     string
	interval time.Duration
	collect  func(ctx context.Context) ([]metricsEntities.Update, error)
}

// collectors is the metric registry: adding a new metric means adding a
// collection function and a name in metrics.Names, then registering it here.
func (s *Service) collectors() []collector {

	cfg := s.metricsCfg

	return []collector{
		{
			name:     string(metricsEntities.NameBufferCacheHit),
			interval: cfg.BufferCacheHitInterval,
			collect:  s.collectBufferCacheHit,
		},
		{
			name:     string(metricsEntities.NameConnections),
			interval: cfg.ConnectionsInterval,
			collect:  s.collectConnections,
		},
		{
			name:     string(metricsEntities.NameDatabaseSize),
			interval: cfg.DatabaseSizeInterval,
			collect:  s.collectDatabaseSize,
		},
	}
}

// runCollector performs an immediate collection and then repeats it on the
// collector interval until the context is cancelled.
//
// The loop is sequential on purpose: it makes overlapping runs impossible even
// when a query is slower than the interval.
func (s *Service) runCollector(ctx context.Context, collector collector) {

	ticker := time.NewTicker(collector.interval)
	defer ticker.Stop()

	s.collectOnce(ctx, collector)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.collectOnce(ctx, collector)
		}
	}
}

// collectOnce collects and publishes measurements of a single collector.
//
// Any error is logged and swallowed: one broken collector must not stop the
// rest of the pipeline.
func (s *Service) collectOnce(ctx context.Context, collector collector) {

	if !s.pgManager.IsConnected() {
		s.logger.Debug("skip metrics collection, postgres is not connected",
			"collector", collector.name,
			"status", s.pgManager.Status(),
		)
		return
	}

	timeout := s.metricsCfg.QueryTimeout

	collectCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	updates, err := collector.collect(collectCtx)
	if err != nil {
		s.logger.Error("collect metric failed", "collector", collector.name, "error", err)
		return
	}

	s.publish(updates...)
}

// collectBufferCacheHit reuses the shared buffers service, which already owns
// the pg_stat_database query.
func (s *Service) collectBufferCacheHit(ctx context.Context) ([]metricsEntities.Update, error) {

	rate, err := s.sharedBuffers.ClusterBuffersHitRate(ctx)
	if err != nil {
		return nil, err
	}

	measuredAt := rate.Timestamp
	if measuredAt.IsZero() {
		measuredAt = time.Now().UTC()
	}

	return []metricsEntities.Update{
		{
			Name:      metricsEntities.NameBufferCacheHit,
			Value:     metricsEntities.Float64(rate.HitRatio),
			Timestamp: measuredAt,
		},
		{
			Name:      metricsEntities.NameBufferHitsTotal,
			Value:     metricsEntities.Int64(rate.HitsTotal),
			Timestamp: measuredAt,
		},
		{
			Name:      metricsEntities.NameBufferReadsTotal,
			Value:     metricsEntities.Int64(rate.ReadsTotal),
			Timestamp: measuredAt,
		},
	}, nil
}

// collectConnections collects two metrics with one query.
func (s *Service) collectConnections(ctx context.Context) ([]metricsEntities.Update, error) {

	stats, err := s.pg.ClusterConnections(ctx)
	if err != nil {
		return nil, err
	}

	measuredAt := stats.Timestamp
	if measuredAt.IsZero() {
		measuredAt = time.Now().UTC()
	}

	return []metricsEntities.Update{
		{
			Name:      metricsEntities.NameConnections,
			Value:     metricsEntities.Int64(stats.Connections),
			Timestamp: measuredAt,
		},
		{
			Name:      metricsEntities.NameMaxConnections,
			Value:     metricsEntities.Int64(stats.MaxConnections),
			Timestamp: measuredAt,
		},
	}, nil
}

func (s *Service) collectDatabaseSize(ctx context.Context) ([]metricsEntities.Update, error) {

	stats, err := s.pg.CurrentDatabaseSize(ctx)
	if err != nil {
		return nil, err
	}

	measuredAt := stats.Timestamp
	if measuredAt.IsZero() {
		measuredAt = time.Now().UTC()
	}

	s.logger.Debug("database size collected", "database", stats.Database, "bytes", stats.SizeBytes)

	return []metricsEntities.Update{
		{
			Name:      metricsEntities.NameDatabaseSize,
			Value:     metricsEntities.Int64(stats.SizeBytes),
			Timestamp: measuredAt,
		},
	}, nil
}

// runRetention periodically drops history points that are older than the
// configured retention window.
func (s *Service) runRetention(ctx context.Context) {

	cfg := s.metricsCfg

	s.cleanupHistory(cfg)

	ticker := time.NewTicker(cfg.HistoryCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanupHistory(cfg)
		}
	}
}

func (s *Service) cleanupHistory(cfg config.MetricsVars) {

	before := time.Now().UTC().Add(-cfg.HistoryRetention)

	deleted, err := s.storage.DeleteBefore(before)
	if err != nil {
		s.logger.Error("cleanup metrics history failed", "error", err)
		return
	}

	if deleted > 0 {
		s.logger.Info("metrics history cleaned up", "points", deleted, "before", before)
	}
}
