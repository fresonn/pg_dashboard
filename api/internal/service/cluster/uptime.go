package cluster

import (
	"context"
	"dashboard/api/internal/model/cluster"
)

func (s *Service) Uptime(ctx context.Context) (cluster.PostgresUptime, error) {

	s.logger.DebugContext(ctx, "try to get cluster uptime")

	if uptime, ok := s.cache.ClusterUptime(ctx); ok {
		s.logger.DebugContext(ctx, "got cluster uptime", "started_at", uptime.StartedAt)
		return uptime, nil
	}

	uptime, err := s.storage.Uptime()
	if err != nil {
		s.logger.ErrorContext(ctx, "get cluster uptime", "error", err)
		return cluster.PostgresUptime{}, err
	}

	s.cache.SetClusterUptime(ctx, uptime)

	s.logger.DebugContext(ctx, "got cluster uptime", "started_at", uptime.StartedAt)

	return uptime, nil
}
