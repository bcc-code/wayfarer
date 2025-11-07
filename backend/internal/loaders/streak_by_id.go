package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// streakByIDBatchFunc batches loading streaks by IDs
func streakByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Streak] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.Streak] {
		// Check cache first for each ID
		streakMap := make(map[string]*model.Streak)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.StreakKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if streak, ok := cached.(*model.Streak); ok {
					streakMap[id] = streak
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetStreaksByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.Streak], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.Streak]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				streak := &model.Streak{
					ID:          row.ID,
					Name:        row.Name,
					Description: row.Description,
					ProjectID:   row.ProjectID,
				}

				streakMap[row.ID] = streak
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.StreakKey(row.ID), streak)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.Streak], len(ids))
		for i, id := range ids {
			if streak, ok := streakMap[id]; ok {
				results[i] = &dataloader.Result[*model.Streak]{Data: streak}
			} else {
				results[i] = &dataloader.Result[*model.Streak]{
					Error: fmt.Errorf("streak not found: %s", id),
				}
			}
		}
		return results
	}
}
