package http

import (
	"context"
	"dashboard/api/gen/openapi"
	"dashboard/api/internal/model/cluster"
	"dashboard/api/internal/model/database"
	"dashboard/api/internal/model/role"
)

type ClusterUseCase interface {
	Connect(ctx context.Context, authData cluster.AuthData) (cluster.Status, error)
	PostgresStatus(ctx context.Context) cluster.Status
	Uptime(ctx context.Context) (cluster.PostgresUptime, error)
	Version(ctx context.Context) (cluster.PostgresVersion, error)
	PostmasterSettings(ctx context.Context) (cluster.PostmasterSettings, error)
	Disconnect(ctx context.Context) error
}

type DatabaseUseCase interface {
	DatabasesDetailed(ctx context.Context, filter database.DatabasesFilter) ([]database.DatabaseDetails, error)
}

type RoleUseCase interface {
	Roles(ctx context.Context) ([]role.RoleView, error)
}

type Handler struct {
	cluster  ClusterUseCase
	roles    RoleUseCase
	database DatabaseUseCase
}

var _ openapi.StrictServerInterface = (*Handler)(nil)

func New(cluster ClusterUseCase, roles RoleUseCase, database DatabaseUseCase) *Handler {
	return &Handler{
		cluster:  cluster,
		roles:    roles,
		database: database,
	}
}
