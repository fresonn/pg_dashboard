package storage

import (
	"dashboard/api/internal/postgres"
	"dashboard/api/internal/scopes/cluster/entities"
	"dashboard/api/internal/utils"
)

type Setting struct {
	Name      string        `db:"name"`
	Setting   string        `db:"setting"`
	Unit      postgres.Text `db:"unit"`
	ShortDesc string        `db:"short_desc"`
}

func toSettingEntity(dto Setting) entities.Setting {
	return entities.Setting{
		Name:        dto.Name,
		Value:       dto.Setting,
		Unit:        dto.Unit.String(),
		Description: dto.ShortDesc,
	}
}

type DatabaseDetails struct {
	OID  int    `db:"oid"`
	Name string `db:"name"`
	// If the OID is invalid or the role has been deleted, the function "pg_get_userbyid(owner_id)", will return NULL
	Owner postgres.Text `db:"owner"`
	// Encoding NULL, if the base is damaged, theoretically
	Encoding          postgres.Text `db:"encoding"`
	Collate           string        `db:"collate"`
	Ctype             string        `db:"ctype"`
	IsTemplate        bool          `db:"is_template"`
	AllowConnections  bool          `db:"allow_connections"`
	ConnectionLimit   int           `db:"connection_limit"`
	SizeBytes         int64         `db:"size_bytes"`
	ActiveConnections int           `db:"active_connections"`
}

func toDatabaseDetailsEntity(dto DatabaseDetails) entities.DatabaseDetails {
	return entities.DatabaseDetails{
		ID:               utils.IntToString(dto.OID),
		Name:             dto.Name,
		Owner:            dto.Owner.String(),
		Encoding:         dto.Encoding.String(),
		Collate:          dto.Collate,
		Ctype:            dto.Ctype,
		IsTemplate:       dto.IsTemplate,
		AllowConnections: dto.AllowConnections,
		ConnectionLimit:  dto.ConnectionLimit,
		SizeBytes:        dto.SizeBytes,
		// SizePretty - skip
		ActiveConnections: dto.ActiveConnections,
	}
}
