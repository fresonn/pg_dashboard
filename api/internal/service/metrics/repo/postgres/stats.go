package postgres

import (
	"context"
	metricsEntities "dashboard/api/internal/model/metrics"
)

/*
Connections and max_connections are returned by a single query: the pipeline
must not open one round trip per indicator.

The counters include every backend visible in pg_stat_activity, which is the
same set Postgres uses for the "connections" gauge in monitoring tools.
*/
const PG_CLUSTER_CONNECTIONS_QUERY = `
SELECT
    count(*)::bigint AS connections,
    COALESCE(
        (SELECT setting::bigint FROM pg_settings WHERE name = 'max_connections'),
        0
    ) AS max_connections,
    now() AS measured_at
FROM pg_stat_activity;
`

func (s *Storage) ClusterConnections(ctx context.Context) (metricsEntities.ConnectionsStats, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return metricsEntities.ConnectionsStats{}, err
	}

	var data connectionsStats

	err = db.GetContext(ctx, &data, PG_CLUSTER_CONNECTIONS_QUERY)
	if err != nil {
		return metricsEntities.ConnectionsStats{}, err
	}

	return buildConnectionsStats(data), nil
}

// pg_database_size is relatively expensive, so it is collected rarely and only
// for the database the application is connected to.
const PG_CURRENT_DATABASE_SIZE_QUERY = `
SELECT
    current_database() AS database_name,
    pg_database_size(current_database())::bigint AS size_bytes,
    now() AS measured_at;
`

func (s *Storage) CurrentDatabaseSize(ctx context.Context) (metricsEntities.DatabaseSizeStats, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return metricsEntities.DatabaseSizeStats{}, err
	}

	var data databaseSizeStats

	err = db.GetContext(ctx, &data, PG_CURRENT_DATABASE_SIZE_QUERY)
	if err != nil {
		return metricsEntities.DatabaseSizeStats{}, err
	}

	return buildDatabaseSizeStats(data), nil
}
