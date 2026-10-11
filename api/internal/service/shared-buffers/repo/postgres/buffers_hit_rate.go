package postgres

import (
	"context"
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
)

const PG_CLUSTER_BUFFER_HIT_RATE_QUERY = `
SELECT
    sum(blks_hit) AS hits,
    sum(blks_read) AS reads,
    round(
        100.0 * sum(blks_hit)
        / NULLIF(sum(blks_hit) + sum(blks_read), 0),
        2
    ) AS hit_ratio
FROM pg_stat_database;
`

func (s *Storage) ClusterBuffersHitRate(ctx context.Context) (sharedBuffers.BuffersHitRate, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return sharedBuffers.BuffersHitRate{}, err
	}

	var data buffersHitRate

	err = db.Get(&data, PG_CLUSTER_BUFFER_HIT_RATE_QUERY)
	if err != nil {
		return sharedBuffers.BuffersHitRate{}, err
	}

	return buildBuffersHitRate(data), nil
}
