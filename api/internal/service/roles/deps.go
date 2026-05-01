package roles

import (
	"context"
	"dashboard/api/internal/model/role"
)

type PostgresRepo interface {
	Roles(ctx context.Context) ([]role.Role, error)
}
