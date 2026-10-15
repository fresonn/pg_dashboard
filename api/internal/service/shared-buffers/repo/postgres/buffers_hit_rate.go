package postgres

import (
	"context"
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
)

/*
One query returns the same indicators for two scopes:

  - cluster  — every row of pg_stat_database (including the shared objects row
    with a null datname), the value the Cluster Overview widget shows;
  - database — only the row of the database the application is connected to, the
    value the database overview page shows.

Postgres rounds both ratios to two decimals, so the frontend prints the value as
it arrives without reformatting it.
*/
const PG_BUFFERS_HIT_RATE_QUERY = `
SELECT
    current_database() AS database_name,
    cluster_scope.hits_total   AS cluster_hits_total,
    cluster_scope.reads_total  AS cluster_reads_total,
    cluster_scope.hit_ratio    AS cluster_hit_ratio,
    database_scope.hits_total  AS database_hits_total,
    database_scope.reads_total AS database_reads_total,
    database_scope.hit_ratio   AS database_hit_ratio,
    now()                      AS measured_at
FROM (
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
        )::float8 AS hit_ratio
    FROM pg_stat_database
) AS cluster_scope,
(
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
        )::float8 AS hit_ratio
    FROM pg_stat_database
    WHERE datname = current_database()
) AS database_scope;
`

func (s *Storage) BuffersHitRate(ctx context.Context) (sharedBuffers.BuffersHitRate, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return sharedBuffers.BuffersHitRate{}, err
	}

	var data buffersHitRate

	err = db.GetContext(ctx, &data, PG_BUFFERS_HIT_RATE_QUERY)
	if err != nil {
		return sharedBuffers.BuffersHitRate{}, err
	}

	return buildBuffersHitRate(data), nil
}
