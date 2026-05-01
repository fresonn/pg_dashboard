package cluster

import (
	"context"
	"errors"
)

func (s *Service) Disconnect(ctx context.Context) error {

	s.logger.Info("try to disconnect postgres")

	isConnected := s.pgManager.IsConnected()

	if !isConnected {
		s.logger.Warn("already disconnected from postgres")
		return errors.New("already disconnected from postgres")
	}

	err := s.pgManager.Disconnect()
	if err != nil {
		s.logger.Error("disconnect from postgres", "error", err)
		return err
	}

	s.cache.Clear(ctx)

	s.logger.Info("postgres disconnected")

	return nil
}
