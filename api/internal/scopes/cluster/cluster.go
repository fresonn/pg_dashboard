package cluster

import (
	"context"
	"dashboard/api/internal/config"
	"dashboard/api/internal/postgres"
	"dashboard/api/internal/scopes/cluster/entities"
	"dashboard/api/pkg/logger"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

type Cluster struct {
	config    config.AppConfig
	logger    *slog.Logger
	pgManager *postgres.Manager
	validate  *validator.Validate
	storage   ClusterStorage
	cache     Cache
}

type ClusterStorage interface {
	Version() (string, error)
	Uptime() (entities.PostgresUptime, error)
	PostmasterSettings(ctx context.Context, params []string) ([]entities.Setting, error)
	DatabasesDetails(ctx context.Context, filters entities.DatabasesFilter) ([]entities.DatabaseDetails, error)
}

type Cache interface {
	DeleteAll(ctx context.Context)
	PgVersion(ctx context.Context) (entities.PostgresVersion, bool)
	SetPgVersion(ctx context.Context, version entities.PostgresVersion)
	ClusterUptime(ctx context.Context) (entities.PostgresUptime, bool)
	SetClusterUptime(ctx context.Context, version entities.PostgresUptime)
}

type Options struct {
	Config          config.AppConfig
	Logger          *slog.Logger
	PostgresManager *postgres.Manager
	Storage         ClusterStorage
	Cache           Cache
}

func New(options Options) *Cluster {

	return &Cluster{
		config:    options.Config,
		logger:    logger.WithScopeLogger(options.Logger, "cluster"),
		validate:  validator.New(validator.WithRequiredStructEnabled()),
		storage:   options.Storage,
		pgManager: options.PostgresManager,
		cache:     options.Cache,
	}
}

func (c *Cluster) PostgresStatus(ctx context.Context) postgres.ConnectionStatus {

	c.logger.DebugContext(ctx, "get postres connection status")

	switch {
	case c.pgManager.IsConnected():
		return postgres.StatusConnected
	case c.pgManager.Connecting():
		return postgres.StatusConnecting
	default:
		return postgres.StatusDisconnected
	}
}
