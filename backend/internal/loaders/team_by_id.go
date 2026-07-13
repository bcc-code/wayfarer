package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// teamByIDBatchFunc batches loading teams by IDs
func teamByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Team] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.Team] {
		// Check cache first for each ID
		teamMap := make(map[string]*model.Team)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.TeamKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if team, ok := cached.(*model.Team); ok {
					teamMap[id] = team
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetTeamsByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.Team], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.Team]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				description := ""
				if row.Description != nil {
					description = *row.Description
				}

				team := &model.Team{
					ID:                  row.ID,
					Name:                row.Name,
					Description:         description,
					ProjectID:           row.ProjectID,
					JoinCode:            row.JoinCode,
					SuperTeamID:         row.SuperTeamID,
					LeaderboardExcluded: row.LeaderboardExcluded,
				}

				teamMap[row.ID] = team
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.TeamKey(row.ID), team)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.Team], len(ids))
		for i, id := range ids {
			if team, ok := teamMap[id]; ok {
				results[i] = &dataloader.Result[*model.Team]{Data: team}
			} else {
				results[i] = &dataloader.Result[*model.Team]{
					Error: fmt.Errorf("team not found: %s", id),
				}
			}
		}
		return results
	}
}
