package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
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

// cachedLeaderboardTags stores the tag assignments for a specific user viewing the leaderboard
// Maps entry ID to their tags
type cachedLeaderboardTags map[string][]model.LeaderboardEntryTag

// teamMemberLeaderboardBatchFunc batches loading team member leaderboards by team IDs
func teamMemberLeaderboardBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.LeaderboardEntry] {
	return func(ctx context.Context, teamIDs []string) []*dataloader.Result[[]model.LeaderboardEntry] {
		currentUserID, _ := middleware.GetUserID(ctx)

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
					score := 0
					if row.Score != nil {
						if scoreInt64, ok := row.Score.(int64); ok {
							score = int(scoreInt64)
						}
					}

					entries[i] = cachedLeaderboardEntry{
						ID:          row.UserID,
						Name:        row.UserName,
						Description: row.ChurchName,
						Score:       score,
						Rank:        int(row.Rank),
						Image:       row.AvatarUrl,
					}
				}

				cachedEntries[teamID] = entries

				// Cache the leaderboard data (without tags)
				c.Set(cache.TeamMemberLeaderboardKey(teamID), entries)
			}
		}

		// Build results with user-specific tags
		results := make([]*dataloader.Result[[]model.LeaderboardEntry], len(teamIDs))
		for i, teamID := range teamIDs {
			entries := cachedEntries[teamID]
			if entries == nil {
				results[i] = &dataloader.Result[[]model.LeaderboardEntry]{Data: []model.LeaderboardEntry{}}
				continue
			}

			// Get or compute tags for this user (includes TEAM_LEAD tag)
			tags := getOrComputeTags(ctx, db, c, teamID, currentUserID, entries)

			// Convert cached entries to model with tags
			modelEntries := make([]model.LeaderboardEntry, len(entries))
			for j, entry := range entries {
				entryTags := tags[entry.ID]
				if entryTags == nil {
					entryTags = []model.LeaderboardEntryTag{}
				}

				modelEntries[j] = model.LeaderboardEntry{
					ID:          entry.ID,
					Name:        entry.Name,
					Description: entry.Description,
					Score:       entry.Score,
					Rank:        entry.Rank,
					Tags:        entryTags,
					Image:       entry.Image,
				}
			}

			results[i] = &dataloader.Result[[]model.LeaderboardEntry]{Data: modelEntries}
		}

		return results
	}
}

// getOrComputeTags retrieves or computes tags for a specific user viewing the leaderboard.
// This includes ME tag and TEAM_LEAD tag (with batched role loading).
func getOrComputeTags(ctx context.Context, db *database.DB, c *cache.CacheWithRegistry, teamID, currentUserID string, entries []cachedLeaderboardEntry) cachedLeaderboardTags {
	// Check tag cache for this user
	tagCacheKey := cache.TeamMemberLeaderboardTagsKey(teamID, currentUserID)
	if cached, ok := c.Get(tagCacheKey); ok {
		if tags, ok := cached.(cachedLeaderboardTags); ok {
			return tags
		}
	}

	// Collect all entry IDs for batched role loading
	entryIDs := make([]string, len(entries))
	for i, entry := range entries {
		entryIDs[i] = entry.ID
	}

	// Batch load roles for all entries in ONE query
	entryRoles := make(map[string][]*model.UserRole)
	if len(entryIDs) > 0 && db != nil {
		rows, err := db.Queries.GetAllRolesForUsers(ctx, entryIDs)
		if err == nil {
			// Group roles by user ID
			for _, row := range rows {
				role := convertToUserRole(row)
				entryRoles[row.UserID] = append(entryRoles[row.UserID], role)
			}
		}
	}

	// Compute tags for this user
	tags := make(cachedLeaderboardTags)
	for _, entry := range entries {
		entryTags := []model.LeaderboardEntryTag{}

		// ME tag - simple ID comparison
		if entry.ID == currentUserID {
			entryTags = append(entryTags, model.LeaderboardEntryTagMe)
		}

		// TEAM_LEAD tag - check if entry is team lead of THIS team
		for _, role := range entryRoles[entry.ID] {
			if role.Role == model.RoleTypeTeamLead && role.Scope != nil {
				if role.Scope.Type == model.ScopeTypeTeam && role.Scope.ID == teamID {
					entryTags = append(entryTags, model.LeaderboardEntryTagTeamLead)
					break
				}
			}
		}

		if len(entryTags) > 0 {
			tags[entry.ID] = entryTags
		}
	}

	// Cache tags for this user
	c.Set(tagCacheKey, tags)

	return tags
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
