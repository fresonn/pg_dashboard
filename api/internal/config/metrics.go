package config

import "time"

// MetricsVars holds the tunables of the metrics pipeline. Every value is
// overridable through environment variables with the METRICS_ prefix, e.g.
// METRICS_BUFFER_CACHE_HIT_INTERVAL=500ms.
//
// Values that are not set (or set to a non-positive value) are replaced by the
// defaults from DefaultMetricsVars through Sanitized, so a broken environment
// cannot disable the pipeline silently or produce a busy loop.
type MetricsVars struct {
	// BufferCacheHitInterval is the collection interval of the shared buffers
	// cache hit rate metrics.
	BufferCacheHitInterval time.Duration `env:"BUFFER_CACHE_HIT_INTERVAL"`
	// ConnectionsInterval is the collection interval of connections metrics.
	ConnectionsInterval time.Duration `env:"CONNECTIONS_INTERVAL"`
	// DatabaseSizeInterval is the collection interval of the database size
	// metric (a relatively expensive query).
	DatabaseSizeInterval time.Duration `env:"DATABASE_SIZE_INTERVAL"`
	// QueryTimeout bounds a single collection, including the SQL query.
	QueryTimeout time.Duration `env:"QUERY_TIMEOUT"`
	// BatchWindow is the SSE batching window: updates that arrive within this
	// window are coalesced into a single event.
	BatchWindow time.Duration `env:"BATCH_WINDOW"`
	// BatchSize is the maximum number of metrics inside one SSE event.
	BatchSize int `env:"BATCH_SIZE"`
	// SubscriberBuffer is the per-subscriber hub queue size.
	SubscriberBuffer int `env:"SUBSCRIBER_BUFFER"`
	// HeartbeatInterval is the SSE comment interval that keeps proxies from
	// closing an idle connection.
	HeartbeatInterval time.Duration `env:"HEARTBEAT_INTERVAL"`
	// RetryInterval is advertised to the browser as the SSE reconnect delay.
	RetryInterval time.Duration `env:"RETRY_INTERVAL"`
	// HistoryRetention is how long metric points are kept in Badger.
	HistoryRetention time.Duration `env:"HISTORY_RETENTION"`
	// HistoryCleanupInterval is the retention cleanup period.
	HistoryCleanupInterval time.Duration `env:"HISTORY_CLEANUP_INTERVAL"`
	// HistoryLimit is the default number of points returned per metric by the
	// history API.
	HistoryLimit int `env:"HISTORY_LIMIT"`
}

// DefaultMetricsVars returns the defaults used in production and in tests.
func DefaultMetricsVars() MetricsVars {
	return MetricsVars{
		BufferCacheHitInterval: time.Second,
		ConnectionsInterval:    3 * time.Second,
		DatabaseSizeInterval:   30 * time.Second,
		QueryTimeout:           3 * time.Second,
		BatchWindow:            50 * time.Millisecond,
		BatchSize:              64,
		SubscriberBuffer:       32,
		HeartbeatInterval:      15 * time.Second,
		RetryInterval:          3 * time.Second,
		HistoryRetention:       time.Hour,
		HistoryCleanupInterval: time.Minute,
		HistoryLimit:           600,
	}
}

// Sanitized returns a copy with defaults applied to unset or invalid values.
func (m MetricsVars) Sanitized() MetricsVars {
	defaults := DefaultMetricsVars()

	if m.BufferCacheHitInterval <= 0 {
		m.BufferCacheHitInterval = defaults.BufferCacheHitInterval
	}

	if m.ConnectionsInterval <= 0 {
		m.ConnectionsInterval = defaults.ConnectionsInterval
	}

	if m.DatabaseSizeInterval <= 0 {
		m.DatabaseSizeInterval = defaults.DatabaseSizeInterval
	}

	if m.QueryTimeout <= 0 {
		m.QueryTimeout = defaults.QueryTimeout
	}

	if m.BatchWindow <= 0 {
		m.BatchWindow = defaults.BatchWindow
	}

	if m.BatchSize <= 0 {
		m.BatchSize = defaults.BatchSize
	}

	if m.SubscriberBuffer <= 0 {
		m.SubscriberBuffer = defaults.SubscriberBuffer
	}

	if m.HeartbeatInterval <= 0 {
		m.HeartbeatInterval = defaults.HeartbeatInterval
	}

	if m.RetryInterval <= 0 {
		m.RetryInterval = defaults.RetryInterval
	}

	if m.HistoryRetention <= 0 {
		m.HistoryRetention = defaults.HistoryRetention
	}

	if m.HistoryCleanupInterval <= 0 {
		m.HistoryCleanupInterval = defaults.HistoryCleanupInterval
	}

	if m.HistoryLimit <= 0 {
		m.HistoryLimit = defaults.HistoryLimit
	}

	return m
}
