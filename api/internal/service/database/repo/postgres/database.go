package postgres

import (
	"context"
	"dashboard/api/internal/model/database"
)

const DATABASE_BY_OID = `
SELECT
  d.oid,
  d.datname as name,
  pg_get_userbyid(d.datdba) AS owner,
  pg_encoding_to_char(d.encoding) AS encoding,
  d.datconnlimit as connection_limit,
  d.datallowconn as allow_connections,
  d.datistemplate as is_template,
  COALESCE(t.spcname, 'pg_default') AS tablespace,
  pg_database_size(d.oid) AS size_bytes,
  sd.description
FROM pg_database d
LEFT JOIN pg_tablespace t ON t.oid = d.dattablespace
LEFT JOIN pg_shdescription sd ON sd.objoid = d.oid AND sd.classoid = 'pg_database'::regclass
WHERE d.oid = $1;
`

func (s *Storage) Database(ctx context.Context, oid int) (database.Database, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return database.Database{}, err
	}

	var dto DatabaseByOID

	err = db.Get(&dto, DATABASE_BY_OID, oid)
	if err != nil {
		return database.Database{}, err
	}

	return toDatabase(dto), nil
}
