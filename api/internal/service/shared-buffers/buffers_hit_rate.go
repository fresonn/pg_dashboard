package sharedbuffers

import (
	"context"
	sharedBuffers "dashboard/api/internal/model/shared-buffers"
	"fmt"
)

func (s *Service) ClusterBuffersHitRate(ctx context.Context) (sharedBuffers.BuffersHitRate, error) {

	rate, err := s.pg.ClusterBuffersHitRate(ctx)
	if err != nil {
		s.logger.Error("postgres: get shared_buffers hit rate", "error", err)
		return sharedBuffers.BuffersHitRate{}, err
	}

	fmt.Printf("%+v\n", rate)

	return rate, nil

}
