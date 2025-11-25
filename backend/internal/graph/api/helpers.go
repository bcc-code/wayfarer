package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
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

// buildLeaderboardConnection builds a GraphQL connection from leaderboard entries
func buildLeaderboardConnection(
	entries []services.LeaderboardEntry,
	meEntry *services.LeaderboardEntry,
	totalCount int,
	currentUserID string,
	first *int,
	last *int,
	after *string,
	before *string,
) *model.LeaderboardConnection {
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
		// Set isMe flag based on entity ID matching current user's entity
		isMe := (meEntry != nil && entry.EntityID == meEntry.EntityID)

		edges[i] = model.LeaderboardEdge{
			Cursor: fmt.Sprintf("%d", entry.Rank),
			Node: &model.LeaderboardEntry{
				ID:          entry.EntityID,
				Name:        entry.Name,
				Description: entry.Name, // TODO: use church name here
				Score:       entry.Score,
				Rank:        int(entry.Rank),
				IsMe:        isMe,
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
		me = &model.LeaderboardEntry{
			ID:    meEntry.EntityID,
			Name:  meEntry.Name,
			Score: meEntry.Score,
			Rank:  int(meEntry.Rank),
			IsMe:  true,
			Image: meEntry.Image,
		}
	}

	return &model.LeaderboardConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
		Me:         me,
	}
}
