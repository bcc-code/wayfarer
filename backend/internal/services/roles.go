package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/bcc-media/wayfarer/internal/cache"
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
	HasTeamMemberFromChurch(ctx context.Context, arg sqlc.HasTeamMemberFromChurchParams) (bool, error)
	GetTeamProjectID(ctx context.Context, teamid string) (string, error)
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
	cache   *cache.CacheWithRegistry
}

// NewRoleService creates a new role service
func NewRoleService(queries RoleQuerier, c *cache.CacheWithRegistry) *RoleService {
	return &RoleService{
		queries: queries,
		cache:   c,
	}
}

// LoadUserRoles loads all roles for a user
func (s *RoleService) LoadUserRoles(ctx context.Context, userID string) ([]*sqlc.UserRole, error) {
	return s.queries.GetUserRoles(ctx, userID)
}

// HasRole checks if a user has a specific global role
func (s *RoleService) HasRole(ctx context.Context, userID string, role RoleType) bool {
	// Check cache first
	cacheKey := cache.HasRoleKey(userID, string(role))
	if cached, ok := s.cache.Get(cacheKey); ok {
		if hasRole, ok := cached.(bool); ok {
			return hasRole
		}
	}

	// Query database
	hasRole, err := s.queries.HasRole(ctx, sqlc.HasRoleParams{
		UserID: userID,
		Role:   string(role),
	})

	if err != nil {
		slog.Error("error when checking roles", userID, err)
		return false
	}

	// Store in cache
	s.cache.Set(cacheKey, hasRole)

	return hasRole
}

// HasRoleInChurch checks if a user has a specific role in a church
func (s *RoleService) HasRoleInChurch(ctx context.Context, userID string, role RoleType, churchID string) bool {
	// Check cache first
	cacheKey := cache.HasRoleInChurchKey(userID, string(role), churchID)
	if cached, ok := s.cache.Get(cacheKey); ok {
		if hasRole, ok := cached.(bool); ok {
			return hasRole
		}
	}

	// Query database
	hasRole, err := s.queries.HasRoleInChurch(ctx, sqlc.HasRoleInChurchParams{
		UserID:   userID,
		Role:     string(role),
		ChurchID: &churchID,
	})

	if err != nil {
		slog.Error("error when checking roles", userID, err)
		return false
	}

	// Store in cache
	s.cache.Set(cacheKey, hasRole)

	return hasRole
}

// HasRoleInProject checks if a user has a specific role in a project
func (s *RoleService) HasRoleInProject(ctx context.Context, userID string, role RoleType, projectID string) bool {
	// Check cache first
	cacheKey := cache.HasRoleInProjectKey(userID, string(role), projectID)
	if cached, ok := s.cache.Get(cacheKey); ok {
		if hasRole, ok := cached.(bool); ok {
			return hasRole
		}
	}

	// Query database
	hasRole, err := s.queries.HasRoleInProject(ctx, sqlc.HasRoleInProjectParams{
		UserID:    userID,
		Role:      string(role),
		ProjectID: &projectID,
	})

	if err != nil {
		slog.Error("error when checking roles", userID, err)
		return false
	}

	// Store in cache
	s.cache.Set(cacheKey, hasRole)

	return hasRole
}

// HasRoleInTeam checks if a user has a specific role in a team
func (s *RoleService) HasRoleInTeam(ctx context.Context, userID string, role RoleType, teamID string) bool {
	// Check cache first
	cacheKey := cache.HasRoleInTeamKey(userID, string(role), teamID)
	if cached, ok := s.cache.Get(cacheKey); ok {
		if hasRole, ok := cached.(bool); ok {
			return hasRole
		}
	}

	// Query database
	hasRole, err := s.queries.HasRoleInTeam(ctx, sqlc.HasRoleInTeamParams{
		UserID: userID,
		Role:   string(role),
		TeamID: &teamID,
	})

	if err != nil {
		slog.Error("error when checking roles", userID, err)
		return false
	}

	// Store in cache
	s.cache.Set(cacheKey, hasRole)

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
// - CHURCH_ADMIN can assign CHURCH_ADMIN (only within their own church) and TEAM_LEAD
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
	// Church admins can assign CHURCH_ADMIN role to users within their own church
	if roleToAssign == RoleChurchAdmin && churchID != nil {
		if s.HasRoleInChurch(ctx, assignerID, RoleChurchAdmin, *churchID) {
			return true, nil
		}
	}

	// Church admins can assign team lead roles
	if roleToAssign == RoleTeamLead && teamID != nil {
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
// - project admin for the team's project
// - church admin (if any team member is from their church)
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

	// Get the team's project ID
	projectID, err := s.queries.GetTeamProjectID(ctx, teamID)
	if err != nil {
		slog.Error("error getting team project ID", "teamID", teamID, "error", err)
		return false
	}

	// Check if user is project admin for the team's project
	if s.HasRoleInProject(ctx, userID, RoleProjectAdmin, projectID) {
		return true
	}

	// Check if user is church admin and team has member from their church
	// Get all user roles to find church admin roles
	roles, err := s.LoadUserRoles(ctx, userID)
	if err != nil {
		slog.Error("error loading user roles", "userID", userID, "error", err)
		return false
	}

	for _, role := range roles {
		if role.Role == string(RoleChurchAdmin) && role.ChurchID != nil {
			// Check if team has any member from this church
			hasMember, err := s.queries.HasTeamMemberFromChurch(ctx, sqlc.HasTeamMemberFromChurchParams{
				Teamid:   teamID,
				Churchid: *role.ChurchID,
			})
			if err != nil {
				slog.Error("error checking team member from church", "teamID", teamID, "churchID", *role.ChurchID, "error", err)
				continue
			}
			if hasMember {
				return true
			}
		}
	}

	return false
}

// CanDeleteTeam checks if a user can delete a specific team
// Only admins and superadmins can delete teams (not team leads)
// Returns true if user is:
// - superadmin
// - admin
func (s *RoleService) CanDeleteTeam(ctx context.Context, userID string) bool {
	return s.IsAdmin(ctx, userID)
}

// IsProjectAdmin checks if a user has project admin role for any project
func (s *RoleService) IsProjectAdmin(ctx context.Context, userID string) bool {
	return s.HasRole(ctx, userID, RoleProjectAdmin)
}

// CanAccessUser checks if currentUserID can access targetUserID's User object.
//
// Returns true if any of these conditions are met:
//   - currentUserID is the target user (self-access)
//   - currentUserID has M2M role (for system integration)
//   - currentUserID has admin or superadmin role
//   - currentUserID has project admin role (for any project)
//   - currentUserID is church admin for targetChurchID
//
// Returns false otherwise.
func (s *RoleService) CanAccessUser(ctx context.Context, currentUserID, targetUserID, targetChurchID string) bool {
	// Self-access always allowed
	if currentUserID == targetUserID {
		return true
	}

	// M2M service accounts can access any user
	if s.HasRole(ctx, currentUserID, RoleM2M) {
		return true
	}

	// Admins and superadmins can access any user
	if s.IsAdmin(ctx, currentUserID) {
		return true
	}

	// Project admins can access any user
	if s.IsProjectAdmin(ctx, currentUserID) {
		return true
	}

	// Church admins can access users in their church
	if s.CanManageChurch(ctx, currentUserID, targetChurchID) {
		return true
	}

	return false
}
