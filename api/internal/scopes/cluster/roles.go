package cluster

import (
	"context"
	"dashboard/api/internal/scopes/cluster/entities"
	"dashboard/api/internal/utils"
	"errors"
)

func (c *Cluster) Roles(ctx context.Context) ([]entities.RoleView, error) {

	roles, err := c.storage.Roles(ctx)
	if err != nil {
		c.logger.ErrorContext(ctx, "get roles", "error", err)
		return nil, errors.New("failed to get cluster roles")
	}

	views := make([]entities.RoleView, 0, len(roles))

	for _, role := range roles {
		views = append(views, computeRoleView(role))
	}

	c.logger.DebugContext(ctx, "got cluster roles", "count", len(roles))

	return views, nil
}

func buildMembership(role entities.Role) []entities.RoleMembership {
	var res []entities.RoleMembership

	for _, r := range role.MemberOf {
		desc, ok := entities.RoleDescriptions[r]
		if !ok {
			desc = r
		}

		res = append(res, entities.RoleMembership{
			Name:        r,
			Description: desc,
		})
	}

	return res
}

func computeRoleView(role entities.Role) entities.RoleView {
	rv := entities.RoleView{
		ID:          utils.IntToString(role.ID),
		Name:        role.Name,
		Membership:  buildMembership(role),
		IsGroupRole: !role.CanLogin,
	}

	var attributes []entities.RoleAttribute

	if role.IsSuper {
		attributes = append(attributes, entities.RoleAttributeSuperuser)
	}
	if role.CanLogin {
		attributes = append(attributes, entities.RoleAttributeLogin)
	}
	if role.CanCreateRole {
		attributes = append(attributes, entities.RoleAttributeCreateRole)
	}
	if role.CanCreateDB {
		attributes = append(attributes, entities.RoleAttributeCreateDB)
	}
	if role.Replication {
		attributes = append(attributes, entities.RoleAttributeReplication)
	}

	rv.Attributes = attributes

	rv.AccessLevel = entities.RoleAccessLevelLimited

	for _, rule := range entities.AccessRules {
		if rule.Check(role) {
			rv.AccessLevel = rule.Level
			break
		}
	}

	return rv
}
