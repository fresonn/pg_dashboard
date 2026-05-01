package postgres

import (
	"context"
	"dashboard/api/internal/model/cluster"

	"github.com/jmoiron/sqlx"
)

const POSTMASTER_SETTINGS_QUERY = `
SELECT
    name,
    setting,
    unit,
    short_desc
FROM pg_settings
WHERE context = 'postmaster' AND name IN (?);`

func (s *Storage) PostmasterSettings(ctx context.Context, params []string) ([]cluster.Setting, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return nil, err
	}

	var dtos []Setting

	query, args, err := sqlx.In(POSTMASTER_SETTINGS_QUERY, params)
	if err != nil {
		return nil, err
	}

	query = db.Rebind(query)

	err = db.Select(&dtos, query, args...)
	if err != nil {
		return nil, err
	}

	settings := make([]cluster.Setting, 0, len(dtos))

	for _, dto := range dtos {
		settings = append(settings, toSettingEntity(dto))
	}

	return settings, nil
}
