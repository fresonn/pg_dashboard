package metrics

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNumberMarshal(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name    string
		value   Number
		want    string
		wantErr bool
	}{
		{name: "integer", value: Int64(26), want: "26"},
		{name: "integer zero", value: Int64(0), want: "0"},
		{name: "negative integer", value: Int64(-1), want: "-1"},
		{name: "float", value: Float64(98.4), want: "98.4"},
		{name: "float zero", value: Float64(0), want: "0"},
		{name: "rounded float", value: Float64(99.999), want: "99.999"},
		{name: "nan", value: Float64(math.NaN()), wantErr: true},
		{name: "inf", value: Float64(math.Inf(1)), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestNumberUnmarshal(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantKind Kind
		wantInt  int64
		wantVal  float64
		wantErr  bool
	}{
		{name: "integer", input: "26", wantKind: KindInt, wantInt: 26},
		{name: "float", input: "98.4", wantKind: KindFloat, wantVal: 98.4},
		{name: "exponent stays float", input: "1e3", wantKind: KindFloat, wantVal: 1000},
		{name: "empty", input: " ", wantErr: true},
		{name: "string", input: `"26"`, wantErr: true},
		{name: "object", input: `{"value":26}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got Number

			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantKind, got.Kind())

			if tt.wantKind == KindInt {
				intValue, ok := got.Int()
				require.True(t, ok)
				assert.Equal(t, tt.wantInt, intValue)
				return
			}

			assert.InDelta(t, tt.wantVal, got.Float(), 0.0001)
		})
	}
}

func TestNumberRoundTrip(t *testing.T) {

	t.Parallel()

	for _, value := range []Number{Int64(1 << 40), Float64(98.42), Int64(0)} {
		encoded, err := json.Marshal(value)
		require.NoError(t, err)

		var decoded Number
		require.NoError(t, json.Unmarshal(encoded, &decoded))

		assert.True(t, value.Equal(decoded), "%s != %s", value, decoded)
		assert.Equal(t, value.Kind(), decoded.Kind())
	}
}

func TestUpdateValidate(t *testing.T) {

	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name    string
		update  Update
		wantErr bool
	}{
		{name: "valid", update: Update{Name: NameConnections, Value: Int64(5), Timestamp: now}},
		{name: "empty name", update: Update{Value: Int64(5), Timestamp: now}, wantErr: true},
		{name: "empty timestamp", update: Update{Name: NameConnections, Value: Int64(5)}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.update.Validate()

			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalidUpdate)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestBatchKeepsNewestSample(t *testing.T) {

	t.Parallel()

	older := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)
	newer := older.Add(time.Second)

	batch := make(Batch, 2)

	batch.Add(Update{Name: NameConnections, Value: Int64(26), Timestamp: newer})
	batch.Add(Update{Name: NameConnections, Value: Int64(20), Timestamp: older})
	batch.Add(Update{Name: NameBufferCacheHit, Value: Float64(98.4), Timestamp: older})

	require.Len(t, batch, 2)

	connections, ok := batch[NameConnections]
	require.True(t, ok)

	intValue, ok := connections.Value.Int()
	require.True(t, ok)
	assert.Equal(t, int64(26), intValue)
	assert.Equal(t, newer, connections.Timestamp)

	// A newer sample replaces the older one.
	batch.Add(Update{Name: NameConnections, Value: Int64(31), Timestamp: newer.Add(time.Second)})

	intValue, _ = batch[NameConnections].Value.Int()
	assert.Equal(t, int64(31), intValue)
}

// TestBatchJSONShape pins the SSE payload contract: one object keyed by metric
// name, each metric carrying its own value and measurement timestamp.
func TestBatchJSONShape(t *testing.T) {

	t.Parallel()

	measuredAt := time.Date(2026, 10, 11, 18, 0, 1, 0, time.UTC)

	batch := Batch{
		NameBufferCacheHit: {Value: Float64(98.4), Timestamp: measuredAt},
		NameConnections:    {Value: Int64(26), Timestamp: measuredAt.Add(-time.Second)},
	}

	payload, err := json.Marshal(batch)
	require.NoError(t, err)

	expected := `{"buffer_cache_hit":{"value":98.4,"timestamp":"2026-10-11T18:00:01Z"},` +
		`"connections":{"value":26,"timestamp":"2026-10-11T18:00:00Z"}}`

	assert.Equal(t, expected, string(payload))
}

func TestHistoryFilterNormalize(t *testing.T) {

	t.Parallel()

	now := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		filter := HistoryFilter{To: now}.Normalize()

		assert.Equal(t, Names(), filter.Names)
		assert.Equal(t, now.Add(-time.Hour), filter.From)
		assert.Equal(t, DefaultHistoryLimit, filter.Limit)
	})

	t.Run("unknown names are dropped", func(t *testing.T) {
		t.Parallel()

		filter := HistoryFilter{
			Names: []Name{NameConnections, "not_a_metric"},
			To:    now,
		}.Normalize()

		assert.Equal(t, []Name{NameConnections}, filter.Names)
	})

	t.Run("from after to is replaced", func(t *testing.T) {
		t.Parallel()

		filter := HistoryFilter{
			From: now.Add(time.Hour),
			To:   now,
		}.Normalize()

		assert.False(t, filter.From.After(filter.To))
	})

	t.Run("limit is clamped", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, MaxHistoryLimit, HistoryFilter{To: now, Limit: MaxHistoryLimit + 1}.Normalize().Limit)
		assert.Equal(t, DefaultHistoryLimit, HistoryFilter{To: now, Limit: -5}.Normalize().Limit)
		assert.Equal(t, 10, HistoryFilter{To: now, Limit: 10}.Normalize().Limit)
	})
}

func TestIsKnown(t *testing.T) {

	t.Parallel()

	assert.True(t, IsKnown(NameBufferCacheHit))
	assert.False(t, IsKnown("unknown"))
	assert.NotEmpty(t, Names())
}
