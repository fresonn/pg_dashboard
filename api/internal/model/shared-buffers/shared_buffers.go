package sharedbuffers

import "time"

// BuffersHitRate is a cluster-wide snapshot of the shared buffers cache
// efficiency. The counters are cumulative since the last statistics reset, so
// HitsTotal and ReadsTotal are exposed next to the ratio: a client can derive
// the hit rate for an arbitrary interval from the deltas.
type BuffersHitRate struct {
	HitRatio   float64
	HitsTotal  int64
	ReadsTotal int64
	Timestamp  time.Time
}
