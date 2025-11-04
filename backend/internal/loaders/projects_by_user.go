package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// projectsByUserBatchFunc batches loading projects by user IDs
func projectsByUserBatchFunc(db *database.DB) func(context.Context, []string) []*dataloader.Result[[]*model.Project] {
	return func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.Project] {
		// Query all user-project relationships for these users using sqlc
		rows, err := db.Queries.GetProjectsByUserIDs(ctx, userIDs)
		if err != nil {
			results := make([]*dataloader.Result[[]*model.Project], len(userIDs))
			for i := range results {
				results[i] = &dataloader.Result[[]*model.Project]{Error: err}
			}
			return results
		}

		// Group projects by user ID and convert to GraphQL model
		projectsByUser := make(map[string][]*model.Project)
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
