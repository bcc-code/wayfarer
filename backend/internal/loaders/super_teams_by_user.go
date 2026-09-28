package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// superTeamsByUserBatchFunc batches loading super teams by user IDs
func superTeamsByUserBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.SuperTeam] {
	return func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.SuperTeam] {
		// Check cache first for each user ID
		superTeamsByUser := make(map[string][]*model.SuperTeam)
		missingUserIDs := []string{}

		for _, userID := range userIDs {
			cacheKey := cache.SuperTeamsByUserKey(userID)
			if cached, ok := c.Get(cacheKey); ok {
				if superTeams, ok := cached.([]*model.SuperTeam); ok {
					superTeamsByUser[userID] = superTeams
					continue
				}
			}
			missingUserIDs = append(missingUserIDs, userID)
		}

		// Query database only for cache misses
		if len(missingUserIDs) > 0 {
			rows, err := db.Queries.GetSuperTeamsByUserIDs(ctx, missingUserIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.SuperTeam], len(userIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.SuperTeam]{Error: err}
				}
				return results
			}

			// Group super teams by user ID and convert to GraphQL model
			// Track which super teams we've already added for each user to handle DISTINCT
			seen := make(map[string]map[string]bool)
			for _, row := range rows {
				description := ""
				if row.Description != nil {
					description = *row.Description
				}

				// Initialize seen map for this user if needed
				if seen[row.UserID] == nil {
					seen[row.UserID] = make(map[string]bool)
				}

				// Skip if we've already added this super team for this user (DISTINCT)
				if seen[row.UserID][row.ID] {
					continue
				}
				seen[row.UserID][row.ID] = true

				superTeam := &model.SuperTeam{
					ID:          row.ID,
					ProjectID:   row.ProjectID,
					Name:        row.Name,
					Description: description,
					ImageUrl:    row.ImageUrl,
					Color:       row.Color,
				}
				superTeamsByUser[row.UserID] = append(superTeamsByUser[row.UserID], superTeam)
			}

			// Populate cache for each user, including empty results
			for _, userID := range missingUserIDs {
				superTeams := superTeamsByUser[userID]
				if superTeams == nil {
					superTeams = []*model.SuperTeam{} // Empty slice, not nil
				}
				superTeamsByUser[userID] = superTeams
				c.Set(cache.SuperTeamsByUserKey(userID), superTeams)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.SuperTeam], len(userIDs))
		for i, userID := range userIDs {
			superTeams := superTeamsByUser[userID]
			if superTeams == nil {
				superTeams = []*model.SuperTeam{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.SuperTeam]{Data: superTeams}
		}
		return results
	}
}
