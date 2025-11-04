package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// projectsByUserBatchFunc batches loading projects by user IDs
func projectsByUserBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.Project] {
	return func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.Project] {
		// Check cache first for each user ID
		projectsByUser := make(map[string][]*model.Project)
		missingUserIDs := []string{}

		for _, userID := range userIDs {
			cacheKey := cache.ProjectsByUserKey(userID)
			if cached, ok := c.Get(cacheKey); ok {
				if projects, ok := cached.([]*model.Project); ok {
					projectsByUser[userID] = projects
					continue
				}
			}
			missingUserIDs = append(missingUserIDs, userID)
		}

		// Query database only for cache misses
		if len(missingUserIDs) > 0 {
			rows, err := db.Queries.GetProjectsByUserIDs(ctx, missingUserIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.Project], len(userIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.Project]{Error: err}
				}
				return results
			}

			// Group projects by user ID and convert to GraphQL model
			for _, row := range rows {
				project := &model.Project{
					ID:          row.ID,
					Name:        row.Name,
					Description: row.Description,
					StartDate:   scalars.DateTime{Time: row.StartDate.Time},
					EndDate:     scalars.DateTime{Time: row.EndDate.Time},
					Branding: &model.Branding{
						Logo: row.LogoUrl,
						Colors: &model.Colors{
							Primary:   row.ColorPrimary,
							Secondary: row.ColorSecondary,
							Tertiary:  row.ColorTertiary,
						},
						Rounding: int(row.Rounding),
					},
				}
				projectsByUser[row.UserID] = append(projectsByUser[row.UserID], project)
			}

			// Populate cache for each user, including empty results
			for _, userID := range missingUserIDs {
				projects := projectsByUser[userID]
				if projects == nil {
					projects = []*model.Project{} // Empty slice, not nil
				}
				projectsByUser[userID] = projects
				c.Set(cache.ProjectsByUserKey(userID), projects)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.Project], len(userIDs))
		for i, userID := range userIDs {
			projects := projectsByUser[userID]
			if projects == nil {
				projects = []*model.Project{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.Project]{Data: projects}
		}
		return results
	}
}
