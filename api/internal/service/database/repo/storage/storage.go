package storage

import (
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/postgres"
	"log/slog"
)

type Storage struct {
	config    config.AppConfig
	logger    *slog.Logger
	pgManager *postgres.Manager
}

func New(config config.AppConfig, logger *slog.Logger, pgManager *postgres.Manager) *Storage {

	return &Storage{
		config:    config,
		logger:    logger,
		pgManager: pgManager,
	}
}
