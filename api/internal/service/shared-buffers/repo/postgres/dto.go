package postgres

import (
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
)

type buffersHitRate struct {
	HitRatio       float64 `db:"hit_ratio"`
	HitsPerSecond  int64   `db:"hits"`
	ReadsPerSecond int64   `db:"reads"`
}

func buildBuffersHitRate(db buffersHitRate) sharedBuffers.BuffersHitRate {

	return sharedBuffers.BuffersHitRate{
		HitRatio:       db.HitRatio,
		HitsPerSecond:  db.HitsPerSecond,
		ReadsPerSecond: db.ReadsPerSecond,
	}
}
