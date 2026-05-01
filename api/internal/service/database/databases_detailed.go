package database

import (
	"context"
	"dashboard/api/internal/model/database"
)

func (s *Service) DatabasesDetailed(ctx context.Context, filter database.DatabasesFilter) ([]database.DatabaseDetails, error) {

	databases, err := s.storage.DatabasesDetails(ctx, filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "available databases", "error", err)
		return nil, err
	}

	return databases, nil
}
