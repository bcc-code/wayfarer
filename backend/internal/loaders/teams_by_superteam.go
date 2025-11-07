package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// teamsBySuperTeamBatchFunc batches loading teams by super team IDs
func teamsBySuperTeamBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.Team] {
	return func(ctx context.Context, superTeamIDs []string) []*dataloader.Result[[]*model.Team] {
		// Check cache first for each super team ID
		teamsBySuperTeam := make(map[string][]*model.Team)
		missingSuperTeamIDs := []string{}

		for _, superTeamID := range superTeamIDs {
			cacheKey := cache.TeamsBySuperTeamKey(superTeamID)
			if cached, ok := c.Get(cacheKey); ok {
				if teams, ok := cached.([]*model.Team); ok {
					teamsBySuperTeam[superTeamID] = teams
					continue
				}
			}
			missingSuperTeamIDs = append(missingSuperTeamIDs, superTeamID)
		}

		// Query database only for cache misses
		if len(missingSuperTeamIDs) > 0 {
			rows, err := db.Queries.GetTeamsBySuperTeamIDs(ctx, missingSuperTeamIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.Team], len(superTeamIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.Team]{Error: err}
				}
				return results
			}

			// Group teams by super team ID and convert to GraphQL model
			for _, row := range rows {
				description := ""
				if row.Description != nil {
					description = *row.Description
				}

				var superTeamID *string
				if row.SuperTeamID != nil {
					superTeamID = row.SuperTeamID
				}

				team := &model.Team{
					ID:          row.ID,
					ProjectID:   row.ProjectID,
					Name:        row.Name,
					Description: description,
					SuperTeamID: superTeamID,
				}

				// row.SuperTeamID is guaranteed to be non-nil because of the query's WHERE clause
				if row.SuperTeamID != nil {
					teamsBySuperTeam[*row.SuperTeamID] = append(teamsBySuperTeam[*row.SuperTeamID], team)
				}
			}

			// Populate cache for each super team, including empty results
			for _, superTeamID := range missingSuperTeamIDs {
				teams := teamsBySuperTeam[superTeamID]
				if teams == nil {
					teams = []*model.Team{} // Empty slice, not nil
				}
				teamsBySuperTeam[superTeamID] = teams
				c.Set(cache.TeamsBySuperTeamKey(superTeamID), teams)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.Team], len(superTeamIDs))
		for i, superTeamID := range superTeamIDs {
			teams := teamsBySuperTeam[superTeamID]
			if teams == nil {
				teams = []*model.Team{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.Team]{Data: teams}
		}
		return results
	}
}
