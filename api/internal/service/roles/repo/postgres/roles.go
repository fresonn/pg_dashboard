package postgres

import (
	"context"
	"dashboard/api/internal/model/role"
)

const ROLES_DETAILS_QUERY = `
SELECT
  r.oid,
  r.rolname,
  r.rolcanlogin,
  r.rolsuper,
  r.rolcreaterole,
  r.rolcreatedb,
  r.rolreplication,
  COALESCE(array_agg(mr.rolname) FILTER (WHERE mr.rolname IS NOT NULL), '{}') AS member_of
FROM pg_roles r
LEFT JOIN pg_auth_members m ON m.member = r.oid
LEFT JOIN pg_roles mr ON mr.oid = m.roleid
WHERE r.rolname NOT LIKE 'pg_%'
GROUP BY
  r.oid,
  r.rolname,
  r.rolcanlogin,
  r.rolsuper,
  r.rolcreaterole,
  r.rolcreatedb,
  r.rolreplication
ORDER BY r.rolname;
`

func (s *Storage) Roles(ctx context.Context) ([]role.Role, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return nil, err
	}

	var dtos []RoleDetails

	err = db.SelectContext(ctx, &dtos, ROLES_DETAILS_QUERY)
	if err != nil {
		return nil, err
	}

	roles := make([]role.Role, 0, len(dtos))

	for _, dto := range dtos {
		roles = append(roles, toRoleEntity(dto))
	}

	return roles, nil
}
