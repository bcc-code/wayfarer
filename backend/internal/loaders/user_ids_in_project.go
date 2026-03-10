package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/graph-gophers/dataloader/v7"
)

// userIDsInProjectBatchFunc batches loading user IDs by project IDs
func userIDsInProjectBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]string] {
	return func(ctx context.Context, projectIDs []string) []*dataloader.Result[[]string] {
		// Check cache first for each project ID
		userIDsByProject := make(map[string][]string)
		missingProjectIDs := []string{}

		for _, projectID := range projectIDs {
			cacheKey := cache.UserIDsInProjectKey(projectID)
			if cached, ok := c.Get(cacheKey); ok {
				if userIDs, ok := cached.([]string); ok {
					userIDsByProject[projectID] = userIDs
					continue
				}
			}
			missingProjectIDs = append(missingProjectIDs, projectID)
		}

		// Query database only for cache misses - one project at a time
		// since GetUserIDsInProject only takes a single project ID
		for _, projectID := range missingProjectIDs {
			userIDs, err := db.Queries.GetUserIDsInProject(ctx, projectID)
			if err != nil {
				results := make([]*dataloader.Result[[]string], len(projectIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]string]{Error: err}
				}
				return results
			}

			if userIDs == nil {
				userIDs = []string{}
			}
			userIDsByProject[projectID] = userIDs
			c.Set(cache.UserIDsInProjectKey(projectID), userIDs)
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]string], len(projectIDs))
		for i, projectID := range projectIDs {
			userIDs := userIDsByProject[projectID]
			if userIDs == nil {
				userIDs = []string{}
			}
			results[i] = &dataloader.Result[[]string]{Data: userIDs}
		}
		return results
	}
}
