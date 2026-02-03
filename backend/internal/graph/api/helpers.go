package api

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
)

// timeToDateTime converts a *time.Time to *scalars.DateTime
func timeToDateTime(t *time.Time) *scalars.DateTime {
	if t == nil {
		return nil
	}
	return &scalars.DateTime{Time: *t}
}

// resolveProjectByID is a helper function to load a project by ID using the dataloader
// and applies translations for the requested language
func resolveProjectByID(ctx context.Context, r *Resolver, projectID string) (*model.Project, error) {
	return r.LoadProjectWithTranslation(ctx, projectID)
}

// resolveEventByID is a helper function to load an event by ID using the dataloader
// and applies translations for the requested language
func resolveEventByID(ctx context.Context, r *Resolver, eventID *string) (*model.Event, error) {
	if eventID == nil {
		return nil, nil
	}
	return r.LoadEventWithTranslation(ctx, *eventID)
}

// resolveChallengeByID is a helper function to load a challenge by ID using the dataloader
// and applies translations for the requested language
func resolveChallengeByID(ctx context.Context, r *Resolver, challengeID *string) (model.Challenge, error) {
	if challengeID == nil {
		return nil, nil
	}
	return r.LoadChallengeWithTranslation(ctx, *challengeID)
}

// resolveAchievedAt is a helper function to load the achievedAt timestamp for an achievement
// for the current user using the dataloader
func resolveAchievedAt(ctx context.Context, r *Resolver, achievementID string) (*scalars.DateTime, error) {
	currentUserID, ok := middleware.GetUserID(ctx)
	if !ok || currentUserID == "" {
		return nil, nil // Return nil for unauthenticated users
	}

	key := loaders.UserAchievementKey{UserID: currentUserID, AchievementID: achievementID}
	thunk := r.Loaders.UserAchievementTimestampLoader.Load(ctx, key)
	ts, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load achievement timestamp: %w", err)
	}

	if ts == nil {
		return nil, nil // User hasn't achieved this
	}

	return &scalars.DateTime{Time: *ts}, nil
}

// resolveCelebratedAt is a helper function to load the celebratedAt timestamp for an achievement
// for the current user using the dataloader
func resolveCelebratedAt(ctx context.Context, r *Resolver, achievementID string) (*scalars.DateTime, error) {
	currentUserID, ok := middleware.GetUserID(ctx)
	if !ok || currentUserID == "" {
		return nil, nil // Return nil for unauthenticated users
	}

	key := loaders.UserAchievementKey{UserID: currentUserID, AchievementID: achievementID}
	thunk := r.Loaders.UserAchievementCelebratedTimestampLoader.Load(ctx, key)
	ts, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load achievement celebrated timestamp: %w", err)
	}

	if ts == nil {
		return nil, nil // User hasn't celebrated this
	}

	return &scalars.DateTime{Time: *ts}, nil
}

// viewerContext holds pre-loaded viewer data to avoid N+1 queries in tag computation
type viewerContext struct {
	teamIDs      map[string]bool // Viewer's team IDs in the current project
	superTeamIDs map[string]bool // Viewer's superteam IDs in the current project
	churchID     string          // Viewer's church ID
}

// computeLeaderboardTags computes tags for a leaderboard entry based on entity type and viewer context.
// This function does NO database lookups - all data is pre-loaded into viewerCtx.
// Note: ADMIN and TEAM_LEAD tags have been removed from project/event leaderboards.
// TEAM_LEAD is only shown on team member leaderboards (handled separately).
func computeLeaderboardTags(
	entityType model.LeaderboardEntityType,
	entityID string,
	currentUserID string,
	viewerCtx *viewerContext,
) []model.LeaderboardEntryTag {
	tags := []model.LeaderboardEntryTag{}

	switch entityType {
	case model.LeaderboardEntityTypePersons:
		// ME tag - simple ID comparison
		if entityID == currentUserID {
			tags = append(tags, model.LeaderboardEntryTagMe)
		}

	case model.LeaderboardEntityTypeTeams:
		// ME tag - check if this is one of viewer's teams
		if viewerCtx.teamIDs[entityID] {
			tags = append(tags, model.LeaderboardEntryTagMe)
		}

	case model.LeaderboardEntityTypeSuperteams:
		// ME tag - check if this is viewer's superteam
		if viewerCtx.superTeamIDs[entityID] {
			tags = append(tags, model.LeaderboardEntryTagMe)
		}

	case model.LeaderboardEntityTypeChurches:
		// ME tag - check if this is viewer's church
		if entityID == viewerCtx.churchID {
			tags = append(tags, model.LeaderboardEntryTagMe)
		}
	}

	return tags
}

// preloadViewerContext loads all viewer-specific data needed for tag computation.
// This is called ONCE before iterating over leaderboard entries.
func preloadViewerContext(
	ctx context.Context,
	currentUserID string,
	projectID string,
	entityType model.LeaderboardEntityType,
	ldrs *loaders.Loaders,
) (*viewerContext, error) {
	viewerCtx := &viewerContext{
		teamIDs:      make(map[string]bool),
		superTeamIDs: make(map[string]bool),
	}

	// Load viewer's teams for TEAMS and SUPERTEAMS entity types
	if entityType == model.LeaderboardEntityTypeTeams || entityType == model.LeaderboardEntityTypeSuperteams {
		teamsThunk := ldrs.TeamsByUserLoader.Load(ctx, currentUserID)
		teams, err := teamsThunk()
		if err != nil {
			return viewerCtx, fmt.Errorf("failed to load viewer teams: %w", err)
		}
		for _, team := range teams {
			if team.ProjectID == projectID {
				viewerCtx.teamIDs[team.ID] = true
				if team.SuperTeam != nil {
					viewerCtx.superTeamIDs[team.SuperTeam.ID] = true
				}
			}
		}
	}

	// Load viewer's church for CHURCHES entity type
	if entityType == model.LeaderboardEntityTypeChurches {
		userThunk := ldrs.UserByIDLoader.Load(ctx, currentUserID)
		user, err := userThunk()
		if err != nil {
			return viewerCtx, fmt.Errorf("failed to load viewer: %w", err)
		}
		if user != nil && user.Church != nil {
			viewerCtx.churchID = user.Church.ID
		}
	}

	return viewerCtx, nil
}

// buildLeaderboardConnection builds a GraphQL connection from leaderboard entries
func buildLeaderboardConnection(
	ctx context.Context,
	entries []services.LeaderboardEntry,
	meEntry *services.LeaderboardEntry,
	totalCount int,
	currentUserID string,
	entityType model.LeaderboardEntityType,
	projectID string,
	ldrs *loaders.Loaders,
	first *int,
	last *int,
	after *string,
	before *string,
) (*model.LeaderboardConnection, error) {
	// Determine if there are more entries
	hasMore := false
	requestedLimit := 10
	if first != nil {
		requestedLimit = *first
		hasMore = len(entries) > requestedLimit
	} else if last != nil {
		requestedLimit = *last
		hasMore = len(entries) > requestedLimit
	}

	// Trim to requested limit
	if hasMore {
		entries = entries[:requestedLimit]
	}

	// Pre-load viewer context ONCE (outside the loop) to avoid N+1 queries
	viewerCtx, err := preloadViewerContext(ctx, currentUserID, projectID, entityType, ldrs)
	if err != nil {
		return nil, fmt.Errorf("failed to preload viewer context: %w", err)
	}

	// Build edges - NO dataloader calls in this loop
	edges := make([]model.LeaderboardEdge, len(entries))
	for i, entry := range entries {
		// Compute tags using pre-loaded viewer context (no DB lookups)
		tags := computeLeaderboardTags(entityType, entry.EntityID, currentUserID, viewerCtx)

		rank := int(entry.Rank)
		edges[i] = model.LeaderboardEdge{
			Cursor: fmt.Sprintf("%d", entry.Rank),
			Node: &model.LeaderboardEntry{
				ID:          entry.EntityID,
				Name:        entry.Name,
				Description: entry.Description,
				Score:       entry.Score,
				Rank:        &rank,
				Tags:        tags,
				Image:       entry.Image,
				LastScoreAt: timeToDateTime(entry.LastScoreAt),
			},
		}
	}

	// Build page info
	var startCursor, endCursor *string
	if len(edges) > 0 {
		s := edges[0].Cursor
		startCursor = &s
		e := edges[len(edges)-1].Cursor
		endCursor = &e
	}

	pageInfo := &model.PageInfo{
		HasNextPage:     hasMore && last == nil,
		HasPreviousPage: hasMore && first == nil,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}

	// Build "me" entry using same pre-loaded context
	var me *model.LeaderboardEntry
	if meEntry != nil {
		meTags := computeLeaderboardTags(entityType, meEntry.EntityID, currentUserID, viewerCtx)
		meRank := int(meEntry.Rank)

		me = &model.LeaderboardEntry{
			ID:          meEntry.EntityID,
			Name:        meEntry.Name,
			Description: meEntry.Description,
			Score:       meEntry.Score,
			Rank:        &meRank,
			Tags:        meTags,
			Image:       meEntry.Image,
			LastScoreAt: timeToDateTime(meEntry.LastScoreAt),
		}
	}

	return &model.LeaderboardConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
		Me:         me,
	}, nil
}

// CalculatePersonLeaderboardLimit returns the maximum number of entries a normal user can see
// in a PERSONS type leaderboard, based on the total number of entries in the filtered leaderboard.
// Rules:
//   - totalCount >= 50: show top 20
//   - 20 <= totalCount < 50: show top 10
//   - totalCount < 20: show top 3
func CalculatePersonLeaderboardLimit(totalCount int) int {
	if totalCount >= 50 {
		return 20
	} else if totalCount >= 20 {
		return 10
	}
	return 3
}

// PersonLeaderboardFilterResult contains the result of filtering person leaderboard entries
type PersonLeaderboardFilterResult struct {
	Entries       []services.LeaderboardEntry
	AdjustedFirst *int
}

// FilterPersonLeaderboardEntries filters leaderboard entries for normal users
// to only include entries with rank <= dynamic limit based on totalCount.
// It also adjusts the 'first' pagination parameter accordingly.
func FilterPersonLeaderboardEntries(
	entries []services.LeaderboardEntry,
	totalCount int,
	first *int,
	after *string,
) PersonLeaderboardFilterResult {
	maxLimit := CalculatePersonLeaderboardLimit(totalCount)

	// Filter entries to only include those with rank <= max
	filteredEntries := make([]services.LeaderboardEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Rank <= int64(maxLimit) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	adjustedFirst := first

	// Adjust first param to cap at remaining entries up to max rank
	if after != nil && *after != "" {
		afterRank, err := parseRankCursor(*after)
		if err == nil {
			remaining := maxLimit - int(afterRank)
			if remaining <= 0 {
				// All entries beyond limit, return empty
				return PersonLeaderboardFilterResult{
					Entries:       nil,
					AdjustedFirst: first,
				}
			}
			if first != nil && *first > remaining {
				cappedFirst := remaining
				adjustedFirst = &cappedFirst
			}
		}
	} else {
		// No after cursor, cap first at max
		if first == nil || *first > maxLimit {
			cappedFirst := maxLimit
			adjustedFirst = &cappedFirst
		}
	}

	return PersonLeaderboardFilterResult{
		Entries:       filteredEntries,
		AdjustedFirst: adjustedFirst,
	}
}

// parseRankCursor parses a rank cursor string to int64
func parseRankCursor(cursor string) (int64, error) {
	var rank int64
	_, err := fmt.Sscanf(cursor, "%d", &rank)
	return rank, err
}

// resolveImageByURL loads image metadata by URL using the dataloader.
// If the URL is empty or nil, returns nil. If no metadata is found in file_uploads,
// returns an Image object with just the URL.
func resolveImageByURL(ctx context.Context, ldrs *loaders.Loaders, url *string) (*model.Image, error) {
	if url == nil || *url == "" {
		return nil, nil
	}

	thunk := ldrs.ImageMetadataByURLLoader.Load(ctx, *url)
	image, err := thunk()
	if err != nil {
		// On error, return image with just the URL
		return &model.Image{URL: *url}, nil
	}

	return image, nil
}

// resolveImageByURLNonNullable is like resolveImageByURL but for non-nullable fields.
// It returns an Image with just the URL even if the URL is empty.
func resolveImageByURLNonNullable(ctx context.Context, ldrs *loaders.Loaders, url string) (*model.Image, error) {
	if url == "" {
		return &model.Image{URL: ""}, nil
	}

	thunk := ldrs.ImageMetadataByURLLoader.Load(ctx, url)
	image, err := thunk()
	if err != nil {
		// On error, return image with just the URL
		return &model.Image{URL: url}, nil
	}

	return image, nil
}
