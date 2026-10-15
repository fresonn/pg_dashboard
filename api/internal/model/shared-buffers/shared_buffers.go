package sharedbuffers

import "time"

// HitRate is the cache efficiency of a single scope: the whole cluster or one
// database.
//
// The counters are cumulative since the last statistics reset, so HitsTotal and
// ReadsTotal are exposed next to the ratio: a client can derive the hit rate for
// an arbitrary interval from the deltas.
type HitRate struct {
	HitRatio   float64
	HitsTotal  int64
	ReadsTotal int64
}

// BuffersHitRate is a snapshot of the shared buffers cache with two scopes: the
// cluster as a whole and the database the application is connected to.
//
// Both scopes are read by a single query over pg_stat_database, so the database
// level metric does not cost a second round trip.
type BuffersHitRate struct {
	Cluster      HitRate
	Database     HitRate
	DatabaseName string
	Timestamp    time.Time
}
