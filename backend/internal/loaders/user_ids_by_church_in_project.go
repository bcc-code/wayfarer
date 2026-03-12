package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// ChurchProjectKey is a composite key for church + project lookups
type ChurchProjectKey struct {
	ChurchID  string
	ProjectID string
}

// userIDsByChurchInProjectBatchFunc batches loading user IDs by church in a specific project
func userIDsByChurchInProjectBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []ChurchProjectKey) []*dataloader.Result[[]string] {
	return func(ctx context.Context, keys []ChurchProjectKey) []*dataloader.Result[[]string] {
		// Group keys by project ID for efficient querying
		keysByProject := make(map[string][]ChurchProjectKey)
		for _, key := range keys {
			keysByProject[key.ProjectID] = append(keysByProject[key.ProjectID], key)
		}

		// Check cache and collect missing keys
		userIDsByKey := make(map[ChurchProjectKey][]string)
		missingKeysByProject := make(map[string][]string) // projectID -> churchIDs

		for _, key := range keys {
			cacheKey := cache.UserIDsByChurchInProjectKey(key.ChurchID, key.ProjectID)
			if cached, ok := c.Get(cacheKey); ok {
				if userIDs, ok := cached.([]string); ok {
					userIDsByKey[key] = userIDs
					continue
				}
			}
			missingKeysByProject[key.ProjectID] = append(missingKeysByProject[key.ProjectID], key.ChurchID)
		}

		// Query database for each project with missing church IDs
		for projectID, churchIDs := range missingKeysByProject {
			if len(churchIDs) == 0 {
				continue
			}

			rows, err := db.Queries.GetUserIDsByChurchAndProject(ctx, sqlc.GetUserIDsByChurchAndProjectParams{
				Churchids: churchIDs,
				Projectid: projectID,
			})
			if err != nil {
				results := make([]*dataloader.Result[[]string], len(keys))
				for i := range results {
					results[i] = &dataloader.Result[[]string]{Error: err}
				}
				return results
			}

			// Group user IDs by church ID
			userIDsByChurch := make(map[string][]string)
			for _, row := range rows {
				userIDsByChurch[row.ChurchID] = append(userIDsByChurch[row.ChurchID], row.UserID)
			}

			// Populate results and cache for each church
			for _, churchID := range churchIDs {
				key := ChurchProjectKey{ChurchID: churchID, ProjectID: projectID}
				userIDs := userIDsByChurch[churchID]
				if userIDs == nil {
					userIDs = []string{}
				}
				userIDsByKey[key] = userIDs
				c.Set(cache.UserIDsByChurchInProjectKey(churchID, projectID), userIDs)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]string], len(keys))
		for i, key := range keys {
			userIDs := userIDsByKey[key]
			if userIDs == nil {
				userIDs = []string{}
			}
			results[i] = &dataloader.Result[[]string]{Data: userIDs}
		}
		return results
	}
}
