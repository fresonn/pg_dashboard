package entities

import "slices"

type RoleAccessLevel string

const (
	RoleAccessLevelAdmin    RoleAccessLevel = "admin"
	RoleAccessLevelStandard RoleAccessLevel = "standard"
	RoleAccessLevelElevated RoleAccessLevel = "elevated"
	RoleAccessLevelLimited  RoleAccessLevel = "limited"
)

type RoleAttribute string

// info: relies on https://www.postgresql.org/docs/current/role-attributes.html#ROLE-ATTRIBUTES
const (
	RoleAttributeSuperuser   RoleAttribute = "superuser"
	RoleAttributeLogin       RoleAttribute = "login"
	RoleAttributeCreateRole  RoleAttribute = "createRole"
	RoleAttributeCreateDB    RoleAttribute = "createDatabase"
	RoleAttributeReplication RoleAttribute = "replication"
)

type Role struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	CanLogin      bool     `json:"canLogin"`
	IsSuper       bool     `json:"isSuper"`
	CanCreateRole bool     `json:"canCreateRole"`
	CanCreateDB   bool     `json:"canCreateDb"`
	Replication   bool     `json:"replication"`
	MemberOf      []string `json:"memberOf"`
}

type RoleMembership struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleView struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	IsGroupRole bool             `json:"isGroup"`
	Membership  []RoleMembership `json:"membership"`
	Attributes  []RoleAttribute  `json:"attributes"`
	AccessLevel RoleAccessLevel  `json:"accessLevel"`
}

var RoleDescriptions = map[string]string{
	// Data access
	"pg_read_all_data":  "Can read all data",
	"pg_write_all_data": "Can modify all data",

	// Monitoring
	"pg_monitor":           "Can access monitoring stats",
	"pg_read_all_settings": "Can read all configuration settings",
	"pg_read_all_stats":    "Can read all statistics",
	"pg_stat_scan_tables":  "Can scan tables for statistics",

	// Server access
	"pg_read_server_files":      "Can read server files",
	"pg_write_server_files":     "Can write server files",
	"pg_execute_server_program": "Can execute server programs",

	// Admin actions
	"pg_signal_backend": "Can terminate backend processes",
	"pg_checkpoint":     "Can trigger checkpoints",

	// Replication
	"pg_create_subscription": "Can create logical replication subscriptions",
}

type AccessRule struct {
	Level RoleAccessLevel
	Check func(role Role) bool
}

var AccessRules = []AccessRule{
	// Admin
	{
		Level: RoleAccessLevelAdmin,
		Check: func(role Role) bool {
			return role.IsSuper ||
				slices.Contains(role.MemberOf, "pg_execute_server_program") ||
				slices.Contains(role.MemberOf, "pg_write_server_files") ||
				slices.Contains(role.MemberOf, "pg_read_server_files")
		},
	},

	// Elevated
	{
		Level: RoleAccessLevelElevated,
		Check: func(role Role) bool {
			return role.CanCreateRole || slices.Contains(role.MemberOf, "pg_write_all_data")
		},
	},

	// Standard
	{
		Level: RoleAccessLevelStandard,
		Check: func(role Role) bool {
			return role.CanCreateDB ||
				slices.Contains(role.MemberOf, "pg_read_all_data") ||
				slices.Contains(role.MemberOf, "pg_monitor")
		},
	},
}
