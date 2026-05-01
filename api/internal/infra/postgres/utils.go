package postgres

import (
	"dashboard/api/internal/config"
	"fmt"
)

func connectionDSN(c config.Connection) string {
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}
