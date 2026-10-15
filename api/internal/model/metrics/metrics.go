// Package metrics contains the entities shared by the metrics pipeline:
// metric names, typed values, measurements and history points.
//
// The package has no dependency on transport, collectors or storage, so it can
// be reused by the HTTP layer, the collector service and the Badger codec.
package metrics

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Name identifies a single metric exposed by the pipeline.
//
// A name is a stable contract between the backend and the frontend: it is used
// as the key inside the SSE payload and inside the history REST response.
type Name string

const (
	// NameBufferCacheHit is the cumulative shared buffers cache hit ratio,
	// calculated by Postgres as blks_hit / (blks_hit + blks_read) in percents.
	NameBufferCacheHit Name = "buffer_cache_hit"
	// NameBufferHitsTotal is the cumulative number of shared buffer hits.
	NameBufferHitsTotal Name = "buffer_hits_total"
	// NameBufferReadsTotal is the cumulative number of shared buffer reads.
	NameBufferReadsTotal Name = "buffer_reads_total"
	// NameDatabaseBufferCacheHit is the cumulative shared buffers cache hit ratio
	// of the database the application is connected to.
	NameDatabaseBufferCacheHit Name = "database_buffer_cache_hit"
	// NameDatabaseBufferHitsTotal is the cumulative number of shared buffer hits
	// of the current database.
	NameDatabaseBufferHitsTotal Name = "database_buffer_hits_total"
	// NameDatabaseBufferReadsTotal is the cumulative number of shared buffer
	// reads of the current database.
	NameDatabaseBufferReadsTotal Name = "database_buffer_reads_total"
	// NameConnections is the current number of cluster connections.
	NameConnections Name = "connections"
	// NameMaxConnections is the configured max_connections value.
	NameMaxConnections Name = "max_connections"
	// NameDatabaseSize is the size of the current database in bytes.
	NameDatabaseSize Name = "database_size"
)

// ErrInvalidUpdate is returned when a measurement cannot be published.
var ErrInvalidUpdate = errors.New("invalid metric update")

// Names returns the list of all metric names known to the application.
//
// The order is stable (sorted) because it is used as a deterministic default
// for the history API.
func Names() []Name {
	return []Name{
		NameBufferCacheHit,
		NameBufferHitsTotal,
		NameBufferReadsTotal,
		NameDatabaseBufferCacheHit,
		NameDatabaseBufferHitsTotal,
		NameDatabaseBufferReadsTotal,
		NameConnections,
		NameMaxConnections,
		NameDatabaseSize,
	}
}

// IsKnown reports whether the name belongs to the registered metric set.
func IsKnown(name Name) bool {
	return slices.Contains(Names(), name)
}

// Kind describes how a Number value should be interpreted.
type Kind uint8

const (
	// KindInt is a whole number value (counters, gauges).
	KindInt Kind = iota
	// KindFloat is a fractional value (ratios, percents).
	KindFloat
)

// Number is a sealed numeric metric value.
//
// Metrics are numeric by nature, so instead of `any` (which would allow
// publishing arbitrary structures and would require runtime type switches in
// every consumer) a value keeps its kind explicitly. It marshals to a plain
// JSON number, which keeps the SSE/history payloads compact:
//
//	{"connections":{"value":26,"timestamp":"2026-10-11T18:00:00Z"}}
type Number struct {
	kind   Kind
	intV   int64
	floatV float64
}

// Int64 creates an integer metric value.
func Int64(v int64) Number {
	return Number{kind: KindInt, intV: v}
}

// Float64 creates a fractional metric value.
func Float64(v float64) Number {
	return Number{kind: KindFloat, floatV: v}
}

// Kind returns the value kind.
func (n Number) Kind() Kind {
	return n.kind
}

// Int returns the value as int64. The second return value reports whether the
// value is an integer.
func (n Number) Int() (int64, bool) {
	if n.kind != KindInt {
		return 0, false
	}

	return n.intV, true
}

// Float returns the value as float64 regardless of the underlying kind.
func (n Number) Float() float64 {
	if n.kind == KindInt {
		return float64(n.intV)
	}

	return n.floatV
}

// Equal reports whether two values are equal, including their kind.
func (n Number) Equal(other Number) bool {
	if n.kind != other.kind {
		return false
	}

	if n.kind == KindInt {
		return n.intV == other.intV
	}

	return n.floatV == other.floatV
}

// String implements fmt.Stringer, mostly for logs and test failure messages.
func (n Number) String() string {
	if n.kind == KindInt {
		return strconv.FormatInt(n.intV, 10)
	}

	return strconv.FormatFloat(n.floatV, 'f', -1, 64)
}

// MarshalJSON encodes the value as a JSON number.
func (n Number) MarshalJSON() ([]byte, error) {
	switch n.kind {
	case KindInt:
		return []byte(strconv.FormatInt(n.intV, 10)), nil

	case KindFloat:
		if math.IsNaN(n.floatV) || math.IsInf(n.floatV, 0) {
			return nil, fmt.Errorf("metrics: value is not a finite number: %v", n.floatV)
		}

		return []byte(strconv.FormatFloat(n.floatV, 'f', -1, 64)), nil

	default:
		return nil, fmt.Errorf("metrics: unknown value kind %d", n.kind)
	}
}

// UnmarshalJSON decodes a JSON number preserving the int/float distinction.
func (n *Number) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)

	if len(trimmed) == 0 {
		return errors.New("metrics: empty value")
	}

	if trimmed[0] == '"' {
		return errors.New("metrics: string values are not supported")
	}

	var number json.Number

	if err := json.Unmarshal(trimmed, &number); err != nil {
		return fmt.Errorf("metrics: decode value: %w", err)
	}

	if !strings.ContainsAny(number.String(), ".eE") {
		intValue, err := number.Int64()
		if err == nil {
			*n = Int64(intValue)
			return nil
		}
	}

	floatValue, err := number.Float64()
	if err != nil {
		return fmt.Errorf("metrics: decode value: %w", err)
	}

	*n = Float64(floatValue)

	return nil
}

// Update is a single measurement produced by a collector.
//
// Timestamp belongs to the measurement itself, not to a batch: metrics are
// collected with different intervals, so a batch may contain samples taken at
// different moments.
type Update struct {
	Name      Name
	Value     Number
	Timestamp time.Time
}

// Validate reports whether the measurement can be published or stored.
func (u Update) Validate() error {
	if u.Name == "" {
		return fmt.Errorf("%w: metric name is empty", ErrInvalidUpdate)
	}

	if u.Timestamp.IsZero() {
		return fmt.Errorf("%w: timestamp is empty for metric %q", ErrInvalidUpdate, u.Name)
	}

	return nil
}

// Sample returns the transport/history representation of the measurement.
func (u Update) Sample() Sample {
	return Sample{
		Value:     u.Value,
		Timestamp: u.Timestamp,
	}
}

// Sample is a value with its measurement timestamp.
type Sample struct {
	Value     Number    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// Batch is a set of measurements coalesced inside a batching window.
//
// One metric appears at most once: when a metric is updated several times
// within a window, the sample with the newest timestamp wins. The payload is a
// plain JSON object keyed by metric name, so a client can merge it into any
// local history without knowing the full metric list.
type Batch map[Name]Sample

// Add merges the measurement into the batch keeping the newest sample.
func (b Batch) Add(u Update) {
	current, ok := b[u.Name]
	if ok && current.Timestamp.After(u.Timestamp) {
		return
	}

	b[u.Name] = u.Sample()
}

// SeriesPoint is a single history point returned by the REST API.
type SeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     Number    `json:"value"`
}

// History is the REST representation of metric series.
type History struct {
	Series map[Name][]SeriesPoint `json:"series"`
}

// HistoryFilter describes a history query.
type HistoryFilter struct {
	Names []Name
	From  time.Time
	To    time.Time
	Limit int
}

// Normalize fills the filter with defaults and returns the effective metric
// list. Unknown names are ignored, so a client cannot ask for arbitrary Badger
// prefixes.
func (f HistoryFilter) Normalize() HistoryFilter {
	known := make(map[Name]struct{}, len(Names()))

	for _, name := range Names() {
		known[name] = struct{}{}
	}

	if len(f.Names) == 0 {
		f.Names = Names()
	} else {
		names := make([]Name, 0, len(f.Names))

		for _, name := range f.Names {
			if _, ok := known[name]; ok {
				names = append(names, name)
			}
		}

		f.Names = names
	}

	if f.To.IsZero() {
		f.To = time.Now().UTC()
	}

	if f.From.IsZero() || f.From.After(f.To) {
		f.From = f.To.Add(-time.Hour)
	}

	if f.Limit <= 0 {
		f.Limit = DefaultHistoryLimit
	}

	if f.Limit > MaxHistoryLimit {
		f.Limit = MaxHistoryLimit
	}

	return f
}

const (
	// DefaultHistoryLimit is the default number of points returned per metric.
	DefaultHistoryLimit = 600
	// MaxHistoryLimit bounds the history response size.
	MaxHistoryLimit = 5000
)
