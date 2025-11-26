package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
)

// resolveProjectByID is a helper function to load a project by ID using the dataloader
func resolveProjectByID(ctx context.Context, r *Resolver, projectID string) (*model.Project, error) {
	thunk := r.Loaders.ProjectByIDLoader.Load(ctx, projectID)
	project, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}
	return project, nil
}

// resolveEventByID is a helper function to load an event by ID using the dataloader
func resolveEventByID(ctx context.Context, r *Resolver, eventID *string) (*model.Event, error) {
	if eventID == nil {
		return nil, nil
	}
	thunk := r.Loaders.EventByIDLoader.Load(ctx, *eventID)
	event, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load event: %w", err)
	}
	return event, nil
}

// resolveChallengeByID is a helper function to load a challenge by ID using the dataloader
func resolveChallengeByID(ctx context.Context, r *Resolver, challengeID *string) (*model.Challenge, error) {
	if challengeID == nil {
		return nil, nil
	}
	thunk := r.Loaders.ChallengeByIDLoader.Load(ctx, *challengeID)
	challenge, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load challenge: %w", err)
	}
	return challenge, nil
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
		}
	} else if entityType == model.LeaderboardEntityTypePersons {
		// For person leaderboards, always return a "me" entry even if not ranked
		// Fetch user info for the default entry
		userThunk := ldrs.UserByIDLoader.Load(ctx, currentUserID)
		user, err := userThunk()
		if err != nil {
			return nil, fmt.Errorf("failed to load user for me entry: %w", err)
		}
		if user != nil {
			var description string
			if user.Church != nil {
				description = user.Church.Name
			}
			me = &model.LeaderboardEntry{
				ID:          currentUserID,
				Name:        user.Name,
				Description: description,
				Score:       0,
				Rank:        nil,
				Tags:        []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe},
				Image:       user.Image,
			}
		}
	}

	return &model.LeaderboardConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
		Me:         me,
	}, nil
}
