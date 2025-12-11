package loaders

import (
	"context"
	"fmt"
	"strings"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// externalContentByIDBatchFunc batches loading external content by IDs
func externalContentByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.ExternalContent] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.ExternalContent] {
		// Check cache first for each ID
		contentMap := make(map[string]*model.ExternalContent)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.ExternalContentKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if content, ok := cached.(*model.ExternalContent); ok {
					contentMap[id] = content
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetExternalContentByIDs(ctx, missingIDs)
			if err != nil {
				results := make([]*dataloader.Result[*model.ExternalContent], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.ExternalContent]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and cache
			for _, row := range rows {
				content := &model.ExternalContent{
					ID:          row.ID,
					PlanID:      row.PlanID,
					TaskID:      row.TaskID,
					ContentType: model.ExternalContentType(strings.ToUpper(row.ContentType)),
					Source:      row.Source,
					SyncedAt:    scalars.DateTime{Time: row.SyncedAt.Time},
					CreatedAt:   scalars.DateTime{Time: row.CreatedAt.Time},
					UpdatedAt:   scalars.DateTime{Time: row.UpdatedAt.Time},
				}

				// Handle nullable fields
				if row.ContentID != nil {
					content.ContentID = row.ContentID
				}
				if row.PublishedAt.Valid {
					content.PublishedAt = &scalars.DateTime{Time: row.PublishedAt.Time}
				}

				contentMap[row.ID] = content
				c.Set(cache.ExternalContentKey(row.ID), content)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[*model.ExternalContent], len(ids))
		for i, id := range ids {
			if content, ok := contentMap[id]; ok {
				results[i] = &dataloader.Result[*model.ExternalContent]{Data: content}
			} else {
				results[i] = &dataloader.Result[*model.ExternalContent]{
					Error: fmt.Errorf("external content not found: %s", id),
				}
			}
		}
		return results
	}
}
