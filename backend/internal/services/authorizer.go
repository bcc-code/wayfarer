package services

import (
	"context"
	"fmt"
)

// Action identifies an authorization decision the caller wants to make. It is the
// single vocabulary that resolvers converge on via RoleService.Can, replacing the
// scattered CanX method calls with one (actor, action, resource) surface.
type Action string

const (
	// ActionAccessUser: may the actor read the target user's full record?
	ActionAccessUser Action = "access_user"
	// ActionManageProject: may the actor manage the given project?
	ActionManageProject Action = "manage_project"
	// ActionManageChurch: may the actor manage the given church?
	ActionManageChurch Action = "manage_church"
	// ActionManageTeam: may the actor manage the given team?
	ActionManageTeam Action = "manage_team"
	// ActionAssignRole: may the actor assign/revoke the given (scoped) role?
	ActionAssignRole Action = "assign_role"
)

// Resource carries the target of an authorization decision. Only the fields
// relevant to the chosen Action are read (documented per-action in Can).
type Resource struct {
	// Manage* actions
	ProjectID string
	ChurchID  string
	TeamID    string

	// ActionAccessUser
	TargetUserID       string
	TargetUserChurchID string

	// ActionAssignRole (scope pointers mirror AssignRole/CanAssignRole)
	Role         RoleType
	ChurchScope  *string
	ProjectScope *string
	TeamScope    *string
}

// Can is the centralized authorization entry point. It dispatches to the existing
// scope-aware RoleService checks, giving resolvers one consistent surface to call
// while keeping each underlying method independently testable. Behavior is identical
// to calling the corresponding method directly.
func (s *RoleService) Can(ctx context.Context, actorID string, action Action, res Resource) (bool, error) {
	switch action {
	case ActionAccessUser:
		return s.CanAccessUser(ctx, actorID, res.TargetUserID, res.TargetUserChurchID), nil
	case ActionManageProject:
		return s.CanManageProject(ctx, actorID, res.ProjectID), nil
	case ActionManageChurch:
		return s.CanManageChurch(ctx, actorID, res.ChurchID), nil
	case ActionManageTeam:
		return s.CanManageTeam(ctx, actorID, res.TeamID), nil
	case ActionAssignRole:
		return s.CanAssignRole(ctx, actorID, res.Role, res.ChurchScope, res.ProjectScope, res.TeamScope)
	default:
		return false, fmt.Errorf("unknown authorization action: %q", action)
	}
}
