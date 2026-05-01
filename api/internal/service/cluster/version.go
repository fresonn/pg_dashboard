package cluster

import (
	"context"
	"dashboard/api/internal/model/cluster"

	"strings"
)

func (s *Service) Version(ctx context.Context) (cluster.PostgresVersion, error) {

	s.logger.DebugContext(ctx, "try to get postgres version")

	if pgVersion, ok := s.cache.PgVersion(ctx); ok {
		s.logger.DebugContext(ctx, "got postgres version", "version", pgVersion.Version)
		return pgVersion, nil
	}

	rawVersion, err := s.storage.Version()
	if err != nil {
		s.logger.ErrorContext(ctx, "get postgres version", "error", err)
		return cluster.PostgresVersion{}, err
	}

	parts := strings.Split(rawVersion, ",")

	if len(parts) == 1 {
		return cluster.PostgresVersion{
			Version: strings.TrimSpace(parts[0]),
		}, nil
	}

	if len(parts) == 2 {
		return cluster.PostgresVersion{
			Version:  strings.TrimSpace(parts[0]),
			Compiler: strings.TrimSpace(parts[1]),
		}, nil
	}

	version := strings.TrimSpace(parts[0])
	bitDepth := strings.TrimSpace(parts[len(parts)-1])
	compiler := strings.TrimSpace(strings.Join(parts[1:len(parts)-1], ","))

	pgVersion := cluster.PostgresVersion{
		Version:  version,
		Compiler: compiler,
		BitDepth: bitDepth,
	}

	s.cache.SetPgVersion(ctx, pgVersion)

	s.logger.DebugContext(ctx, "got postgres version", "version", pgVersion.Version)

	return pgVersion, nil
}
