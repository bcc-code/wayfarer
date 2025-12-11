package api

import (
	"encoding/json"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

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
	default:
		return nil, fmt.Errorf("unknown achievement type: %s", row.AchievementType)
	}
}

func convertRowToSimpleAchievement(row *sqlc.GetAchievementsFilteredCursorRow, hidden bool) model.Achievement {
	return &model.SimpleAchievement{
		ID:                   row.ID,
		Name:                 row.Name,
		DescriptionPending:   row.DescriptionPending,
		DescriptionCompleted: row.DescriptionCompleted,
		ImagePending:         row.ImagePending,
		ImageCompleted:       row.ImageCompleted,
		Points:               int(row.Points),
		Hidden:               hidden,
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

	return &model.ContentAchievement{
		ID:                   row.ID,
		Name:                 row.Name,
		DescriptionPending:   row.DescriptionPending,
		DescriptionCompleted: row.DescriptionCompleted,
		ImagePending:         row.ImagePending,
		ImageCompleted:       row.ImageCompleted,
		Points:               int(row.Points),
		Hidden:               hidden,
		ProjectID:            row.ProjectID,
		EventID:              row.EventID,
		ChallengeID:          row.ChallengeID,
		TotalItems:           totalItems,
		// Items, UserCompletedItems, NextItem, and CompletedItemCount will be populated by resolvers
	}, nil
}

func convertRowToStreakAchievement(row *sqlc.GetAchievementsFilteredCursorRow, hidden bool) (model.Achievement, error) {
	if row.StreakID == nil || row.NeededStreak == nil {
		return nil, fmt.Errorf("streak achievement missing streak data")
	}

	return &model.StreakAchievement{
		ID:                   row.ID,
		Name:                 row.Name,
		DescriptionPending:   row.DescriptionPending,
		DescriptionCompleted: row.DescriptionCompleted,
		ImagePending:         row.ImagePending,
		ImageCompleted:       row.ImageCompleted,
		Points:               int(row.Points),
		Hidden:               hidden,
		ProjectID:            row.ProjectID,
		EventID:              row.EventID,
		ChallengeID:          row.ChallengeID,
		NeededStreak:         int(*row.NeededStreak),
		StreakID:             *row.StreakID,
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

	return &model.ContentAchievement{
		ID:                   row.ID,
		Name:                 row.Name,
		DescriptionPending:   row.DescriptionPending,
		DescriptionCompleted: row.DescriptionCompleted,
		ImagePending:         row.ImagePending,
		ImageCompleted:       row.ImageCompleted,
		Points:               int(row.Points),
		Hidden:               hidden,
		ProjectID:            row.ProjectID,
		EventID:              row.EventID,
		ChallengeID:          row.ChallengeID,
		TotalItems:           totalItems,
	}
}
