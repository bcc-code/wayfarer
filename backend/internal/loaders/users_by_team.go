package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// usersByTeamBatchFunc batches loading team members by team IDs
func usersByTeamBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.TeamMember] {
	return func(ctx context.Context, teamIDs []string) []*dataloader.Result[[]*model.TeamMember] {
		// Check cache first for each team ID
		teamMembers := make(map[string][]*model.TeamMember)
		missingTeamIDs := []string{}

		for _, teamID := range teamIDs {
			cacheKey := cache.UsersByTeamKey(teamID)
			if cached, ok := c.Get(cacheKey); ok {
				if members, ok := cached.([]*model.TeamMember); ok {
					teamMembers[teamID] = members
					continue
				}
			}
			missingTeamIDs = append(missingTeamIDs, teamID)
		}

		// Query database only for cache misses
		if len(missingTeamIDs) > 0 {
			rows, err := db.Queries.GetUsersByTeamIDs(ctx, missingTeamIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.TeamMember], len(teamIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.TeamMember]{Error: err}
				}
				return results
			}

			// Group team members by team ID and convert to GraphQL model
			for _, row := range rows {
				member := &model.TeamMember{
					ID:         row.ID,
					Name:       row.Name,
					IsTeamLead: row.IsTeamLead,
					JoinedAt:   row.JoinedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
					UserID:     row.ID,
					ChurchID:   row.ChurchID,
				}
				teamMembers[row.TeamID] = append(teamMembers[row.TeamID], member)
			}

			// Populate cache for each team, including empty results
			for _, teamID := range missingTeamIDs {
				members := teamMembers[teamID]
				if members == nil {
					members = []*model.TeamMember{} // Empty slice, not nil
				}
				teamMembers[teamID] = members
				c.Set(cache.UsersByTeamKey(teamID), members)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.TeamMember], len(teamIDs))
		for i, teamID := range teamIDs {
			members := teamMembers[teamID]
			if members == nil {
				members = []*model.TeamMember{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.TeamMember]{Data: members}
		}
		return results
	}
}
