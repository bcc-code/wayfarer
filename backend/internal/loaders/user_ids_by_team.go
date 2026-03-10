package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/graph-gophers/dataloader/v7"
)

// userIDsByTeamBatchFunc batches loading user IDs by team IDs
func userIDsByTeamBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]string] {
	return func(ctx context.Context, teamIDs []string) []*dataloader.Result[[]string] {
		// Check cache first for each team ID
		userIDsByTeam := make(map[string][]string)
		missingTeamIDs := []string{}

		for _, teamID := range teamIDs {
			cacheKey := cache.UserIDsByTeamKey(teamID)
			if cached, ok := c.Get(cacheKey); ok {
				if userIDs, ok := cached.([]string); ok {
					userIDsByTeam[teamID] = userIDs
					continue
				}
			}
			missingTeamIDs = append(missingTeamIDs, teamID)
		}

		// Query database only for cache misses
		if len(missingTeamIDs) > 0 {
			rows, err := db.Queries.GetUserIDsByTeamIDs(ctx, missingTeamIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]string], len(teamIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]string]{Error: err}
				}
				return results
			}

			// Group user IDs by team ID
			for _, row := range rows {
				userIDsByTeam[row.TeamID] = append(userIDsByTeam[row.TeamID], row.UserID)
			}

			// Populate cache for each team, including empty results
			for _, teamID := range missingTeamIDs {
				userIDs := userIDsByTeam[teamID]
				if userIDs == nil {
					userIDs = []string{}
				}
				userIDsByTeam[teamID] = userIDs
				c.Set(cache.UserIDsByTeamKey(teamID), userIDs)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]string], len(teamIDs))
		for i, teamID := range teamIDs {
			userIDs := userIDsByTeam[teamID]
			if userIDs == nil {
				userIDs = []string{}
			}
			results[i] = &dataloader.Result[[]string]{Data: userIDs}
		}
		return results
	}
}
