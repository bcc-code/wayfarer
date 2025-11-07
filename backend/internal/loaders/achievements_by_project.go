package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
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

	switch row.AchievementType {
	case "SIMPLE":
		return &model.SimpleAchievement{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			Image:       row.ImageUrl,
			Points:      int(row.Points),
			Hidden:      hidden,
			ProjectID:   row.ProjectID,
			EventID:     row.EventID,
			ChallengeID: row.ChallengeID,
		}, nil

	case "READING":
		return &model.ReadingAchievement{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			Image:       row.ImageUrl,
			Points:      int(row.Points),
			Hidden:      hidden,
			ProjectID:   row.ProjectID,
			EventID:     row.EventID,
			ChallengeID: row.ChallengeID,
			Articles:    []model.Article{}, // Will be populated by resolver if needed
			UserHasRead: []model.Article{}, // Will be populated by resolver if needed
			NextArticle: nil,               // Will be populated by resolver if needed
		}, nil

	case "LISTENING":
		return &model.ListeningAchievement{
			ID:              row.ID,
			Name:            row.Name,
			Description:     row.Description,
			Image:           row.ImageUrl,
			Points:          int(row.Points),
			Hidden:          hidden,
			ProjectID:       row.ProjectID,
			EventID:         row.EventID,
			ChallengeID:     row.ChallengeID,
			Tracks:          []model.Track{}, // Will be populated by resolver if needed
			UserHasListened: []model.Track{}, // Will be populated by resolver if needed
			NextTrack:       nil,             // Will be populated by resolver if needed
		}, nil

	case "STREAK":
		if row.StreakID == nil || row.NeededStreak == nil {
			return nil, fmt.Errorf("streak achievement missing required fields: streak_id or needed_streak")
		}
		return &model.StreakAchievement{
			ID:           row.ID,
			Name:         row.Name,
			Description:  row.Description,
			Image:        row.ImageUrl,
			Points:       int(row.Points),
			Hidden:       hidden,
			ProjectID:    row.ProjectID,
			EventID:      row.EventID,
			ChallengeID:  row.ChallengeID,
			StreakID:     *row.StreakID,
			NeededStreak: int(*row.NeededStreak),
			Streak:       nil, // Will be populated by resolver
		}, nil

	default:
		return nil, fmt.Errorf("unknown achievement type: %s", row.AchievementType)
	}
}
