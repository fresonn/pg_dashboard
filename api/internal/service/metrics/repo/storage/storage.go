package storage

import (
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/logger"

	"github.com/dgraph-io/badger/v4"
)

/*
HISTORY:

	metric:{name}:{padded timestamp}  → point (protobuf)

The timestamp is zero padded to keep keys in lexicographic (and therefore
chronological) order, which allows range scans by prefix.

The history is shared with the rest of the application through the same Badger
instance and is not partitioned by cluster or database: series are identified by
the metric name only.
*/
type Storage struct {
	config config.AppConfig
	logger logger.Logger
	db     *badger.DB
}

func New(config config.AppConfig, logger logger.Logger, db *badger.DB) *Storage {
	return &Storage{
		config: config,
		logger: logger,
		db:     db,
	}
}
