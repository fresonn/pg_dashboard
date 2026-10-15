package sharedbuffers

import (
	"context"
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
)

type PostgresRepo interface {
	BuffersHitRate(ctx context.Context) (sharedBuffers.BuffersHitRate, error)
}

type Storage interface{}
