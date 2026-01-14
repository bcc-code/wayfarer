package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// teamsByUserBatchFunc batches loading teams by user IDs
func teamsByUserBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.Team] {
	return func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.Team] {
		// Check cache first for each user ID
		teamsByUser := make(map[string][]*model.Team)
		missingUserIDs := []string{}

		for _, userID := range userIDs {
			cacheKey := cache.TeamsByUserKey(userID)
			if cached, ok := c.Get(cacheKey); ok {
				if teams, ok := cached.([]*model.Team); ok {
					teamsByUser[userID] = teams
					continue
				}
			}
			missingUserIDs = append(missingUserIDs, userID)
		}

		// Query database only for cache misses
		if len(missingUserIDs) > 0 {
			rows, err := db.Queries.GetTeamsByUserIDs(ctx, missingUserIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.Team], len(userIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.Team]{Error: err}
				}
				return results
			}

			// Group teams by user ID and convert to GraphQL model
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
					ID:                  row.ID,
					ProjectID:           row.ProjectID,
					Name:                row.Name,
					Description:         description,
					SuperTeamID:         superTeamID,
					JoinCode:            row.JoinCode,
					LeaderboardExcluded: row.LeaderboardExcluded,
				}
				teamsByUser[row.UserID] = append(teamsByUser[row.UserID], team)
			}

			// Populate cache for each user, including empty results
			for _, userID := range missingUserIDs {
				teams := teamsByUser[userID]
				if teams == nil {
					teams = []*model.Team{} // Empty slice, not nil
				}
				teamsByUser[userID] = teams
				c.Set(cache.TeamsByUserKey(userID), teams)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.Team], len(userIDs))
		for i, userID := range userIDs {
			teams := teamsByUser[userID]
			if teams == nil {
				teams = []*model.Team{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.Team]{Data: teams}
		}
		return results
	}
}
