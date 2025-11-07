package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// streaksByProjectBatchFunc batches loading streaks by project IDs
func streaksByProjectBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.Streak] {
	return func(ctx context.Context, projectIDs []string) []*dataloader.Result[[]*model.Streak] {
		// Check cache first for each project ID
		streaksByProject := make(map[string][]*model.Streak)
		missingProjectIDs := []string{}

		for _, projectID := range projectIDs {
			cacheKey := cache.StreaksByProjectKey(projectID)
			if cached, ok := c.Get(cacheKey); ok {
				if streaks, ok := cached.([]*model.Streak); ok {
					streaksByProject[projectID] = streaks
					continue
				}
			}
			missingProjectIDs = append(missingProjectIDs, projectID)
		}

		// Query database only for cache misses
		if len(missingProjectIDs) > 0 {
			rows, err := db.Queries.GetStreaksByProjectIDs(ctx, missingProjectIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.Streak], len(projectIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.Streak]{Error: err}
				}
				return results
			}

			// Group streaks by project ID and convert to GraphQL model
			for _, row := range rows {
				streak := &model.Streak{
					ID:          row.ID,
					Name:        row.Name,
					Description: row.Description,
					ProjectID:   row.ProjectID,
				}
				streaksByProject[row.ProjectID] = append(streaksByProject[row.ProjectID], streak)
			}

			// Populate cache for each project, including empty results
			for _, projectID := range missingProjectIDs {
				streaks := streaksByProject[projectID]
				if streaks == nil {
					streaks = []*model.Streak{} // Empty slice, not nil
				}
				streaksByProject[projectID] = streaks
				c.Set(cache.StreaksByProjectKey(projectID), streaks)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.Streak], len(projectIDs))
		for i, projectID := range projectIDs {
			streaks := streaksByProject[projectID]
			if streaks == nil {
				streaks = []*model.Streak{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.Streak]{Data: streaks}
		}
		return results
	}
}
