package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// cachedLeaderboardEntry stores leaderboard data without user-specific tags
type cachedLeaderboardEntry struct {
	ID          string
	Name        string
	Description string
	Score       int
	Rank        int
	Image       *string
}

// cachedLeaderboardTags stores the viewer-independent tag assignments for a leaderboard
// Maps entry ID to their tags
type cachedLeaderboardTags map[string][]model.LeaderboardEntryTag

// teamMemberLeaderboardBatchFunc batches loading team member leaderboards by team IDs
func teamMemberLeaderboardBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.LeaderboardEntry] {
	return func(ctx context.Context, teamIDs []string) []*dataloader.Result[[]model.LeaderboardEntry] {
		// Store cached entries (without tags) per team
		cachedEntries := make(map[string][]cachedLeaderboardEntry)
		missingTeamIDs := []string{}

		// Check cache for leaderboard data (shared across all users)
		for _, teamID := range teamIDs {
			cacheKey := cache.TeamMemberLeaderboardKey(teamID)
			if cached, ok := c.Get(cacheKey); ok {
				if entries, ok := cached.([]cachedLeaderboardEntry); ok {
					cachedEntries[teamID] = entries
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

				// Convert database rows to cached entries (without tags)
				entries := make([]cachedLeaderboardEntry, len(rows))
				for i, row := range rows {
					entries[i] = cachedLeaderboardEntry{
						ID:          row.UserID,
						Name:        row.UserName,
						Description: row.ChurchName,
						Score:       int(row.Score),
						Rank:        int(row.Rank),
						Image:       row.AvatarUrl,
					}
				}

				cachedEntries[teamID] = entries

				// Cache the leaderboard data (without tags)
				c.Set(cache.TeamMemberLeaderboardKey(teamID), entries)
			}
		}

		// Build results with viewer-independent tags (ME is added by the resolver)
		results := make([]*dataloader.Result[[]model.LeaderboardEntry], len(teamIDs))
		for i, teamID := range teamIDs {
			entries := cachedEntries[teamID]
			if entries == nil {
				results[i] = &dataloader.Result[[]model.LeaderboardEntry]{Data: []model.LeaderboardEntry{}}
				continue
			}

			tags := getOrComputeTags(ctx, db, c, teamID, entries)

			// Convert cached entries to model with tags
			modelEntries := make([]model.LeaderboardEntry, len(entries))
			for j, entry := range entries {
				entryTags := tags[entry.ID]
				if entryTags == nil {
					entryTags = []model.LeaderboardEntryTag{}
				}

				rank := entry.Rank
				modelEntries[j] = model.LeaderboardEntry{
					ID:          entry.ID,
					Name:        entry.Name,
					Description: entry.Description,
					Score:       entry.Score,
					Rank:        &rank,
					Tags:        entryTags,
					ImageURL:    entry.Image,
				}
			}

			results[i] = &dataloader.Result[[]model.LeaderboardEntry]{Data: modelEntries}
		}

		return results
	}
}

// getOrComputeTags computes the viewer-independent tags for a team leaderboard.
// TEAM_LEAD tags are cached per team. The ME tag is not set here: a batch runs
// with the context of whichever request started it, so the viewer is unknown.
// The MemberLeaderboard resolver adds ME for its own caller.
func getOrComputeTags(ctx context.Context, db *database.DB, c *cache.CacheWithRegistry, teamID string, entries []cachedLeaderboardEntry) cachedLeaderboardTags {
	// Get or compute TEAM_LEAD tags (cached per team, not per user)
	teamLeadTags := getOrComputeTeamLeadTags(ctx, db, c, teamID, entries)

	tags := make(cachedLeaderboardTags)
	for _, entry := range entries {
		entryTags := []model.LeaderboardEntryTag{}

		// TEAM_LEAD tag - from cached per-team data
		if teamLeadTags[entry.ID] {
			entryTags = append(entryTags, model.LeaderboardEntryTagTeamLead)
		}

		if len(entryTags) > 0 {
			tags[entry.ID] = entryTags
		}
	}

	return tags
}

// getOrComputeTeamLeadTags returns a map of entry IDs that are team leads for this team.
// This is cached per team (not per viewer) since TEAM_LEAD is viewer-independent.
func getOrComputeTeamLeadTags(ctx context.Context, db *database.DB, c *cache.CacheWithRegistry, teamID string, entries []cachedLeaderboardEntry) map[string]bool {
	cacheKey := cache.TeamMemberLeaderboardTeamLeadTagsKey(teamID)
	if cached, ok := c.Get(cacheKey); ok {
		if tags, ok := cached.(map[string]bool); ok {
			return tags
		}
	}

	// Cache miss - collect all entry IDs for batched role loading
	entryIDs := make([]string, len(entries))
	for i, entry := range entries {
		entryIDs[i] = entry.ID
	}

	// Batch load roles for all entries in ONE query
	teamLeadTags := make(map[string]bool)
	if len(entryIDs) > 0 && db != nil {
		rows, err := db.Queries.GetAllRolesForUsers(ctx, entryIDs)
		if err == nil {
			for _, row := range rows {
				role := convertToUserRole(row)
				if role.Role == model.RoleTypeTeamLead && role.Scope != nil {
					if role.Scope.Type == model.ScopeTypeTeam && role.Scope.ID == teamID {
						teamLeadTags[row.UserID] = true
					}
				}
			}
		}
	}

	// Cache TEAM_LEAD tags per team
	c.Set(cacheKey, teamLeadTags)
	return teamLeadTags
}

// convertToUserRole converts a database role row to a model.UserRole
func convertToUserRole(row *sqlc.GetAllRolesForUsersRow) *model.UserRole {
	role := &model.UserRole{
		ID:   row.ID,
		Role: model.RoleType(row.Role),
	}

	// Build scope if any scope fields are set
	if row.ChurchID != nil || row.ProjectID != nil || row.TeamID != nil {
		scope := &model.RoleScope{}
		if row.ChurchID != nil {
			scope.Type = model.ScopeTypeChurch
			scope.ID = *row.ChurchID
		} else if row.ProjectID != nil {
			scope.Type = model.ScopeTypeProject
			scope.ID = *row.ProjectID
		} else if row.TeamID != nil {
			scope.Type = model.ScopeTypeTeam
			scope.ID = *row.TeamID
		}
		role.Scope = scope
	}

	return role
}
