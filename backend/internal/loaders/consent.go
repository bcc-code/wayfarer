package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// consentByIDBatchFunc batches loading consents by IDs
func consentByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Consent] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.Consent] {
		// Check cache first for each ID
		consentMap := make(map[string]*model.Consent)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.ConsentKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if consent, ok := cached.(*model.Consent); ok {
					consentMap[id] = consent
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetConsentsByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.Consent], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.Consent]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				var publishedAt *scalars.DateTime
				if row.PublishedAt.Valid {
					dt := scalars.DateTime{Time: row.PublishedAt.Time}
					publishedAt = &dt
				}

				consent := &model.Consent{
					ID:          row.ID,
					Key:         row.Key,
					Version:     int(row.Version),
					Title:       row.Title,
					BodyMarkdown: row.Body,
					PublishedAt: publishedAt,
				}

				consentMap[row.ID] = consent
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.ConsentKey(row.ID), consent)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.Consent], len(ids))
		for i, id := range ids {
			if consent, ok := consentMap[id]; ok {
				results[i] = &dataloader.Result[*model.Consent]{Data: consent}
			} else {
				results[i] = &dataloader.Result[*model.Consent]{
					Error: fmt.Errorf("consent not found: %s", id),
				}
			}
		}
		return results
	}
}
