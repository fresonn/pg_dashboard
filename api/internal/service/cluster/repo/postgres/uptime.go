package postgres

import (
	"dashboard/api/internal/model/cluster"
)

const PG_RUNTIME_QUERY = "SELECT pg_postmaster_start_time() as cluster_started_at;"

func (s *Storage) Uptime() (cluster.PostgresUptime, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return cluster.PostgresUptime{}, err
	}

	var uptime cluster.PostgresUptime

	err = db.Get(&uptime, PG_RUNTIME_QUERY)
	if err != nil {
		return cluster.PostgresUptime{}, err
	}

	return uptime, nil
}
