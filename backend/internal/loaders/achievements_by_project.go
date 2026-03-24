package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// achievementsByProjectBatchFunc batches loading achievements by project IDs
func achievementsByProjectBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.Achievement] {
	return func(ctx context.Context, projectIDs []string) []*dataloader.Result[[]model.Achievement] {
		// Check cache first for each project ID
		achievementsByProject := make(map[string][]model.Achievement)
		missingProjectIDs := []string{}

		for _, projectID := range projectIDs {
			cacheKey := cache.AchievementsByProjectKey(projectID)
			if cached, ok := c.Get(cacheKey); ok {
				if achievements, ok := cached.([]model.Achievement); ok {
					achievementsByProject[projectID] = achievements
					continue
				}
			}
			missingProjectIDs = append(missingProjectIDs, projectID)
		}

		// Query database only for cache misses
		if len(missingProjectIDs) > 0 {
			rows, err := db.Queries.GetAchievementsByProjectIDs(ctx, missingProjectIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]model.Achievement], len(projectIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]model.Achievement]{Error: err}
				}
				return results
			}

			// Convert rows to achievements
			for _, row := range rows {
				achievement, err := convertRowToAchievement(row)
				if err != nil {
					results := make([]*dataloader.Result[[]model.Achievement], len(projectIDs))
					for i := range results {
						results[i] = &dataloader.Result[[]model.Achievement]{Error: err}
					}
					return results
				}
				achievementsByProject[row.ProjectID] = append(achievementsByProject[row.ProjectID], achievement)
			}

			// Populate cache for each project, including empty results
			for _, projectID := range missingProjectIDs {
				achievements := achievementsByProject[projectID]
				if achievements == nil {
					achievements = []model.Achievement{} // Empty slice, not nil
				}
				achievementsByProject[projectID] = achievements
				c.Set(cache.AchievementsByProjectKey(projectID), achievements)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]model.Achievement], len(projectIDs))
		for i, projectID := range projectIDs {
			achievements := achievementsByProject[projectID]
			if achievements == nil {
				achievements = []model.Achievement{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]model.Achievement]{Data: achievements}
		}
		return results
	}
}

// convertRowToAchievement converts a database row to the appropriate Achievement type
func convertRowToAchievement(row *sqlc.GetAchievementsByProjectIDsRow) (model.Achievement, error) {
	hidden := false
	if row.Hidden != nil {
		hidden = *row.Hidden
	}

	var awardableFrom *scalars.DateTime
	if row.AwardableFrom.Valid {
		awardableFrom = &scalars.DateTime{Time: row.AwardableFrom.Time}
	}

	switch row.AchievementType {
	case "SIMPLE":
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
		}, nil

	case "CONTENT":
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
			// Items, UserCompletedItems, NextItem, TotalItems, and CompletedItemCount will be populated by resolvers
		}, nil

	case "STREAK":
		if row.StreakAchievementID == nil {
			return nil, fmt.Errorf("streak achievement missing streak data")
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
			// Items, UserCompletedItems, NextItem, TotalItems, and CompletedItemCount will be populated by resolvers
		}, nil

	case "QUIZ":
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

	default:
		return nil, fmt.Errorf("unknown achievement type: %s", row.AchievementType)
	}
}
