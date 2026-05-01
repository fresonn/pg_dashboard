package cluster

import (
	"dashboard/api/internal/infra/postgres"
	"log/slog"
	"time"
)

type AuthData struct {
	Host     string  `json:"host" validate:"required,hostname"`
	Port     int     `json:"port" validate:"gt=0"`
	User     string  `json:"user" validate:"required"`
	Password string  `json:"password" validate:"required"`
	Database *string `json:"database,omitempty"`
	SSLMode  string  `json:"sslmode"`
}

// todo: not the best solution
func (a AuthData) PrettyLog() {

	dbName := "nil"

	if a.Database != nil {
		dbName = *a.Database
	}

	slog.Info("connection data",
		slog.String("host", a.Host),
		slog.Int("port", a.Port),
		slog.String("user", a.User),
		slog.String("password", "*****"),
		slog.Any("database", dbName),
		slog.Any("sslmode", a.SSLMode),
	)
}

type PostgresVersion struct {
	Version  string `json:"version"`
	Compiler string `json:"compiler"`
	BitDepth string `json:"bitDepth"`
}

type PostgresUptime struct {
	StartedAt time.Time `json:"startedAt" db:"cluster_started_at"`
}

type Status struct {
	CurrentUser      *string
	CurrentDatabase  *string
	ConnectionStatus postgres.ConnectionStatus
}
