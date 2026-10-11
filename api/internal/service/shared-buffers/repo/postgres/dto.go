package postgres

import (
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
	"time"
)

type buffersHitRate struct {
	HitRatio   float64   `db:"hit_ratio"`
	HitsTotal  int64     `db:"hits_total"`
	ReadsTotal int64     `db:"reads_total"`
	MeasuredAt time.Time `db:"measured_at"`
}

func buildBuffersHitRate(db buffersHitRate) sharedBuffers.BuffersHitRate {

	return sharedBuffers.BuffersHitRate{
		HitRatio:   db.HitRatio,
		HitsTotal:  db.HitsTotal,
		ReadsTotal: db.ReadsTotal,
		Timestamp:  db.MeasuredAt,
	}
}
