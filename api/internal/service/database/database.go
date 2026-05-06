package database

import (
	"context"
	"dashboard/api/internal/model/database"
)

func (s *Service) Database(ctx context.Context, id int) (database.Database, error) {

	if db, ok := s.cache.Database(ctx, id); ok {
		s.logger.DebugContext(ctx, "get database from cache", "id", id, "value", db)
		return db, nil
	}

	db, err := s.pg.Database(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "get database by id", "id", id, "error", err)
		return database.Database{}, err
	}

	s.cache.SetDatabase(ctx, id, db)

	s.logger.DebugContext(ctx, "get database by id", "id", id, "data", db)

	return db, nil
}
