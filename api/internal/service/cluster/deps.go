package cluster

import (
	"context"
	"dashboard/api/internal/model/cluster"
)

type PostgresRepo interface {
	Version() (string, error)
	Uptime() (cluster.PostgresUptime, error)
	PostmasterSettings(ctx context.Context, params []string) ([]cluster.Setting, error)
}

type Cache interface {
	Clear(ctx context.Context)
	PgVersion(ctx context.Context) (cluster.PostgresVersion, bool)
	SetPgVersion(ctx context.Context, version cluster.PostgresVersion)
	ClusterUptime(ctx context.Context) (cluster.PostgresUptime, bool)
	SetClusterUptime(ctx context.Context, version cluster.PostgresUptime)
}
