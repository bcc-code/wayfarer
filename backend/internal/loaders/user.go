package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// userByIDBatchFunc batches user loading by IDs
func userByIDBatchFunc(db *database.DB) func(context.Context, []string) []*dataloader.Result[*model.User] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.User] {
		// Query database once for all IDs using sqlc
		rows, err := db.Queries.GetUsersByIDs(ctx, ids)
		if err != nil {
			// Return error for all IDs
			results := make([]*dataloader.Result[*model.User], len(ids))
			for i := range results {
				results[i] = &dataloader.Result[*model.User]{Error: err}
			}
			return results
		}

		// Create map for O(1) lookup and convert to GraphQL model
		userMap := make(map[string]*model.User)
		for _, row := range rows {
			userMap[row.ID] = &model.User{
				ID:        row.ID,
				MembersID: row.MembersID,
				Gender:    model.Gender(row.Gender),
				ChurchID:  row.ChurchID,
				Age:       int(row.Age),
				Email:     row.Email,
				Name:      row.Name,
				Image:     row.AvatarUrl,
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.User], len(ids))
		for i, id := range ids {
			if user, ok := userMap[id]; ok {
				results[i] = &dataloader.Result[*model.User]{Data: user}
			} else {
				results[i] = &dataloader.Result[*model.User]{
					Error: fmt.Errorf("user not found: %s", id),
				}
			}
		}
		return results
	}
}
