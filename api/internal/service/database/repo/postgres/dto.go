package postgres

import (
	"dashboard/api/internal/helper"
	"dashboard/api/internal/infra/postgres"
	"dashboard/api/internal/model/database"
)

type DatabaseDetails struct {
	OID  int    `db:"oid"`
	Name string `db:"name"`
	// If the OID is invalid or the role has been deleted, the function "pg_get_userbyid(owner_id)", will return NULL
	Owner postgres.Text `db:"owner"`
	// Encoding NULL, if the base is damaged, theoretically
	Encoding         postgres.Text `db:"encoding"`
	Collate          string        `db:"collate"`
	Ctype            string        `db:"ctype"`
	IsTemplate       bool          `db:"is_template"`
	AllowConnections bool          `db:"allow_connections"`
	ConnectionLimit  int           `db:"connection_limit"`
	SizeBytes        int64         `db:"size_bytes"`
	TotalConnections int           `db:"total_connections"`
}

func toDatabaseDetailsEntity(dto DatabaseDetails) database.DatabaseDetails {
	return database.DatabaseDetails{
		ID:               helper.IntToString(dto.OID),
		Name:             dto.Name,
		Owner:            dto.Owner.String(),
		Encoding:         dto.Encoding.String(),
		Collate:          dto.Collate,
		Ctype:            dto.Ctype,
		IsTemplate:       dto.IsTemplate,
		AllowConnections: dto.AllowConnections,
		ConnectionLimit:  dto.ConnectionLimit,
		SizeBytes:        dto.SizeBytes,
		SizePretty:       helper.PrettyByteSize(dto.SizeBytes),
		TotalConnections: dto.TotalConnections,
	}
}

type DatabaseByOID struct {
	OID  int    `db:"oid"`
	Name string `db:"name"`
	// If the OID is invalid or the role has been deleted, the function "pg_get_userbyid(owner_id)", will return NULL
	Owner postgres.Text `db:"owner"`
	// Encoding NULL, if the base is damaged, theoretically
	Encoding         postgres.Text `db:"encoding"`
	ConnectionLimit  int           `db:"connection_limit"`
	AllowConnections bool          `db:"allow_connections"`
	IsTemplate       bool          `db:"is_template"`
	TableSpace       string        `db:"tablespace"`
	SizeBytes        int64         `db:"size_bytes"`
	Description      postgres.Text `db:"description"`
}

func toDatabase(dto DatabaseByOID) database.Database {
	return database.Database{
		ID:               helper.IntToString(dto.OID),
		Name:             dto.Name,
		Owner:            dto.Owner.String(),
		Encoding:         dto.Encoding.String(),
		ConnectionLimit:  dto.ConnectionLimit,
		IsTemplate:       dto.IsTemplate,
		AllowConnections: dto.AllowConnections,
		Tablespace:       dto.TableSpace,
		SizeBytes:        dto.SizeBytes,
		SizePretty:       helper.PrettyByteSize(dto.SizeBytes),
		Description:      dto.Description.String(),
	}
}
