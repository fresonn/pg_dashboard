package storage

import (
	"dashboard/api/internal/infra/postgres"
	"dashboard/api/internal/model/cluster"
)

type Setting struct {
	Name      string        `db:"name"`
	Setting   string        `db:"setting"`
	Unit      postgres.Text `db:"unit"`
	ShortDesc string        `db:"short_desc"`
}

func toSettingEntity(dto Setting) cluster.Setting {
	return cluster.Setting{
		Name:        dto.Name,
		Value:       dto.Setting,
		Unit:        dto.Unit.String(),
		Description: dto.ShortDesc,
	}
}
