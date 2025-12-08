package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// Helper functions for converting SQLC query rows to Challenge models

func convertUpdateChallengeTimestampsRowToChallenge(row sqlc.UpdateChallengeTimestampsRow) *model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
}

func convertUpdateChallengeRequirementsRowToChallenge(row sqlc.UpdateChallengeRequirementsRow) *model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
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
