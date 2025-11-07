package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
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
