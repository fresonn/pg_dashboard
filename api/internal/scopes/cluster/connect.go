package cluster

import (
	"context"
	"dashboard/api/internal/config"
	"dashboard/api/internal/scopes/cluster/entities"
	"errors"
)

func (c *Cluster) Connect(ctx context.Context, authData entities.AuthData) (*entities.Status, error) {

	c.logger.Info("try to establish postgres connection")

	authData.PrettyLog()

	if ok := c.pgManager.IsConnected(); ok {
		c.logger.WarnContext(ctx, "connection already established")
		return nil, errors.New("connection already established")
	}

	err := c.validate.Struct(&authData)
	if err != nil {
		c.logger.ErrorContext(ctx, "connection validation", "error", err)
		return nil, err
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

	err = c.pgManager.UpdateConnection(ctx, conn)
	if err != nil {
		c.logger.Error("set connection", "error", err)
		return nil, err
	}

	c.logger.Info("postgres connection established")

	return &entities.Status{
		ConnectionStatus: c.pgManager.Status(),
		CurrentUser:      &conn.User,
		CurrentDatabase:  &conn.Database,
	}, nil
}
