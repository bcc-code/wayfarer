package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// churchBatchFunc batches church loading by IDs
func churchBatchFunc(db *database.DB) func(context.Context, []string) []*dataloader.Result[*model.Church] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.Church] {
		// Query database once for all IDs using sqlc
		rows, err := db.Queries.GetChurchesByIDs(ctx, ids)
		if err != nil {
			// Return error for all IDs
			results := make([]*dataloader.Result[*model.Church], len(ids))
			for i := range results {
				results[i] = &dataloader.Result[*model.Church]{Error: err}
			}
			return results
		}

		// Create map for O(1) lookup and convert to GraphQL model
		churchMap := make(map[string]*model.Church)
		for _, row := range rows {
			churchMap[row.ID] = &model.Church{
				ID:       row.ID,
				Name:     row.Name,
				Country:  row.Country,
				Category: model.ChurchCategory(row.Category),
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.Church], len(ids))
		for i, id := range ids {
			if church, ok := churchMap[id]; ok {
				results[i] = &dataloader.Result[*model.Church]{Data: church}
			} else {
				results[i] = &dataloader.Result[*model.Church]{
					Error: fmt.Errorf("church not found: %s", id),
				}
			}
		}
		return results
	}
}
