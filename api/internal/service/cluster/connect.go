package cluster

import (
	"context"
	"dashboard/api/internal/config"
	"dashboard/api/internal/model/cluster"

	"errors"
)

// todo: cluster.AuthData -> cluster.NewConnection
func (s *Service) Connect(ctx context.Context, authData cluster.AuthData) (cluster.Status, error) {

	s.logger.Info("try to establish postgres connection")

	authData.PrettyLog()

	if ok := s.pgManager.IsConnected(); ok {
		s.logger.WarnContext(ctx, "connection already established")
		return cluster.Status{}, errors.New("connection already established")
	}

	err := s.validate.Struct(&authData)
	if err != nil {
		s.logger.ErrorContext(ctx, "connection validation", "error", err)
		return cluster.Status{}, err
	}

	conn := config.Connection{
		Host:     authData.Host,
		Port:     authData.Port,
		User:     authData.User,
		Password: authData.Password,
		SSLMode:  authData.SSLMode,
	}

	if authData.Database != nil {
		conn.Database = *authData.Database
	} else {
		conn.Database = "postgres"
	}

	err = s.pgManager.UpdateConnection(ctx, conn)
	if err != nil {
		s.logger.Error("set connection", "error", err)
		return cluster.Status{}, err
	}

	s.logger.Info("postgres connection established")

	return cluster.Status{
		ConnectionStatus: s.pgManager.Status(),
		CurrentUser:      &conn.User,
		CurrentDatabase:  &conn.Database,
	}, nil
}
