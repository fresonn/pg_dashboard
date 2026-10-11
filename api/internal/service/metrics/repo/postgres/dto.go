package postgres

import (
	metricsEntities "dashboard/api/internal/model/metrics"
	"time"
)

type connectionsStats struct {
	Connections    int64     `db:"connections"`
	MaxConnections int64     `db:"max_connections"`
	MeasuredAt     time.Time `db:"measured_at"`
}

func buildConnectionsStats(dto connectionsStats) metricsEntities.ConnectionsStats {

	return metricsEntities.ConnectionsStats{
		Connections:    dto.Connections,
		MaxConnections: dto.MaxConnections,
		Timestamp:      dto.MeasuredAt,
	}
}

type databaseSizeStats struct {
	Database   string    `db:"database_name"`
	SizeBytes  int64     `db:"size_bytes"`
	MeasuredAt time.Time `db:"measured_at"`
}

func buildDatabaseSizeStats(dto databaseSizeStats) metricsEntities.DatabaseSizeStats {

	return metricsEntities.DatabaseSizeStats{
		Database:  dto.Database,
		SizeBytes: dto.SizeBytes,
		Timestamp: dto.MeasuredAt,
	}
}
