package postgres

import (
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
	"time"
)

type buffersHitRate struct {
	DatabaseName       string    `db:"database_name"`
	ClusterHitsTotal   int64     `db:"cluster_hits_total"`
	ClusterReadsTotal  int64     `db:"cluster_reads_total"`
	ClusterHitRatio    float64   `db:"cluster_hit_ratio"`
	DatabaseHitsTotal  int64     `db:"database_hits_total"`
	DatabaseReadsTotal int64     `db:"database_reads_total"`
	DatabaseHitRatio   float64   `db:"database_hit_ratio"`
	MeasuredAt         time.Time `db:"measured_at"`
}

func buildBuffersHitRate(db buffersHitRate) sharedBuffers.BuffersHitRate {

	return sharedBuffers.BuffersHitRate{
		Cluster: sharedBuffers.HitRate{
			HitRatio:   db.ClusterHitRatio,
			HitsTotal:  db.ClusterHitsTotal,
			ReadsTotal: db.ClusterReadsTotal,
		},
		Database: sharedBuffers.HitRate{
			HitRatio:   db.DatabaseHitRatio,
			HitsTotal:  db.DatabaseHitsTotal,
			ReadsTotal: db.DatabaseReadsTotal,
		},
		DatabaseName: db.DatabaseName,
		Timestamp:    db.MeasuredAt,
	}
}
