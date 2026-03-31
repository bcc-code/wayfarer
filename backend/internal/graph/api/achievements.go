package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
)

// BulkAwardTarget holds the resolved target for a bulk award achievement operation
type BulkAwardTarget struct {
	Achievement model.Achievement
	ProjectID   string
	EventID     *string
	UserIDs     []string
}

// resolveBulkAwardTarget resolves the target users and achievement for a bulk award operation.
// It validates input, loads the achievement, checks if it's awardable, and resolves user IDs
// from either direct IDs or team membership.
func resolveBulkAwardTarget(
	ctx context.Context,
	queries *sqlc.Queries,
	loadersInstance *loaders.Loaders,
	userIds []string,
	teamID *string,
	achievementID string,
	force bool,
) (*BulkAwardTarget, error) {
	// Validate input
	hasUserIds := len(userIds) > 0
	hasTeamId := teamID != nil && *teamID != ""
	if !hasUserIds && !hasTeamId {
		return nil, fmt.Errorf("at least one of userIds or teamId must be provided")
	}

	// Load achievement
	thunk := loadersInstance.AchievementByIDLoader.Load(ctx, achievementID)
	achievement, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load achievement: %w", err)
	}

	// Check if achievement is awardable based on awardable_from timestamp
	if err := isAchievementAwardable(getAchievementAwardableFrom(achievement)); err != nil {
		return nil, err
	}

	// Check if project is finished (unless force=true)
	if err := checkProjectFinished(ctx, loadersInstance, getAchievementProjectID(achievement), force); err != nil {
		return nil, err
	}

	// Resolve target user IDs
	userIDSet := make(map[string]bool)
	allUserIDs := make([]string, 0)

	// Add explicitly provided user IDs
	for _, uid := range userIds {
		if !userIDSet[uid] {
			userIDSet[uid] = true
			allUserIDs = append(allUserIDs, uid)
		}
	}

	// If teamID is provided, get team members
	if hasTeamId {
		teamUserIDs, err := queries.GetUserIDsInTeams(ctx, []string{*teamID})
		if err != nil {
			return nil, fmt.Errorf("failed to get team members: %w", err)
		}
		for _, uid := range teamUserIDs {
			if !userIDSet[uid] {
				userIDSet[uid] = true
				allUserIDs = append(allUserIDs, uid)
			}
		}
	}

	return &BulkAwardTarget{
		Achievement: achievement,
		ProjectID:   getAchievementProjectID(achievement),
		EventID:     getAchievementEventID(achievement),
		UserIDs:     allUserIDs,
	}, nil
}

// buildAchievementFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildAchievementFilterParamsCursor(filter *model.AchievementFilter, first *int, after *string, last *int, before *string) (sqlc.GetAchievementsFilteredCursorParams, error) {
	params := sqlc.GetAchievementsFilteredCursorParams{}

	// Apply filters if provided
	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.EventID != nil {
			params.Eventid = *filter.EventID
		}
	}

	// Handle cursor pagination
	isBackward := false
	var limit int

	if first != nil && last != nil {
		return params, fmt.Errorf("cannot specify both first and last")
	}

	if first != nil {
		limit = *first + 1 // Fetch one extra to determine hasNextPage
		isBackward = false
	} else if last != nil {
		limit = *last + 1 // Fetch one extra to determine hasPreviousPage
		isBackward = true
	} else {
		// Default page size
		limit = 11 // 10 items + 1 to check for next page
		isBackward = false
	}

	params.Querylimit = int32(limit)
	params.Isbackward = isBackward

	// Set cursors
	if after != nil && *after != "" {
		params.Aftercursor = *after
	}

	if before != nil && *before != "" {
		params.Beforecursor = *before
	}

	return params, nil
}

// buildCountAchievementsFilterParams converts GraphQL filter to database count query parameters
func buildCountAchievementsFilterParams(filter *model.AchievementFilter) sqlc.CountAchievementsFilteredParams {
	params := sqlc.CountAchievementsFilteredParams{}

	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.EventID != nil {
			params.Eventid = *filter.EventID
		}
	}

	return params
}

// buildAchievementCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildAchievementCacheKeyParams(filter *model.AchievementFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if len(filter.Ids) > 0 {
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
		}
		if filter.ProjectID != nil {
			params["projectid"] = *filter.ProjectID
		}
		if filter.EventID != nil {
			params["eventid"] = *filter.EventID
		}
	}

	// Add pagination parameters
	if first != nil {
		params["first"] = fmt.Sprintf("%d", *first)
	}
	if after != nil && *after != "" {
		params["after"] = *after
	}
	if last != nil {
		params["last"] = fmt.Sprintf("%d", *last)
	}
	if before != nil && *before != "" {
		params["before"] = *before
	}

	return params
}

// convertRowToAchievement converts a database row to the appropriate Achievement implementation
func convertRowToAchievement(row *sqlc.GetAchievementsFilteredCursorRow) (model.Achievement, error) {
	hidden := false
	if row.Hidden != nil {
		hidden = *row.Hidden
	}

	switch row.AchievementType {
	case "SIMPLE":
		return convertRowToSimpleAchievement(row, hidden), nil
	case "CONTENT":
		return convertRowToContentAchievement(row, hidden)
	case "STREAK":
		return convertRowToStreakAchievement(row, hidden)
	case "QUIZ":
		return convertRowToQuizAchievement(row, hidden)
	default:
		return nil, fmt.Errorf("unknown achievement type: %s", row.AchievementType)
	}
}

func convertRowToSimpleAchievement(row *sqlc.GetAchievementsFilteredCursorRow, hidden bool) model.Achievement {
	var awardableFrom *scalars.DateTime
	if row.AwardableFrom.Valid {
		awardableFrom = &scalars.DateTime{Time: row.AwardableFrom.Time}
	}

	return &model.SimpleAchievement{
		ID:                   row.ID,
		Name:                 row.Name,
		DescriptionPending:   row.DescriptionPending,
		DescriptionCompleted: row.DescriptionCompleted,
		NotificationText:     row.NotificationText,
		ImagePending:         row.ImagePending,
		ImageCompleted:       row.ImageCompleted,
		Points:               int(row.Points),
		Hidden:               hidden,
		AwardableFrom:        awardableFrom,
		ProjectID:            row.ProjectID,
		EventID:              row.EventID,
		ChallengeID:          row.ChallengeID,
	}
}

func convertRowToContentAchievement(row *sqlc.GetAchievementsFilteredCursorRow, hidden bool) (model.Achievement, error) {
	// Count content items from JSON if available
	totalItems := 0
	if row.ContentItems != nil {
		var itemsData []map[string]interface{}
		jsonBytes, err := json.Marshal(row.ContentItems)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal content items: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &itemsData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal content items: %w", err)
		}
		totalItems = len(itemsData)
	}

	var awardableFrom *scalars.DateTime
	if row.AwardableFrom.Valid {
		awardableFrom = &scalars.DateTime{Time: row.AwardableFrom.Time}
	}

	return &model.ContentAchievement{
		ID:                   row.ID,
		Name:                 row.Name,
		DescriptionPending:   row.DescriptionPending,
		DescriptionCompleted: row.DescriptionCompleted,
		NotificationText:     row.NotificationText,
		ImagePending:         row.ImagePending,
		ImageCompleted:       row.ImageCompleted,
		Points:               int(row.Points),
		Hidden:               hidden,
		AwardableFrom:        awardableFrom,
		ProjectID:            row.ProjectID,
		EventID:              row.EventID,
		ChallengeID:          row.ChallengeID,
		TotalItems:           totalItems,
		// Items, UserCompletedItems, NextItem, and CompletedItemCount will be populated by resolvers
	}, nil
}

func convertRowToStreakAchievement(row *sqlc.GetAchievementsFilteredCursorRow, hidden bool) (model.Achievement, error) {
	if row.StreakAchievementID == nil {
		return nil, fmt.Errorf("streak achievement missing streak data")
	}

	// Count streak items from JSON if available
	totalItems := 0
	if row.StreakItems != nil {
		var itemsData []map[string]interface{}
		jsonBytes, err := json.Marshal(row.StreakItems)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal streak items: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &itemsData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal streak items: %w", err)
		}
		totalItems = len(itemsData)
	}

	var awardableFrom *scalars.DateTime
	if row.AwardableFrom.Valid {
		awardableFrom = &scalars.DateTime{Time: row.AwardableFrom.Time}
	}

	return &model.StreakAchievement{
		ID:                   row.ID,
		Name:                 row.Name,
		DescriptionPending:   row.DescriptionPending,
		DescriptionCompleted: row.DescriptionCompleted,
		NotificationText:     row.NotificationText,
		ImagePending:         row.ImagePending,
		ImageCompleted:       row.ImageCompleted,
		Points:               int(row.Points),
		Hidden:               hidden,
		AwardableFrom:        awardableFrom,
		ProjectID:            row.ProjectID,
		EventID:              row.EventID,
		ChallengeID:          row.ChallengeID,
		TotalItems:           totalItems,
	}, nil
}

func convertRowToQuizAchievement(row *sqlc.GetAchievementsFilteredCursorRow, hidden bool) (model.Achievement, error) {
	var awardableFrom *scalars.DateTime
	if row.AwardableFrom.Valid {
		awardableFrom = &scalars.DateTime{Time: row.AwardableFrom.Time}
	}

	var minScorePercentage *int
	if row.MinScorePercentage != nil {
		v := int(*row.MinScorePercentage)
		minScorePercentage = &v
	}

	requireCompletion := false
	if row.RequireCompletion != nil {
		requireCompletion = *row.RequireCompletion
	}

	var quizID string
	if row.QuizID != nil {
		quizID = *row.QuizID
	}

	return &model.QuizAchievement{
		ID:                   row.ID,
		Name:                 row.Name,
		DescriptionPending:   row.DescriptionPending,
		DescriptionCompleted: row.DescriptionCompleted,
		NotificationText:     row.NotificationText,
		ImagePending:         row.ImagePending,
		ImageCompleted:       row.ImageCompleted,
		Points:               int(row.Points),
		Hidden:               hidden,
		AwardableFrom:        awardableFrom,
		ProjectID:            row.ProjectID,
		EventID:              row.EventID,
		ChallengeID:          row.ChallengeID,
		QuizID:               quizID,
		MinScorePercentage:   minScorePercentage,
		RequireCompletion:    requireCompletion,
	}, nil
}

// convertPublishedContentAchievementRow converts GetPublishedContentAchievementsByExternalContentRow to ContentAchievement model
func convertPublishedContentAchievementRow(row *sqlc.GetPublishedContentAchievementsByExternalContentRow) *model.ContentAchievement {
	// Count content items from JSON if available
	totalItems := 0
	if row.ContentItems != nil {
		var itemsData []map[string]interface{}
		jsonBytes, err := json.Marshal(row.ContentItems)
		if err == nil {
			if err := json.Unmarshal(jsonBytes, &itemsData); err == nil {
				totalItems = len(itemsData)
			}
		}
	}

	hidden := false
	if row.Hidden != nil {
		hidden = *row.Hidden
	}

	var awardableFrom *scalars.DateTime
	if row.AwardableFrom.Valid {
		awardableFrom = &scalars.DateTime{Time: row.AwardableFrom.Time}
	}

	return &model.ContentAchievement{
		ID:                   row.ID,
		Name:                 row.Name,
		DescriptionPending:   row.DescriptionPending,
		DescriptionCompleted: row.DescriptionCompleted,
		NotificationText:     row.NotificationText,
		ImagePending:         row.ImagePending,
		ImageCompleted:       row.ImageCompleted,
		Points:               int(row.Points),
		Hidden:               hidden,
		AwardableFrom:        awardableFrom,
		ProjectID:            row.ProjectID,
		EventID:              row.EventID,
		ChallengeID:          row.ChallengeID,
		TotalItems:           totalItems,
	}
}

// convertPublishedStreakAchievementRow converts GetPublishedStreakAchievementsByExternalContentRow to StreakAchievement model
func convertPublishedStreakAchievementRow(row *sqlc.GetPublishedStreakAchievementsByExternalContentRow) *model.StreakAchievement {
	// Count streak items from JSON if available
	totalItems := 0
	if row.StreakItems != nil {
		var itemsData []map[string]interface{}
		jsonBytes, err := json.Marshal(row.StreakItems)
		if err == nil {
			if err := json.Unmarshal(jsonBytes, &itemsData); err == nil {
				totalItems = len(itemsData)
			}
		}
	}

	hidden := false
	if row.Hidden != nil {
		hidden = *row.Hidden
	}

	var awardableFrom *scalars.DateTime
	if row.AwardableFrom.Valid {
		awardableFrom = &scalars.DateTime{Time: row.AwardableFrom.Time}
	}

	return &model.StreakAchievement{
		ID:                   row.ID,
		Name:                 row.Name,
		DescriptionPending:   row.DescriptionPending,
		DescriptionCompleted: row.DescriptionCompleted,
		NotificationText:     row.NotificationText,
		ImagePending:         row.ImagePending,
		ImageCompleted:       row.ImageCompleted,
		Points:               int(row.Points),
		Hidden:               hidden,
		AwardableFrom:        awardableFrom,
		ProjectID:            row.ProjectID,
		EventID:              row.EventID,
		ChallengeID:          row.ChallengeID,
		TotalItems:           totalItems,
	}
}
