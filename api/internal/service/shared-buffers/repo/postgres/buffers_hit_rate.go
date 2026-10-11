package postgres

import (
	"context"
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
)

const PG_CLUSTER_BUFFER_HIT_RATE_QUERY = `
SELECT
    COALESCE(sum(blks_hit), 0)::bigint AS hits_total,
    COALESCE(sum(blks_read), 0)::bigint AS reads_total,
    COALESCE(
        round(
            100.0 * sum(blks_hit)
            / NULLIF(sum(blks_hit) + sum(blks_read), 0),
            2
        ),
        0
    )::float8 AS hit_ratio,
    now() AS measured_at
FROM pg_stat_database;
`

func (s *Storage) ClusterBuffersHitRate(ctx context.Context) (sharedBuffers.BuffersHitRate, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return sharedBuffers.BuffersHitRate{}, err
	}

	var data buffersHitRate

	err = db.GetContext(ctx, &data, PG_CLUSTER_BUFFER_HIT_RATE_QUERY)
	if err != nil {
		return sharedBuffers.BuffersHitRate{}, err
	}

	return buildBuffersHitRate(data), nil
}
