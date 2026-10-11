package storage

import (
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/logger"
	metricsEntities "dashboard/api/internal/model/metrics"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() logger.Logger {
	return logger.WithScopeLogger(
		logger.New(config.AppConfig{Env: config.EnvironmentVars{Runtime: config.TestRuntime}}),
		"storage-test",
	)
}

// newTestStorage opens a real Badger instance in a temporary directory: the
// history layer is the only component that cannot be faked meaningfully.
func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	db, err := badger.Open(badger.DefaultOptions(t.TempDir()))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return New(config.AppConfig{Env: config.EnvironmentVars{Runtime: config.TestRuntime}}, testLogger(), db)
}

func integerUpdate(name metricsEntities.Name, value int64, at time.Time) metricsEntities.Update {
	return metricsEntities.Update{Name: name, Value: metricsEntities.Int64(value), Timestamp: at}
}

func floatUpdate(name metricsEntities.Name, value float64, at time.Time) metricsEntities.Update {
	return metricsEntities.Update{Name: name, Value: metricsEntities.Float64(value), Timestamp: at}
}

func TestAppendAndRange(t *testing.T) {

	t.Parallel()

	storage := newTestStorage(t)

	base := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)

	require.NoError(t, storage.Append([]metricsEntities.Update{
		floatUpdate(metricsEntities.NameBufferCacheHit, 98.1, base),
		floatUpdate(metricsEntities.NameBufferCacheHit, 98.4, base.Add(time.Second)),
		floatUpdate(metricsEntities.NameBufferCacheHit, 97.7, base.Add(2*time.Second)),
		integerUpdate(metricsEntities.NameConnections, 26, base.Add(time.Second)),
	}))

	history, err := storage.Range(metricsEntities.HistoryFilter{
		Names: []metricsEntities.Name{metricsEntities.NameBufferCacheHit},
		From:  base.Add(-time.Second),
		To:    base.Add(time.Minute),
	})
	require.NoError(t, err)

	points := history.Series[metricsEntities.NameBufferCacheHit]
	require.Len(t, points, 3)

	for index, expected := range []float64{98.1, 98.4, 97.7} {
		assert.InDelta(t, expected, points[index].Value.Float(), 0.0001)
		assert.Equal(t, metricsEntities.KindFloat, points[index].Value.Kind())
		assert.True(t, points[index].Timestamp.Equal(base.Add(time.Duration(index)*time.Second)),
			"point %d timestamp = %s", index, points[index].Timestamp)
	}

	_, ok := history.Series[metricsEntities.NameConnections]
	assert.False(t, ok, "unrequested metrics must not be returned")
}

func TestRangeBarriers(t *testing.T) {

	t.Parallel()

	storage := newTestStorage(t)

	base := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)

	require.NoError(t, storage.Append([]metricsEntities.Update{
		integerUpdate(metricsEntities.NameConnections, 1, base),
		integerUpdate(metricsEntities.NameConnections, 2, base.Add(time.Minute)),
		integerUpdate(metricsEntities.NameConnections, 3, base.Add(2*time.Minute)),
	}))

	t.Run("time filter", func(t *testing.T) {
		history, err := storage.Range(metricsEntities.HistoryFilter{
			Names: []metricsEntities.Name{metricsEntities.NameConnections},
			From:  base.Add(30 * time.Second),
			To:    base.Add(90 * time.Second),
		})
		require.NoError(t, err)

		points := history.Series[metricsEntities.NameConnections]
		require.Len(t, points, 1)

		value, ok := points[0].Value.Int()
		require.True(t, ok)
		assert.Equal(t, int64(2), value)
	})

	t.Run("limit keeps the newest points", func(t *testing.T) {
		history, err := storage.Range(metricsEntities.HistoryFilter{
			Names: []metricsEntities.Name{metricsEntities.NameConnections},
			From:  base.Add(-time.Minute),
			To:    base.Add(time.Hour),
			Limit: 2,
		})
		require.NoError(t, err)

		points := history.Series[metricsEntities.NameConnections]
		require.Len(t, points, 2)

		first, _ := points[0].Value.Int()
		second, _ := points[1].Value.Int()
		assert.Equal(t, []int64{2, 3}, []int64{first, second})
	})

	t.Run("unknown names are ignored", func(t *testing.T) {
		history, err := storage.Range(metricsEntities.HistoryFilter{
			Names: []metricsEntities.Name{"not_a_metric"},
			From:  base,
			To:    base.Add(time.Hour),
		})
		require.NoError(t, err)
		assert.Empty(t, history.Series)
	})
}

// TestRangeReturnsEmptySeries ensures a metric without points is serialized as
// an empty JSON array, so the client can distinguish "no data yet" from a
// missing metric.
func TestRangeReturnsEmptySeries(t *testing.T) {

	t.Parallel()

	storage := newTestStorage(t)

	base := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)

	history, err := storage.Range(metricsEntities.HistoryFilter{
		Names: []metricsEntities.Name{metricsEntities.NameDatabaseSize},
		From:  base,
		To:    base.Add(time.Minute),
	})
	require.NoError(t, err)

	points, ok := history.Series[metricsEntities.NameDatabaseSize]
	require.True(t, ok)
	assert.Empty(t, points)
	assert.NotNil(t, points)

	payload, err := json.Marshal(history)
	require.NoError(t, err)

	assert.JSONEq(t, `{"series":{"database_size":[]}}`, string(payload))
}

func TestDeleteBefore(t *testing.T) {

	t.Parallel()

	storage := newTestStorage(t)

	now := time.Now().UTC()
	old := now.Add(-2 * time.Hour)

	require.NoError(t, storage.Append([]metricsEntities.Update{
		integerUpdate(metricsEntities.NameConnections, 1, old),
		integerUpdate(metricsEntities.NameConnections, 2, old.Add(time.Minute)),
		integerUpdate(metricsEntities.NameBufferCacheHit, 3, now),
	}))

	deleted, err := storage.DeleteBefore(now.Add(-time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 2, deleted)

	history, err := storage.Range(metricsEntities.HistoryFilter{
		Names: []metricsEntities.Name{metricsEntities.NameConnections},
		From:  now.Add(-3 * time.Hour),
		To:    now.Add(time.Hour),
	})
	require.NoError(t, err)
	assert.Empty(t, history.Series[metricsEntities.NameConnections])

	// Second run has nothing to delete.
	deleted, err = storage.DeleteBefore(now.Add(-time.Hour))
	require.NoError(t, err)
	assert.Zero(t, deleted)
}

func TestAppendRejectsInvalidUpdate(t *testing.T) {

	t.Parallel()

	storage := newTestStorage(t)

	err := storage.Append([]metricsEntities.Update{{Name: metricsEntities.NameConnections}})
	require.Error(t, err)
}

func TestProtobufKeepsValueKind(t *testing.T) {

	t.Parallel()

	storage := newTestStorage(t)

	base := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)

	require.NoError(t, storage.Append([]metricsEntities.Update{
		integerUpdate(metricsEntities.NameBufferHitsTotal, 1<<40, base),
		floatUpdate(metricsEntities.NameBufferCacheHit, 98.42, base.Add(time.Second)),
	}))

	history, err := storage.Range(metricsEntities.HistoryFilter{
		Names: []metricsEntities.Name{metricsEntities.NameBufferHitsTotal, metricsEntities.NameBufferCacheHit},
		From:  base.Add(-time.Second),
		To:    base.Add(time.Minute),
	})
	require.NoError(t, err)

	hits := history.Series[metricsEntities.NameBufferHitsTotal]
	require.Len(t, hits, 1)

	intValue, ok := hits[0].Value.Int()
	require.True(t, ok, "an integer metric must stay an integer")
	assert.Equal(t, int64(1<<40), intValue)

	ratio := history.Series[metricsEntities.NameBufferCacheHit]
	require.Len(t, ratio, 1)
	assert.Equal(t, metricsEntities.KindFloat, ratio[0].Value.Kind())
	assert.InDelta(t, 98.42, ratio[0].Value.Float(), 0.0001)
}

func TestConcurrentAppendAndRange(t *testing.T) {

	t.Parallel()

	storage := newTestStorage(t)

	base := time.Now().UTC()

	var wg sync.WaitGroup

	for worker := range 4 {
		wg.Add(1)

		go func(worker int) {
			defer wg.Done()

			for i := range 25 {
				update := integerUpdate(
					metricsEntities.NameConnections,
					int64(worker*25+i),
					base.Add(time.Duration(worker*25+i)*time.Millisecond),
				)

				require.NoError(t, storage.Append([]metricsEntities.Update{update}))
			}
		}(worker)
	}

	for range 4 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := storage.Range(metricsEntities.HistoryFilter{
				Names: []metricsEntities.Name{metricsEntities.NameConnections},
				From:  base.Add(-time.Minute),
				To:    base.Add(time.Hour),
			})

			require.NoError(t, err)
		}()
	}

	wg.Wait()

	history, err := storage.Range(metricsEntities.HistoryFilter{
		Names: []metricsEntities.Name{metricsEntities.NameConnections},
		From:  base.Add(-time.Minute),
		To:    base.Add(time.Hour),
	})
	require.NoError(t, err)
	assert.Len(t, history.Series[metricsEntities.NameConnections], 100)
}
