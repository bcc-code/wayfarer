package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// superTeamByIDBatchFunc batches loading super teams by IDs
func superTeamByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.SuperTeam] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.SuperTeam] {
		// Check cache first for each ID
		superTeamMap := make(map[string]*model.SuperTeam)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.SuperTeamKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if superTeam, ok := cached.(*model.SuperTeam); ok {
					superTeamMap[id] = superTeam
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetSuperTeamsByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.SuperTeam], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.SuperTeam]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				description := ""
				if row.Description != nil {
					description = *row.Description
				}

				superTeam := &model.SuperTeam{
					ID:          row.ID,
					Name:        row.Name,
					Description: description,
					ImageUrl:    row.ImageUrl,
					Color:       row.Color,
					ProjectID:   row.ProjectID,
				}

				superTeamMap[row.ID] = superTeam
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.SuperTeamKey(row.ID), superTeam)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.SuperTeam], len(ids))
		for i, id := range ids {
			if superTeam, ok := superTeamMap[id]; ok {
				results[i] = &dataloader.Result[*model.SuperTeam]{Data: superTeam}
			} else {
				results[i] = &dataloader.Result[*model.SuperTeam]{
					Error: fmt.Errorf("super team not found: %s", id),
				}
			}
		}
		return results
	}
}
