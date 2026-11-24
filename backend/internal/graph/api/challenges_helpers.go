package api

import (
	"context"
	"fmt"
	"slices"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
)

// resolveEnrollmentTarget converts an EnrollmentTargetInput to a list of user IDs
func (r *mutationResolver) resolveEnrollmentTarget(ctx context.Context, target model.EnrollmentTargetInput) ([]string, error) {
	// Validate exactly one target type is specified
	setCount := 0
	if len(target.UserIds) > 0 {
		setCount++
	}
	if target.ChurchInProject != nil {
		setCount++
	}
	if len(target.TeamIds) > 0 {
		setCount++
	}
	if len(target.SuperTeamIds) > 0 {
		setCount++
	}
	if target.AllProjectMembers != nil {
		setCount++
	}

	if setCount == 0 {
		return nil, fmt.Errorf("must specify at least one target type")
	}
	if setCount > 1 {
		return nil, fmt.Errorf("must specify exactly one target type")
	}

	// Resolve based on target type
	if len(target.UserIds) > 0 {
		return target.UserIds, nil
	}

	if target.ChurchInProject != nil {
		userIDs, err := r.DB.Queries.GetUserIDsInChurchAndProject(ctx, sqlc.GetUserIDsInChurchAndProjectParams{
			Churchid:  target.ChurchInProject.ChurchID,
			Projectid: target.ChurchInProject.ProjectID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get users in church and project: %w", err)
		}
		return userIDs, nil
	}

	if len(target.TeamIds) > 0 {
		userIDs, err := r.DB.Queries.GetUserIDsInTeams(ctx, target.TeamIds)
		if err != nil {
			return nil, fmt.Errorf("failed to get team members: %w", err)
		}
		return userIDs, nil
	}

	if len(target.SuperTeamIds) > 0 {
		userIDs, err := r.DB.Queries.GetUserIDsInSuperTeams(ctx, target.SuperTeamIds)
		if err != nil {
			return nil, fmt.Errorf("failed to get super team members: %w", err)
		}
		return userIDs, nil
	}

	if target.AllProjectMembers != nil {
		userIDs, err := r.DB.Queries.GetUserIDsInProject(ctx, *target.AllProjectMembers)
		if err != nil {
			return nil, fmt.Errorf("failed to get project members: %w", err)
		}
		return userIDs, nil
	}

	return nil, fmt.Errorf("no valid target specified")
}

// bulkEnrollmentRoles is the role lookups authorizeBulkEnrollment needs.
type bulkEnrollmentRoles interface {
	CanManageProject(ctx context.Context, userID, projectID string) bool
	LoadUserRoles(ctx context.Context, userID string) ([]*sqlc.UserRole, error)
}

// churchAdminScopeQuerier is the query authorizeBulkEnrollment needs.
type churchAdminScopeQuerier interface {
	GetUsersOutsideChurchAdminScope(ctx context.Context, arg sqlc.GetUsersOutsideChurchAdminScopeParams) ([]string, error)
}

// authorizeBulkEnrollment checks that callerID may enroll userIDs in a
// challenge of projectID. Admins, project admins of the project and m2m may
// enroll anyone. A church admin may enroll only members of the project who
// are in a church they administer, or on a team in the project that has a
// member from one of those churches (the kickoff enrolls the leads of such
// teams). The request is rejected as a whole if any user is outside that.
func authorizeBulkEnrollment(ctx context.Context, roles bulkEnrollmentRoles, q churchAdminScopeQuerier, callerID, projectID string, userIDs []string) error {
	// m2m comes from the token, as in @requireRole: M2M callers have no
	// user_roles row to look up.
	if slices.Contains(middleware.GetUserRoles(ctx), "m2m") {
		return nil
	}
	if roles.CanManageProject(ctx, callerID, projectID) {
		return nil
	}

	userRoles, err := roles.LoadUserRoles(ctx, callerID)
	if err != nil {
		return fmt.Errorf("failed to load roles: %w", err)
	}
	var churchIDs []string
	for _, role := range userRoles {
		if role.Role == string(services.RoleChurchAdmin) && role.ChurchID != nil {
			churchIDs = append(churchIDs, *role.ChurchID)
		}
	}
	if len(churchIDs) == 0 {
		return fmt.Errorf("not authorized to enroll users in this challenge")
	}

	outside, err := q.GetUsersOutsideChurchAdminScope(ctx, sqlc.GetUsersOutsideChurchAdminScopeParams{
		Userids:   userIDs,
		Projectid: projectID,
		Churchids: churchIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to check enrollment targets: %w", err)
	}
	if len(outside) > 0 {
		return fmt.Errorf("not allowed to enroll %d user(s) outside your church in this challenge's project", len(outside))
	}
	return nil
}
