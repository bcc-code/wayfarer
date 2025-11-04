package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// rolesByUserBatchFunc batches loading roles by user IDs
func rolesByUserBatchFunc(db *database.DB) func(context.Context, []string) []*dataloader.Result[[]*model.UserRole] {
	return func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.UserRole] {
		// Query all user roles for these users using sqlc
		rows, err := db.Queries.GetAllRolesForUsers(ctx, userIDs)
		if err != nil {
			results := make([]*dataloader.Result[[]*model.UserRole], len(userIDs))
			for i := range results {
				results[i] = &dataloader.Result[[]*model.UserRole]{Error: err}
			}
			return results
		}

		// Group roles by user ID and convert to GraphQL model
		rolesByUser := make(map[string][]*model.UserRole)
		for _, row := range rows {
			// Determine scope type and ID
			var scope *model.RoleScope
			if row.ChurchID != nil {
				scope = &model.RoleScope{
					Type: model.ScopeTypeChurch,
					ID:   *row.ChurchID,
				}
			} else if row.ProjectID != nil {
				scope = &model.RoleScope{
					Type: model.ScopeTypeProject,
					ID:   *row.ProjectID,
				}
			} else if row.TeamID != nil {
				scope = &model.RoleScope{
					Type: model.ScopeTypeTeam,
					ID:   *row.TeamID,
				}
			}

			role := &model.UserRole{
				ID:   row.ID,
				Role: model.RoleType(row.Role),
				// Store partial User objects with just IDs - resolvers will load full data
				User:       &model.User{ID: row.UserID},
				Scope:      scope,
				AssignedBy: &model.User{ID: row.AssignedBy},
				AssignedAt: scalars.DateTime{Time: row.AssignedAt.Time},
			}
			rolesByUser[row.UserID] = append(rolesByUser[row.UserID], role)
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.UserRole], len(userIDs))
		for i, userID := range userIDs {
			roles := rolesByUser[userID]
			if roles == nil {
				roles = []*model.UserRole{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.UserRole]{Data: roles}
		}
		return results
	}
}
