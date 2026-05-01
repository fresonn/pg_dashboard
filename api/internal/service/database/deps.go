package database

import (
	"context"
	"dashboard/api/internal/model/database"
)

type PostgresRepo interface {
	DatabasesDetails(ctx context.Context, filters database.DatabasesFilter) ([]database.DatabaseDetails, error)
}
