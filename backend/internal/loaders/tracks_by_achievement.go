package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// tracksByAchievementBatchFunc batches loading tracks by achievement IDs
func tracksByAchievementBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.Track] {
	return func(ctx context.Context, achievementIDs []string) []*dataloader.Result[[]model.Track] {
		// Check cache first for each achievement ID
		tracksByAchievement := make(map[string][]model.Track)
		missingAchievementIDs := []string{}

		for _, achievementID := range achievementIDs {
			cacheKey := cache.TracksByAchievementKey(achievementID)
			if cached, ok := c.Get(cacheKey); ok {
				if tracks, ok := cached.([]model.Track); ok {
					tracksByAchievement[achievementID] = tracks
					continue
				}
			}
			missingAchievementIDs = append(missingAchievementIDs, achievementID)
		}

		// Query database only for cache misses
		if len(missingAchievementIDs) > 0 {
			rows, err := db.Queries.GetTracksByAchievementIDs(ctx, missingAchievementIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]model.Track], len(achievementIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]model.Track]{Error: err}
				}
				return results
			}

			// Group tracks by achievement ID and convert to GraphQL model
			for _, row := range rows {
				track := model.Track{
					ID:                row.ID,
					ExternalContentID: row.ExternalContentID,
				}
				tracksByAchievement[row.AchievementID] = append(tracksByAchievement[row.AchievementID], track)
			}

			// Populate cache for each achievement, including empty results
			for _, achievementID := range missingAchievementIDs {
				tracks := tracksByAchievement[achievementID]
				if tracks == nil {
					tracks = []model.Track{} // Empty slice, not nil
				}
				tracksByAchievement[achievementID] = tracks
				c.Set(cache.TracksByAchievementKey(achievementID), tracks)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]model.Track], len(achievementIDs))
		for i, achievementID := range achievementIDs {
			tracks := tracksByAchievement[achievementID]
			if tracks == nil {
				tracks = []model.Track{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]model.Track]{Data: tracks}
		}
		return results
	}
}
