package storage

import (
	"context"
	"dashboard/api/internal/scopes/cluster/entities"
	"dashboard/api/internal/utils"
	"fmt"
)

const DATABASES_DETAILS_QUERY = `
WITH databases_cte AS (
  SELECT
  	oid,
    datname AS name,
    datdba AS owner_id,
    encoding AS encoding_id,
    datcollate AS collate,
    datctype AS ctype,
    datistemplate AS is_template,
    datallowconn AS allow_connections,
    datconnlimit AS connection_limit,
    pg_database_size(datname) AS size_bytes
  FROM pg_database
),
connections AS (
  SELECT
    datname AS name,
    count(*) AS active_connections
  FROM pg_stat_activity
  WHERE state = 'active'
  GROUP BY datname
)
SELECT
  d.oid,
  d.name,
  pg_get_userbyid(d.owner_id) AS owner,
  pg_encoding_to_char(d.encoding_id) AS encoding,
  d.collate,
  d.ctype,
  d.is_template,
  d.allow_connections,
  d.connection_limit,
  d.size_bytes,
  COALESCE(c.active_connections, 0) AS active_connections
FROM databases_cte d
LEFT JOIN connections c ON c.name = d.name`

func (s *Storage) DatabasesDetails(ctx context.Context, filter entities.DatabasesFilter) ([]entities.DatabaseDetails, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return nil, err
	}

	query := generateDatabasesDetailsQuery(DATABASES_DETAILS_QUERY, filter)

	fmt.Println(query)

	var dtos []DatabaseDetails

	err = db.SelectContext(ctx, &dtos, query)
	if err != nil {
		return nil, err
	}

	databases := make([]entities.DatabaseDetails, 0, len(dtos))

	for _, d := range dtos {
		e := toDatabaseDetailsEntity(d)
		// app function looks better here
		e.SizePretty = utils.PrettyByteSize(d.SizeBytes)

		databases = append(databases, e)
	}

	return databases, nil
}

func generateDatabasesDetailsQuery(query string, filter entities.DatabasesFilter) string {
	if filter.Sort == "" {
		return query
	}

	orderBy := ""
	switch filter.Sort {
	case "connection":
		orderBy = "COALESCE(c.active_connections, 0)"
		// orderBy = "c.active_connections"
	case "size":
		orderBy = "d.size_bytes"
	default:
		orderBy = "d.size_bytes"
	}

	sortMode := "DESC"
	switch filter.Order {
	case "asc":
		sortMode = "ASC"
	case "desc":
		sortMode = "DESC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, sortMode)
	return query
}
