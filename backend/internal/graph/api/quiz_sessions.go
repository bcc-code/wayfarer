package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/pubsub"
	"github.com/bcc-media/wayfarer/internal/services"
	"go.opentelemetry.io/otel/trace"
)

// quizSessionGrantAccessContext holds the loaded data needed for granting quiz session access
type quizSessionGrantAccessContext struct {
	UserID    string
	Session   *sqlc.QuizSession
	Quiz      *model.Quiz
	ProjectID string
}

// loadQuizSessionForGrantAccess loads session, quiz, and verifies authorization for granting access.
// Returns the context needed for grant operations or an error if unauthorized.
func loadQuizSessionForGrantAccess(
	ctx context.Context,
	span trace.Span,
	ldrs *loaders.Loaders,
	roleService *services.RoleService,
	sessionID string,
) (*quizSessionGrantAccessContext, error) {
	// Get user from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user not authenticated")
	}

	// Load session using loader
	sessionThunk := ldrs.QuizSessionByIDLoader.Load(ctx, sessionID)
	session, err := sessionThunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Load quiz to check authorization
	quizThunk := ldrs.QuizByIDLoader.Load(ctx, session.QuizID)
	quiz, err := quizThunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load quiz: %w", err)
	}
	if quiz == nil {
		return nil, fmt.Errorf("quiz not found")
	}

	// Check if user can grant access (project manager OR session creator)
	if session.CreatedBy != userID && !roleService.CanManageProject(ctx, userID, quiz.ProjectID) {
		return nil, fmt.Errorf("unauthorized to grant session access")
	}

	return &quizSessionGrantAccessContext{
		UserID:    userID,
		Session:   session,
		Quiz:      quiz,
		ProjectID: quiz.ProjectID,
	}, nil
}

// buildGrantQuizSessionAccessParams creates the params struct from input
func buildGrantQuizSessionAccessParams(
	input model.GrantQuizSessionAccessInput,
	projectID string,
	grantedBy string,
) pubsub.BulkGrantQuizSessionAccessParams {
	return pubsub.BulkGrantQuizSessionAccessParams{
		SessionID:       input.SessionID,
		UserIDs:         input.UserIds,
		TeamIDs:         input.TeamIds,
		SuperTeamIDs:    input.SuperTeamIds,
		ChurchIDs:       input.ChurchIds,
		AllProjectUsers: input.AllProjectUsers != nil && *input.AllProjectUsers,
		ProjectID:       projectID,
		GrantedBy:       grantedBy,
	}
}

// estimateGrantAccessTotalCount estimates the total number of users that will be granted access.
// Uses loaders for efficient batched queries.
func estimateGrantAccessTotalCount(
	ctx context.Context,
	ldrs *loaders.Loaders,
	input model.GrantQuizSessionAccessInput,
	projectID string,
) (int, error) {
	totalCount := len(input.UserIds)

	// Team members
	if len(input.TeamIds) > 0 {
		for _, teamID := range input.TeamIds {
			thunk := ldrs.UserIDsByTeamLoader.Load(ctx, teamID)
			userIDs, err := thunk()
			if err != nil {
				return 0, fmt.Errorf("failed to count team members: %w", err)
			}
			totalCount += len(userIDs)
		}
	}

	// Super team members
	if len(input.SuperTeamIds) > 0 {
		for _, superTeamID := range input.SuperTeamIds {
			thunk := ldrs.UserIDsBySuperTeamLoader.Load(ctx, superTeamID)
			userIDs, err := thunk()
			if err != nil {
				return 0, fmt.Errorf("failed to count super team members: %w", err)
			}
			totalCount += len(userIDs)
		}
	}

	// Church members
	if len(input.ChurchIds) > 0 {
		for _, churchID := range input.ChurchIds {
			key := loaders.ChurchProjectKey{ChurchID: churchID, ProjectID: projectID}
			thunk := ldrs.UserIDsByChurchInProjectLoader.Load(ctx, key)
			userIDs, err := thunk()
			if err != nil {
				return 0, fmt.Errorf("failed to count church members: %w", err)
			}
			totalCount += len(userIDs)
		}
	}

	// All project users
	if input.AllProjectUsers != nil && *input.AllProjectUsers {
		thunk := ldrs.UserIDsInProjectLoader.Load(ctx, projectID)
		userIDs, err := thunk()
		if err != nil {
			return 0, fmt.Errorf("failed to count project users: %w", err)
		}
		totalCount += len(userIDs)
	}

	return totalCount, nil
}
