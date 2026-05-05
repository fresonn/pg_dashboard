package database

import (
	"context"
	"dashboard/api/internal/model/database"
)

type PostgresRepo interface {
	Database(ctx context.Context, id int) (database.Database, error)
	DatabasesDetails(ctx context.Context, filters database.DatabasesFilter) ([]database.DatabaseDetails, error)
}

type Cache interface {
	Database(ctx context.Context, id int) (database.Database, bool)
	SetDatabase(ctx context.Context, id int, db database.Database)
}
