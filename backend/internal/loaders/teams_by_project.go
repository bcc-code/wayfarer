package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// teamsByProjectBatchFunc batches loading teams by project IDs
func teamsByProjectBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.Team] {
	return func(ctx context.Context, projectIDs []string) []*dataloader.Result[[]*model.Team] {
		// Check cache first for each project ID
		teamsByProject := make(map[string][]*model.Team)
		missingProjectIDs := []string{}

		for _, projectID := range projectIDs {
			cacheKey := cache.TeamsByProjectKey(projectID)
			if cached, ok := c.Get(cacheKey); ok {
				if teams, ok := cached.([]*model.Team); ok {
					teamsByProject[projectID] = teams
					continue
				}
			}
			missingProjectIDs = append(missingProjectIDs, projectID)
		}

		// Query database only for cache misses
		if len(missingProjectIDs) > 0 {
			rows, err := db.Queries.GetTeamsByProjectIDs(ctx, missingProjectIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.Team], len(projectIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.Team]{Error: err}
				}
				return results
			}

			// Group teams by project ID and convert to GraphQL model
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

				teamsByProject[row.ProjectID] = append(teamsByProject[row.ProjectID], team)
			}

			// Populate cache for each project, including empty results
			for _, projectID := range missingProjectIDs {
				teams := teamsByProject[projectID]
				if teams == nil {
					teams = []*model.Team{} // Empty slice, not nil
				}
				teamsByProject[projectID] = teams
				c.Set(cache.TeamsByProjectKey(projectID), teams)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.Team], len(projectIDs))
		for i, projectID := range projectIDs {
			teams := teamsByProject[projectID]
			if teams == nil {
				teams = []*model.Team{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.Team]{Data: teams}
		}
		return results
	}
}
