package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/graph-gophers/dataloader/v7"
)

// userIDsBySuperTeamBatchFunc batches loading user IDs by super team IDs
func userIDsBySuperTeamBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]string] {
	return func(ctx context.Context, superTeamIDs []string) []*dataloader.Result[[]string] {
		// Check cache first for each super team ID
		userIDsBySuperTeam := make(map[string][]string)
		missingSuperTeamIDs := []string{}

		for _, superTeamID := range superTeamIDs {
			cacheKey := cache.UserIDsBySuperTeamKey(superTeamID)
			if cached, ok := c.Get(cacheKey); ok {
				if userIDs, ok := cached.([]string); ok {
					userIDsBySuperTeam[superTeamID] = userIDs
					continue
				}
			}
			missingSuperTeamIDs = append(missingSuperTeamIDs, superTeamID)
		}

		// Query database only for cache misses
		if len(missingSuperTeamIDs) > 0 {
			rows, err := db.Queries.GetUserIDsBySuperTeamIDs(ctx, missingSuperTeamIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]string], len(superTeamIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]string]{Error: err}
				}
				return results
			}

			// Group user IDs by super team ID
			for _, row := range rows {
				if row.SuperTeamID != nil {
					userIDsBySuperTeam[*row.SuperTeamID] = append(userIDsBySuperTeam[*row.SuperTeamID], row.UserID)
				}
			}

			// Populate cache for each super team, including empty results
			for _, superTeamID := range missingSuperTeamIDs {
				userIDs := userIDsBySuperTeam[superTeamID]
				if userIDs == nil {
					userIDs = []string{}
				}
				userIDsBySuperTeam[superTeamID] = userIDs
				c.Set(cache.UserIDsBySuperTeamKey(superTeamID), userIDs)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]string], len(superTeamIDs))
		for i, superTeamID := range superTeamIDs {
			userIDs := userIDsBySuperTeam[superTeamID]
			if userIDs == nil {
				userIDs = []string{}
			}
			results[i] = &dataloader.Result[[]string]{Data: userIDs}
		}
		return results
	}
}
