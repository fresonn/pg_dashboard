package http

import (
	"context"
	"dashboard/api/internal/model/cluster"
	"dashboard/api/internal/model/database"
	metricsEntities "dashboard/api/internal/model/metrics"
	"dashboard/api/internal/model/role"
)

type UseCases interface {
	SwitchDatabase(ctx context.Context, id int) (cluster.Status, error)
}

type ClusterService interface {
	Connect(ctx context.Context, conn cluster.NewConnection) (cluster.Status, error)
	PostgresStatus(ctx context.Context) cluster.Status
	Uptime(ctx context.Context) (cluster.PostgresUptime, error)
	Version(ctx context.Context) (cluster.PostgresVersion, error)
	PostmasterSettings(ctx context.Context) (cluster.PostmasterSettings, error)
	Disconnect(ctx context.Context) error
}

type DatabaseService interface {
	Database(ctx context.Context, id int) (database.Database, error)
	DatabasesDetailed(ctx context.Context, filter database.DatabasesFilter) ([]database.DatabaseDetails, error)
	StatsOverview(ctx context.Context, databaseID int) (database.DatabaseStatsOverview, error)
}

type RoleService interface {
	Roles(ctx context.Context) ([]role.RoleView, error)
}

type MetricsService interface {
	History(ctx context.Context, filter metricsEntities.HistoryFilter) (metricsEntities.History, error)
}
