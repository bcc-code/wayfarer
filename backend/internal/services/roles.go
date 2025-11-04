package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
)

// RoleQuerier defines the database operations needed for role management
type RoleQuerier interface {
	GetUserRoles(ctx context.Context, userID string) ([]*sqlc.UserRole, error)
	AssignRole(ctx context.Context, arg sqlc.AssignRoleParams) (*sqlc.UserRole, error)
	RevokeRole(ctx context.Context, arg sqlc.RevokeRoleParams) error
	HasRole(ctx context.Context, arg sqlc.HasRoleParams) (bool, error)
	HasRoleInChurch(ctx context.Context, arg sqlc.HasRoleInChurchParams) (bool, error)
	HasRoleInProject(ctx context.Context, arg sqlc.HasRoleInProjectParams) (bool, error)
	HasRoleInTeam(ctx context.Context, arg sqlc.HasRoleInTeamParams) (bool, error)
}

// RoleType represents the available roles in the system
type RoleType string

const (
	RoleSuperAdmin   RoleType = "SUPERADMIN"
	RoleAdmin        RoleType = "ADMIN"
	RoleChurchAdmin  RoleType = "CHURCH_ADMIN"
	RoleProjectAdmin RoleType = "PROJECT_ADMIN"
	RoleTeamLead     RoleType = "TEAM_LEAD"
	RoleUser         RoleType = "USER"
	RoleM2M          RoleType = "M2M"
)

// ScopeType represents the scope of a role
type ScopeType string

const (
	ScopeChurch  ScopeType = "CHURCH"
	ScopeProject ScopeType = "PROJECT"
	ScopeTeam    ScopeType = "TEAM"
)

// RoleService provides role and permission management functionality
type RoleService struct {
	queries RoleQuerier
}

// NewRoleService creates a new role service
func NewRoleService(queries RoleQuerier) *RoleService {
	return &RoleService{
		queries: queries,
	}
}

// LoadUserRoles loads all roles for a user
func (s *RoleService) LoadUserRoles(ctx context.Context, userID string) ([]*sqlc.UserRole, error) {
	return s.queries.GetUserRoles(ctx, userID)
}

// HasRole checks if a user has a specific global role
func (s *RoleService) HasRole(ctx context.Context, userID string, role RoleType) bool {
	hasRole, err := s.queries.HasRole(ctx, sqlc.HasRoleParams{
		UserID: userID,
		Role:   string(role),
	})

	if err != nil {
		slog.Error("error when checking roles", userID, err)
		return false
	}

	return hasRole
}

// HasRoleInChurch checks if a user has a specific role in a church
func (s *RoleService) HasRoleInChurch(ctx context.Context, userID string, role RoleType, churchID string) bool {
	hasRole, err := s.queries.HasRoleInChurch(ctx, sqlc.HasRoleInChurchParams{
		UserID:   userID,
		Role:     string(role),
		ChurchID: &churchID,
	})

	if err != nil {
		slog.Error("error when checking roles", userID, err)
		return false
	}

	return hasRole
}

// HasRoleInProject checks if a user has a specific role in a project
func (s *RoleService) HasRoleInProject(ctx context.Context, userID string, role RoleType, projectID string) bool {
	hasRole, err := s.queries.HasRoleInProject(ctx, sqlc.HasRoleInProjectParams{
		UserID:    userID,
		Role:      string(role),
		ProjectID: &projectID,
	})

	if err != nil {
		slog.Error("error when checking roles", userID, err)
		return false
	}

	return hasRole
}

// HasRoleInTeam checks if a user has a specific role in a team
func (s *RoleService) HasRoleInTeam(ctx context.Context, userID string, role RoleType, teamID string) bool {
	hasRole, err := s.queries.HasRoleInTeam(ctx, sqlc.HasRoleInTeamParams{
		UserID: userID,
		Role:   string(role),
		TeamID: &teamID,
	})

	if err != nil {
		slog.Error("error when checking roles", userID, err)
		return false
	}

	return hasRole
}

// AssignRole assigns a role to a user with proper authorization checks
func (s *RoleService) AssignRole(ctx context.Context, assignerID, userID string, role RoleType, churchID, projectID, teamID *string) (*sqlc.UserRole, error) {
	// Check if assigner has permission to assign this role
	canAssign, err := s.CanAssignRole(ctx, assignerID, role, churchID, projectID, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to check assignment permission: %w", err)
	}
	if !canAssign {
		return nil, fmt.Errorf("user %s does not have permission to assign role %s", assignerID, role)
	}

	// Generate new role ID
	roleID := ulid.NewUserRoleID()

	// Assign the role
	return s.queries.AssignRole(ctx, sqlc.AssignRoleParams{
		ID:         roleID,
		UserID:     userID,
		Role:       string(role),
		ChurchID:   churchID,
		ProjectID:  projectID,
		TeamID:     teamID,
		AssignedBy: assignerID,
		AssignedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
}

// RevokeRole revokes a role from a user with proper authorization checks
func (s *RoleService) RevokeRole(ctx context.Context, revokerID, userID string, role RoleType, churchID, projectID, teamID *string) error {
	// Check if revoker has permission to revoke this role
	canRevoke, err := s.CanAssignRole(ctx, revokerID, role, churchID, projectID, teamID)
	if err != nil {
		return fmt.Errorf("failed to check revocation permission: %w", err)
	}
	if !canRevoke {
		return fmt.Errorf("user %s does not have permission to revoke role %s", revokerID, role)
	}

	// Revoke the role
	return s.queries.RevokeRole(ctx, sqlc.RevokeRoleParams{
		UserID:    userID,
		Role:      string(role),
		ChurchID:  churchID,
		ProjectID: projectID,
		TeamID:    teamID,
	})
}

// CanAssignRole checks if a user has permission to assign a specific role
// Permission hierarchy:
// - SUPERADMIN can assign all roles
// - ADMIN can assign CHURCH_ADMIN, PROJECT_ADMIN, TEAM_LEAD
// - CHURCH_ADMIN can assign TEAM_LEAD (only within their church)
func (s *RoleService) CanAssignRole(ctx context.Context, assignerID string, roleToAssign RoleType, churchID, projectID, teamID *string) (bool, error) {
	// Check if assigner is a superadmin
	if s.IsSuperAdmin(ctx, assignerID) {
		return true, nil // Superadmins can assign any role
	}

	// Check if assigner is an admin
	if s.HasRole(ctx, assignerID, RoleAdmin) {
		// Admins can assign CHURCH_ADMIN, PROJECT_ADMIN, TEAM_LEAD
		return roleToAssign == RoleChurchAdmin || roleToAssign == RoleProjectAdmin || roleToAssign == RoleTeamLead, nil
	}

	// Check if assigner is a church admin
	if roleToAssign == RoleTeamLead && teamID != nil {
		// Church admins can only assign team lead roles
		// We need to check if the team belongs to a church where the assigner is a church admin
		// For now, we'll check if the assigner is a church admin in any church
		// A more sophisticated check would verify the team's church relationship
		roles, err := s.LoadUserRoles(ctx, assignerID)
		if err != nil {
			return false, err
		}

		for _, role := range roles {
			if role.Role == string(RoleChurchAdmin) && role.ChurchID != nil {
				// TODO: Verify that the team belongs to this church
				return true, nil
			}
		}
	}

	return false, nil
}

// IsSuperAdmin checks if a user is a superadmin
func (s *RoleService) IsSuperAdmin(ctx context.Context, userID string) bool {
	return s.HasRole(ctx, userID, RoleSuperAdmin)
}

// IsAdmin checks if a user is an admin or superadmin
func (s *RoleService) IsAdmin(ctx context.Context, userID string) bool {
	if s.HasRole(ctx, userID, RoleSuperAdmin) {
		return true
	}

	return s.HasRole(ctx, userID, RoleAdmin)
}

// CanManageProject checks if a user can manage a specific project
// Returns true if user is:
// - superadmin
// - admin
// - project admin for this project
func (s *RoleService) CanManageProject(ctx context.Context, userID, projectID string) bool {
	// Check if user is superadmin or admin
	if s.IsAdmin(ctx, userID) {
		return true
	}

	// Check if user is project admin for this specific project
	return s.HasRoleInProject(ctx, userID, RoleProjectAdmin, projectID)
}

// CanManageChurch checks if a user can manage a specific church
// Returns true if user is:
// - superadmin
// - admin
// - church admin for this church
func (s *RoleService) CanManageChurch(ctx context.Context, userID, churchID string) bool {
	// Check if user is superadmin or admin
	if s.IsAdmin(ctx, userID) {
		return true
	}

	// Check if user is church admin for this specific church
	return s.HasRoleInChurch(ctx, userID, RoleChurchAdmin, churchID)
}

// CanManageTeam checks if a user can manage a specific team
// Returns true if user is:
// - superadmin
// - admin
// - church admin for the team's church
// - project admin for the team's project
// - team lead for this team
func (s *RoleService) CanManageTeam(ctx context.Context, userID, teamID string) bool {
	// Check if user is superadmin or admin
	if s.IsAdmin(ctx, userID) {
		return true
	}

	// Check if user is team lead for this specific team
	if s.HasRoleInTeam(ctx, userID, RoleTeamLead, teamID) {
		return true
	}

	// TODO: Check if user is church admin or project admin for the team's church/project
	// This requires additional queries to get the team's church and project
	// For now, return false - will be implemented when needed

	return false
}

// GetPrimaryRole returns the highest priority role for a user
// Priority order: SUPERADMIN > ADMIN > CHURCH_ADMIN > PROJECT_ADMIN > TEAM_LEAD > USER > M2M
func (s *RoleService) GetPrimaryRole(ctx context.Context, userID string) (RoleType, error) {
	roles, err := s.LoadUserRoles(ctx, userID)
	if err != nil {
		return "", err
	}

	if len(roles) == 0 {
		return RoleUser, nil // Default to USER role
	}

	// Check in priority order
	rolePriority := []RoleType{
		RoleSuperAdmin,
		RoleAdmin,
		RoleChurchAdmin,
		RoleProjectAdmin,
		RoleTeamLead,
		RoleUser,
		RoleM2M,
	}

	for _, priority := range rolePriority {
		for _, userRole := range roles {
			if RoleType(userRole.Role) == priority {
				return priority, nil
			}
		}
	}

	return RoleUser, nil
}
