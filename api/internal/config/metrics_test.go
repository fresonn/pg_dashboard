package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsVarsSanitized(t *testing.T) {

	t.Parallel()

	defaults := DefaultMetricsVars()

	t.Run("zero value gets defaults", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, defaults, MetricsVars{}.Sanitized())
	})

	t.Run("custom values are preserved", func(t *testing.T) {
		t.Parallel()

		custom := MetricsVars{
			BufferCacheHitInterval: 500 * time.Millisecond,
			ConnectionsInterval:    5 * time.Second,
			DatabaseSizeInterval:   time.Minute,
			QueryTimeout:           2 * time.Second,
			BatchWindow:            10 * time.Millisecond,
			BatchSize:              8,
			SubscriberBuffer:       4,
			HeartbeatInterval:      time.Second,
			RetryInterval:          2 * time.Second,
			HistoryRetention:       30 * time.Minute,
			HistoryCleanupInterval: 10 * time.Second,
			HistoryLimit:           120,
		}

		assert.Equal(t, custom, custom.Sanitized())
	})

	t.Run("invalid values fall back to defaults", func(t *testing.T) {
		t.Parallel()

		sanitized := MetricsVars{
			BufferCacheHitInterval: -time.Second,
			BatchWindow:            -time.Millisecond,
			BatchSize:              -1,
			SubscriberBuffer:       -10,
			HeartbeatInterval:      -time.Second,
			HistoryRetention:       -time.Hour,
			HistoryLimit:           -1,
		}.Sanitized()

		assert.Equal(t, defaults, sanitized)
	})
}

// TestMetricsEnvParsing documents the METRICS_* environment contract: nested
// struct with a prefix and time.Duration values.
func TestMetricsEnvParsing(t *testing.T) {

	t.Setenv("METRICS_BUFFER_CACHE_HIT_INTERVAL", "500ms")
	t.Setenv("METRICS_CONNECTIONS_INTERVAL", "7s")
	t.Setenv("METRICS_BATCH_WINDOW", "123ms")
	t.Setenv("METRICS_BATCH_SIZE", "16")
	t.Setenv("METRICS_HEARTBEAT_INTERVAL", "1m")
	t.Setenv("METRICS_HISTORY_LIMIT", "42")

	var parsed struct {
		Metrics MetricsVars `envPrefix:"METRICS_"`
	}

	require.NoError(t, env.Parse(&parsed))

	cfg := parsed.Metrics.Sanitized()

	assert.Equal(t, 500*time.Millisecond, cfg.BufferCacheHitInterval)
	assert.Equal(t, 7*time.Second, cfg.ConnectionsInterval)
	assert.Equal(t, 123*time.Millisecond, cfg.BatchWindow)
	assert.Equal(t, 16, cfg.BatchSize)
	assert.Equal(t, time.Minute, cfg.HeartbeatInterval)
	assert.Equal(t, 42, cfg.HistoryLimit)

	// Values that were not overridden keep their defaults.
	assert.Equal(t, DefaultMetricsVars().DatabaseSizeInterval, cfg.DatabaseSizeInterval)
	assert.Equal(t, DefaultMetricsVars().HistoryRetention, cfg.HistoryRetention)
}

func TestEnvironmentVarsMetricsPrefix(t *testing.T) {

	t.Parallel()

	field, ok := reflect.TypeOf(EnvironmentVars{}).FieldByName("Metrics")
	require.True(t, ok, "EnvironmentVars must expose the metrics settings")

	assert.Equal(t, "METRICS_", field.Tag.Get("envPrefix"))
}
