package metrics

import (
	"context"
	"dashboard/api/internal/infra/postgres"
	metricsEntities "dashboard/api/internal/model/metrics"
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
	"time"
)

// PgManager reports the state of the shared Postgres connection.
//
// Collectors never create their own connection: they use the application wide
// manager through this interface and skip the collection while it is not
// connected.
type PgManager interface {
	IsConnected() bool
	Status() postgres.ConnectionStatus
}

// PostgresRepo contains the metric specific SQL queries.
type PostgresRepo interface {
	ClusterConnections(ctx context.Context) (metricsEntities.ConnectionsStats, error)
	CurrentDatabaseSize(ctx context.Context) (metricsEntities.DatabaseSizeStats, error)
}

// SharedBuffers is the existing shared buffers service.
//
// The cache hit rate SQL lives there, so the pipeline reuses it instead of
// duplicating the query.
type SharedBuffers interface {
	BuffersHitRate(ctx context.Context) (sharedBuffers.BuffersHitRate, error)
}

// Publisher receives successfully collected measurements.
type Publisher interface {
	Publish(updates ...metricsEntities.Update)
}

// Storage keeps the metric history.
type Storage interface {
	// Append writes measurements, one point per metric.
	Append(updates []metricsEntities.Update) error
	// Range returns points of the requested metrics inside the time range.
	Range(filter metricsEntities.HistoryFilter) (metricsEntities.History, error)
	// DeleteBefore drops points older than the given time and returns the
	// number of deleted points.
	DeleteBefore(before time.Time) (int, error)
}
