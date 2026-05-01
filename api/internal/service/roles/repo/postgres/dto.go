package postgres

import (
	"dashboard/api/internal/model/role"

	"github.com/lib/pq"
)

type RoleDetails struct {
	OID           int            `db:"oid"`
	RoleName      string         `db:"rolname"`
	CanLogin      bool           `db:"rolcanlogin"`
	RoleSuper     bool           `db:"rolsuper"`
	CanCreateRole bool           `db:"rolcreaterole"`
	CanCreateDB   bool           `db:"rolcreatedb"`
	Replication   bool           `db:"rolreplication"`
	MemberOf      pq.StringArray `db:"member_of"`
}

func toRoleEntity(dto RoleDetails) role.Role {

	return role.Role{
		ID:            dto.OID,
		Name:          dto.RoleName,
		CanLogin:      dto.CanLogin,
		IsSuper:       dto.RoleSuper,
		CanCreateRole: dto.CanCreateRole,
		CanCreateDB:   dto.CanCreateDB,
		Replication:   dto.Replication,
		MemberOf:      []string(dto.MemberOf),
	}
}
