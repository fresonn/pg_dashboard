package roles

import (
	"context"
	"dashboard/api/internal/helper"
	roleModels "dashboard/api/internal/model/role"

	"errors"
)

func (s *Service) Roles(ctx context.Context) ([]roleModels.RoleView, error) {

	roles, err := s.storage.Roles(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "get roles", "error", err)
		return nil, errors.New("failed to get cluster roles")
	}

	views := make([]roleModels.RoleView, 0, len(roles))

	for _, role := range roles {
		views = append(views, computeRoleView(role))
	}

	s.logger.DebugContext(ctx, "got cluster roles", "count", len(roles))

	return views, nil
}

func buildMembership(role roleModels.Role) []roleModels.RoleMembership {
	var res []roleModels.RoleMembership

	for _, r := range role.MemberOf {
		desc, ok := roleModels.RoleDescriptions[r]
		if !ok {
			desc = r
		}

		res = append(res, roleModels.RoleMembership{
			Name:        r,
			Description: desc,
		})
	}

	return res
}

func computeRoleView(role roleModels.Role) roleModels.RoleView {
	rv := roleModels.RoleView{
		ID:          helper.IntToString(role.ID),
		Name:        role.Name,
		Membership:  buildMembership(role),
		IsGroupRole: !role.CanLogin,
	}

	var attributes []roleModels.RoleAttribute

	if role.IsSuper {
		attributes = append(attributes, roleModels.RoleAttributeSuperuser)
	}
	if role.CanLogin {
		attributes = append(attributes, roleModels.RoleAttributeLogin)
	}
	if role.CanCreateRole {
		attributes = append(attributes, roleModels.RoleAttributeCreateRole)
	}
	if role.CanCreateDB {
		attributes = append(attributes, roleModels.RoleAttributeCreateDB)
	}
	if role.Replication {
		attributes = append(attributes, roleModels.RoleAttributeReplication)
	}

	rv.Attributes = attributes

	rv.AccessLevel = roleModels.RoleAccessLevelLimited

	for _, rule := range roleModels.AccessRules {
		if rule.Check(role) {
			rv.AccessLevel = rule.Level
			break
		}
	}

	return rv
}
