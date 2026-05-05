package database

import (
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/logger"
	"dashboard/api/internal/infra/postgres"

	"github.com/go-playground/validator/v10"
)

type Service struct {
	config    config.AppConfig
	logger    logger.Logger
	pgManager *postgres.Manager
	validate  *validator.Validate
	pg        PostgresRepo
	cache     Cache
}

type Options struct {
	Config          config.AppConfig
	Logger          logger.Logger
	PostgresManager *postgres.Manager
	PostgresRepo    PostgresRepo
	Cache           Cache
}

func New(options Options) *Service {

	return &Service{
		config:    options.Config,
		logger:    options.Logger,
		validate:  validator.New(validator.WithRequiredStructEnabled()),
		pgManager: options.PostgresManager,
		pg:        options.PostgresRepo,
		cache:     options.Cache,
	}
}
