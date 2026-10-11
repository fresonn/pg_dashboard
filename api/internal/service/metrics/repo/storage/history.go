package storage

import (
	"bytes"
	metricsEntities "dashboard/api/internal/model/metrics"
	pb "dashboard/api/internal/service/metrics/repo/storage/codec"
	"fmt"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"
	"google.golang.org/protobuf/proto"
)

const (
	// metricNamespace isolates metric history keys inside the shared Badger.
	metricNamespace = "metric"
	// metricKeyParts is the number of parts of "metric:{name}:{timestamp}".
	metricKeyParts = 3
	// deleteBatchSize bounds a single retention transaction.
	deleteBatchSize = 1000
	// maxPointKeyWidth keeps timestamps lexicographically sortable.
	maxPointKeyWidth = 19
)

var (
	metricNamespacePrefix = []byte(metricNamespace + ":")
	metricKeySeparator    = []byte(":")
)

func pointKey(name metricsEntities.Name, timestamp time.Time) []byte {
	return fmt.Appendf(nil, "%s:%s:%0*d", metricNamespace, name, maxPointKeyWidth, timestamp.UTC().UnixMilli())
}

func pointKeyPrefix(name metricsEntities.Name) []byte {
	return fmt.Appendf(nil, "%s:%s:", metricNamespace, name)
}

func pointKeyTimestamp(key []byte) (int64, bool) {
	parts := bytes.SplitN(key, metricKeySeparator, metricKeyParts)

	if len(parts) != metricKeyParts {
		return 0, false
	}

	milliseconds, err := strconv.ParseInt(string(parts[2]), 10, 64)
	if err != nil {
		return 0, false
	}

	return milliseconds, true
}

// Append writes one history point per measurement.
//
// All measurements of a single collection share one Badger transaction, so a
// partially written collection is impossible.
func (s *Storage) Append(updates []metricsEntities.Update) error {

	if len(updates) == 0 {
		return nil
	}

	return s.db.Update(func(txn *badger.Txn) error {
		for _, update := range updates {
			if err := update.Validate(); err != nil {
				return err
			}

			data, err := proto.Marshal(EncodePoint(metricsEntities.SeriesPoint{
				Timestamp: update.Timestamp,
				Value:     update.Value,
			}))
			if err != nil {
				return err
			}

			if err := txn.Set(pointKey(update.Name, update.Timestamp), data); err != nil {
				return err
			}
		}

		return nil
	})
}

// Range returns the requested series inside the time range.
//
// A metric without points is present in the response with an empty list, so a
// client can distinguish "no data yet" from "unknown metric".
func (s *Storage) Range(filter metricsEntities.HistoryFilter) (metricsEntities.History, error) {

	filter = filter.Normalize()

	series := make(map[metricsEntities.Name][]metricsEntities.SeriesPoint, len(filter.Names))

	for _, name := range filter.Names {
		points, err := s.rangeSeries(name, filter.From, filter.To, filter.Limit)
		if err != nil {
			return metricsEntities.History{}, err
		}

		series[name] = points
	}

	return metricsEntities.History{Series: series}, nil
}

// rangeSeries reads a single series and keeps the newest limit points.
func (s *Storage) rangeSeries(
	name metricsEntities.Name,
	from time.Time,
	to time.Time,
	limit int,
) ([]metricsEntities.SeriesPoint, error) {

	prefix := pointKeyPrefix(name)
	start := pointKey(name, from)
	end := pointKey(name, to)

	points := make([]metricsEntities.SeriesPoint, 0, min(limit, 64))

	err := s.db.View(func(txn *badger.Txn) error {

		options := badger.DefaultIteratorOptions
		options.Prefix = prefix
		options.PrefetchValues = true
		options.PrefetchSize = 100

		iterator := txn.NewIterator(options)
		defer iterator.Close()

		for iterator.Seek(start); iterator.ValidForPrefix(prefix); iterator.Next() {

			item := iterator.Item()

			if bytes.Compare(item.Key(), end) > 0 {
				break
			}

			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			var message pb.MetricPoint

			if err := proto.Unmarshal(value, &message); err != nil {
				return err
			}

			points = append(points, DecodePoint(&message))

			// Retention usually keeps the series short, but a large explicit
			// range must not grow the response without a bound.
			if len(points) > limit {
				points = points[1:]
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return points, nil
}

// DeleteBefore drops every point older than the given time and returns the
// number of deleted points.
//
// Deletion happens in bounded batches so a large cleanup does not hold a single
// long running transaction.
func (s *Storage) DeleteBefore(before time.Time) (int, error) {

	cutoff := before.UTC().UnixMilli()
	deleted := 0

	for {
		keys, err := s.expiredKeys(cutoff)
		if err != nil {
			return deleted, err
		}

		if len(keys) == 0 {
			return deleted, nil
		}

		err = s.db.Update(func(txn *badger.Txn) error {
			for _, key := range keys {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return deleted, err
		}

		deleted += len(keys)

		if len(keys) < deleteBatchSize {
			return deleted, nil
		}
	}
}

func (s *Storage) expiredKeys(cutoff int64) ([][]byte, error) {

	keys := make([][]byte, 0, deleteBatchSize)

	err := s.db.View(func(txn *badger.Txn) error {

		options := badger.DefaultIteratorOptions
		options.Prefix = metricNamespacePrefix
		options.PrefetchValues = false

		iterator := txn.NewIterator(options)
		defer iterator.Close()

		for iterator.Rewind(); iterator.ValidForPrefix(metricNamespacePrefix); iterator.Next() {

			item := iterator.Item()

			timestamp, ok := pointKeyTimestamp(item.Key())
			if !ok {
				// A foreign key inside the namespace is not ours to delete.
				continue
			}

			if timestamp >= cutoff {
				continue
			}

			keys = append(keys, item.KeyCopy(nil))

			if len(keys) >= deleteBatchSize {
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return keys, nil
}
