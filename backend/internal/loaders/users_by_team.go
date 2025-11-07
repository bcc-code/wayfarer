package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// usersByTeamBatchFunc batches loading users by team IDs
func usersByTeamBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.User] {
	return func(ctx context.Context, teamIDs []string) []*dataloader.Result[[]*model.User] {
		// Check cache first for each team ID
		usersByTeam := make(map[string][]*model.User)
		missingTeamIDs := []string{}

		for _, teamID := range teamIDs {
			cacheKey := cache.UsersByTeamKey(teamID)
			if cached, ok := c.Get(cacheKey); ok {
				if users, ok := cached.([]*model.User); ok {
					usersByTeam[teamID] = users
					continue
				}
			}
			missingTeamIDs = append(missingTeamIDs, teamID)
		}

		// Query database only for cache misses
		if len(missingTeamIDs) > 0 {
			rows, err := db.Queries.GetUsersByTeamIDs(ctx, missingTeamIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.User], len(teamIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.User]{Error: err}
				}
				return results
			}

			// Group users by team ID and convert to GraphQL model
			for _, row := range rows {
				// Convert birthdate to string (always valid since birthdate is required)
				birthdateStr := row.Birthdate.Time.Format("2006-01-02")

				user := &model.User{
					ID:        row.ID,
					MembersID: row.MembersID,
					Gender:    model.Gender(row.Gender),
					ChurchID:  row.ChurchID,
					Birthdate: birthdateStr,
					Email:     row.Email,
					Name:      row.Name,
					Image:     row.AvatarUrl,
				}
				usersByTeam[row.TeamID] = append(usersByTeam[row.TeamID], user)
			}

			// Populate cache for each team, including empty results
			for _, teamID := range missingTeamIDs {
				users := usersByTeam[teamID]
				if users == nil {
					users = []*model.User{} // Empty slice, not nil
				}
				usersByTeam[teamID] = users
				c.Set(cache.UsersByTeamKey(teamID), users)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.User], len(teamIDs))
		for i, teamID := range teamIDs {
			users := usersByTeam[teamID]
			if users == nil {
				users = []*model.User{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.User]{Data: users}
		}
		return results
	}
}
