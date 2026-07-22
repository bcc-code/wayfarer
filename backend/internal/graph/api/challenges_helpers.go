package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
)

// actorIsM2M reports whether the request is made by an M2M service account, based
// on the roles carried in the JWT (M2M accounts do not exist in the database).
func actorIsM2M(ctx context.Context) bool {
	for _, role := range middleware.GetUserRoles(ctx) {
		if role == "m2m" {
			return true
		}
	}
	return false
}

// authorizeChallengeEnrollment verifies that the actor may bulk-enroll the given
// target into a challenge belonging to projectID.
//
//   - admin / superadmin / m2m: unrestricted.
//   - church_admin (the only other role @requireRole admits on the bulk-enroll
//     mutations): may ONLY enroll a single church within the challenge's own project,
//     and only a church they administer. Every other target shape (userIds, teamIds,
//     superTeamIds, allProjectMembers) is rejected because it could span churches.
func (r *mutationResolver) authorizeChallengeEnrollment(ctx context.Context, actorID string, isM2M bool, target model.EnrollmentTargetInput, projectID string) error {
	if isM2M || r.RoleService.IsAdmin(ctx, actorID) {
		return nil
	}

	if target.ChurchInProject == nil ||
		len(target.UserIds) > 0 ||
		len(target.TeamIds) > 0 ||
		len(target.SuperTeamIds) > 0 ||
		target.AllProjectMembers != nil {
		return fmt.Errorf("permission denied: church admins may only bulk-enroll a church within the challenge's project")
	}

	if target.ChurchInProject.ProjectID != projectID {
		return fmt.Errorf("permission denied: target project does not match the challenge's project")
	}

	if !r.RoleService.CanManageChurch(ctx, actorID, target.ChurchInProject.ChurchID) {
		return fmt.Errorf("permission denied: not a church admin for the target church")
	}

	return nil
}

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
