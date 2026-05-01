package database

import (
	"dashboard/api/internal/config"
	"dashboard/api/internal/postgres"
	"dashboard/api/pkg/logger"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

type Service struct {
	config    config.AppConfig
	logger    *slog.Logger
	pgManager *postgres.Manager
	validate  *validator.Validate
	storage   Storage
}

type Options struct {
	Config          config.AppConfig
	Logger          *slog.Logger
	PostgresManager *postgres.Manager
	Storage         Storage
}

func New(options Options) *Service {

	return &Service{
		config:    options.Config,
		logger:    logger.WithScopeLogger(options.Logger, "database"),
		validate:  validator.New(validator.WithRequiredStructEnabled()),
		storage:   options.Storage,
		pgManager: options.PostgresManager,
	}
}
