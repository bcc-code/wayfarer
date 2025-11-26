package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/graph-gophers/dataloader/v7"
)

// teamMemberLeaderboardBatchFunc batches loading team member leaderboards by team IDs
func teamMemberLeaderboardBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.LeaderboardEntry] {
	return func(ctx context.Context, teamIDs []string) []*dataloader.Result[[]model.LeaderboardEntry] {
		// Get current user ID for isMe field
		currentUserID, _ := middleware.GetUserID(ctx)

		// Check cache first for each team ID
		leaderboards := make(map[string][]model.LeaderboardEntry)
		missingTeamIDs := []string{}

		for _, teamID := range teamIDs {
			cacheKey := cache.TeamMemberLeaderboardKey(teamID)
			if cached, ok := c.Get(cacheKey); ok {
				if entries, ok := cached.([]model.LeaderboardEntry); ok {
					leaderboards[teamID] = entries
					continue
				}
			}
			missingTeamIDs = append(missingTeamIDs, teamID)
		}

		// Query database only for cache misses
		if len(missingTeamIDs) > 0 {
			for _, teamID := range missingTeamIDs {
				rows, err := db.Queries.GetTeamMemberLeaderboard(ctx, teamID)
				if err != nil {
					results := make([]*dataloader.Result[[]model.LeaderboardEntry], len(teamIDs))
					for i := range results {
						results[i] = &dataloader.Result[[]model.LeaderboardEntry]{Error: err}
					}
					return results
				}

				// Convert database rows to GraphQL LeaderboardEntry model
				entries := make([]model.LeaderboardEntry, len(rows))
				for i, row := range rows {
					// Convert score from interface{} to int
					score := 0
					if row.Score != nil {
						if scoreInt64, ok := row.Score.(int64); ok {
							score = int(scoreInt64)
						}
					}

					entries[i] = model.LeaderboardEntry{
						ID:          row.UserID,
						Name:        row.UserName,
						Description: row.ChurchName,
						Score:       score,
						Rank:        int(row.Rank),
						IsMe:        row.UserID == currentUserID,
						Image:       row.AvatarUrl,
					}
				}

				leaderboards[teamID] = entries

				// Cache the result
				c.Set(cache.TeamMemberLeaderboardKey(teamID), entries)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]model.LeaderboardEntry], len(teamIDs))
		for i, teamID := range teamIDs {
			entries := leaderboards[teamID]
			if entries == nil {
				entries = []model.LeaderboardEntry{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]model.LeaderboardEntry]{Data: entries}
		}
		return results
	}
}
