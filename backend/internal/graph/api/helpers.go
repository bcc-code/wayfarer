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
