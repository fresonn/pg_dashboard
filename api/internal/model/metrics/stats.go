package metrics

import "time"

// ConnectionsStats is a raw cluster level connection snapshot.
//
// Both values come from a single query: the pipeline avoids one round trip per
// metric when a query returns several indicators at once.
type ConnectionsStats struct {
	Connections    int64
	MaxConnections int64
	Timestamp      time.Time
}

// DatabaseSizeStats is a raw snapshot of the size of the current database.
type DatabaseSizeStats struct {
	Database  string
	SizeBytes int64
	Timestamp time.Time
}
