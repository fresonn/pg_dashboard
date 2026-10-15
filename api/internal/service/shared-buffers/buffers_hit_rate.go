package sharedbuffers

import (
	"context"
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
)

// BuffersHitRate returns the shared buffers cache efficiency of the cluster and
// of the database the application is connected to.
func (s *Service) BuffersHitRate(ctx context.Context) (sharedBuffers.BuffersHitRate, error) {

	rate, err := s.pg.BuffersHitRate(ctx)
	if err != nil {
		s.logger.Error("postgres: get shared_buffers hit rate", "error", err)
		return sharedBuffers.BuffersHitRate{}, err
	}

	return rate, nil
}
