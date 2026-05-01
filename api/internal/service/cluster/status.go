package cluster

import (
	"context"
	"dashboard/api/internal/model/cluster"
)

func (s *Service) PostgresStatus(ctx context.Context) cluster.Status {

	status := s.pgManager.Status()

	s.logger.DebugContext(ctx, "get postgres status", "status", status)

	var currentUser, currentDatabase *string

	connection := s.pgManager.Connection()
	if connection != nil {
		currentUser = &connection.User
		currentDatabase = &connection.Database
	}

	return cluster.Status{
		ConnectionStatus: status,
		CurrentUser:      currentUser,
		CurrentDatabase:  currentDatabase,
	}
}
