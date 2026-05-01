package postgres

import (
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/logger"
	"dashboard/api/internal/infra/postgres"
)

type Storage struct {
	config    config.AppConfig
	logger    logger.Logger
	pgManager *postgres.Manager
}

func New(config config.AppConfig, logger logger.Logger, pgManager *postgres.Manager) *Storage {

	return &Storage{
		config:    config,
		logger:    logger,
		pgManager: pgManager,
	}
}
