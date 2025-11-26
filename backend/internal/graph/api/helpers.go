package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
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

// computeLeaderboardTags computes tags for a leaderboard entry based on entity type and viewer context
func computeLeaderboardTags(
	ctx context.Context,
	entityType model.LeaderboardEntityType,
	entityID string,
	projectID string,
	currentUserID string,
	ldrs *loaders.Loaders,
) ([]model.LeaderboardEntryTag, error) {
	tags := []model.LeaderboardEntryTag{}

	switch entityType {
	case model.LeaderboardEntityTypePersons:
		// Check ME tag
		if entityID == currentUserID {
			tags = append(tags, model.LeaderboardEntryTagMe)
		}

		// Check ADMIN tag
		userRoles, ok := ctx.Value(middleware.UserRolesKey).([]string)
		if ok {
			for _, role := range userRoles {
				if role == "superadmin" || role == "admin" || role == "project_admin" {
					// Check if the entity (person in leaderboard) has admin roles
					personRolesThunk := ldrs.RolesByUserLoader.Load(ctx, entityID)
					personRoles, err := personRolesThunk()
					if err == nil {
						for _, personRole := range personRoles {
							if personRole.Role == model.RoleTypeSuperadmin ||
								personRole.Role == model.RoleTypeAdmin ||
								personRole.Role == model.RoleTypeProjectAdmin {
								tags = append(tags, model.LeaderboardEntryTagAdmin)
								goto checkTeamLead // Break out of nested loops
							}
						}
					}
				}
			}
		}

	checkTeamLead:
		// Check TEAM_LEAD tag
		// Load viewer's teams
		viewerTeamsThunk := ldrs.TeamsByUserLoader.Load(ctx, currentUserID)
		allViewerTeams, err := viewerTeamsThunk()
		if err != nil {
			return tags, fmt.Errorf("failed to load viewer teams: %w", err)
		}

		// Filter teams by project and get team IDs
		viewerTeamIDs := make(map[string]bool)
		for _, team := range allViewerTeams {
			if team.ProjectID == projectID {
				viewerTeamIDs[team.ID] = true
			}
		}

		// Load person's roles
		personRolesThunk := ldrs.RolesByUserLoader.Load(ctx, entityID)
		personRoles, err := personRolesThunk()
		if err != nil {
			return tags, fmt.Errorf("failed to load person roles: %w", err)
		}

		// Check if person is TEAM_LEAD of any of viewer's teams
		for _, role := range personRoles {
			if role.Role == model.RoleTypeTeamLead && role.Scope != nil {
				if role.Scope.Type == model.ScopeTypeTeam {
					if viewerTeamIDs[role.Scope.ID] {
						tags = append(tags, model.LeaderboardEntryTagTeamLead)
						break
					}
				}
			}
		}

	case model.LeaderboardEntityTypeTeams:
		// For teams, only check ME tag
		// Load viewer's teams
		viewerTeamsThunk := ldrs.TeamsByUserLoader.Load(ctx, currentUserID)
		allViewerTeams, err := viewerTeamsThunk()
		if err != nil {
			return tags, fmt.Errorf("failed to load viewer teams: %w", err)
		}

		// Filter by project and check if entityID matches
		for _, team := range allViewerTeams {
			if team.ProjectID == projectID && team.ID == entityID {
				tags = append(tags, model.LeaderboardEntryTagMe)
				break
			}
		}

	case model.LeaderboardEntityTypeSuperteams:
		// For superteams, only check ME tag
		// Load viewer's teams
		viewerTeamsThunk := ldrs.TeamsByUserLoader.Load(ctx, currentUserID)
		allViewerTeams, err := viewerTeamsThunk()
		if err != nil {
			return tags, fmt.Errorf("failed to load viewer teams: %w", err)
		}

		// Filter by project and check if any team belongs to this superteam
		for _, team := range allViewerTeams {
			if team.ProjectID == projectID && team.SuperTeam != nil && team.SuperTeam.ID == entityID {
				tags = append(tags, model.LeaderboardEntryTagMe)
				break
			}
		}

	case model.LeaderboardEntityTypeChurches:
		// For churches, only check ME tag
		// Load viewer's user data
		viewerThunk := ldrs.UserByIDLoader.Load(ctx, currentUserID)
		viewer, err := viewerThunk()
		if err != nil {
			return tags, fmt.Errorf("failed to load viewer: %w", err)
		}

		if viewer.Church.ID == entityID {
			tags = append(tags, model.LeaderboardEntryTagMe)
		}
	}

	return tags, nil
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

	// Build edges
	edges := make([]model.LeaderboardEdge, len(entries))
	for i, entry := range entries {
		// Compute tags based on entity type and viewer context
		tags, err := computeLeaderboardTags(ctx, entityType, entry.EntityID, projectID, currentUserID, ldrs)
		if err != nil {
			return nil, fmt.Errorf("failed to compute tags for entry %s: %w", entry.EntityID, err)
		}

		edges[i] = model.LeaderboardEdge{
			Cursor: fmt.Sprintf("%d", entry.Rank),
			Node: &model.LeaderboardEntry{
				ID:          entry.EntityID,
				Name:        entry.Name,
				Description: entry.Description,
				Score:       entry.Score,
				Rank:        int(entry.Rank),
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

	// Build "me" entry
	var me *model.LeaderboardEntry
	if meEntry != nil {
		// Compute tags for "me" entry
		meTags, err := computeLeaderboardTags(ctx, entityType, meEntry.EntityID, projectID, currentUserID, ldrs)
		if err != nil {
			return nil, fmt.Errorf("failed to compute tags for me entry: %w", err)
		}

		me = &model.LeaderboardEntry{
			ID:          meEntry.EntityID,
			Name:        meEntry.Name,
			Description: meEntry.Description,
			Score:       meEntry.Score,
			Rank:        int(meEntry.Rank),
			Tags:        meTags,
			Image:       meEntry.Image,
		}
	}

	return &model.LeaderboardConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
		Me:         me,
	}, nil
}
