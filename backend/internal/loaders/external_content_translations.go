package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// externalContentTranslationsBatchFunc batches loading translations for external content
func externalContentTranslationsBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.ExternalContentTranslation] {
	return func(ctx context.Context, externalContentIDs []string) []*dataloader.Result[[]model.ExternalContentTranslation] {
		// Check cache first for each external content ID
		translationsMap := make(map[string][]model.ExternalContentTranslation)
		missingIDs := []string{}

		for _, ecID := range externalContentIDs {
			cacheKey := cache.ExternalContentTranslationsKey(ecID)
			if cached, ok := c.Get(cacheKey); ok {
				if translations, ok := cached.([]model.ExternalContentTranslation); ok {
					translationsMap[ecID] = translations
					continue
				}
			}
			missingIDs = append(missingIDs, ecID)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetExternalContentTranslationsByContentIDs(ctx, missingIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]model.ExternalContentTranslation], len(externalContentIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]model.ExternalContentTranslation]{Error: err}
				}
				return results
			}

			// Group translations by external content ID
			for _, row := range rows {
				translation := model.ExternalContentTranslation{
					LanguageCode: row.LanguageCode,
					Title:        row.Title,
				}
				translationsMap[row.ExternalContentID] = append(translationsMap[row.ExternalContentID], translation)
			}

			// Cache results, including empty slices for negative caching
			for _, ecID := range missingIDs {
				translations := translationsMap[ecID]
				if translations == nil {
					translations = []model.ExternalContentTranslation{}
				}
				translationsMap[ecID] = translations
				c.Set(cache.ExternalContentTranslationsKey(ecID), translations)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]model.ExternalContentTranslation], len(externalContentIDs))
		for i, ecID := range externalContentIDs {
			translations := translationsMap[ecID]
			if translations == nil {
				translations = []model.ExternalContentTranslation{}
			}
			results[i] = &dataloader.Result[[]model.ExternalContentTranslation]{Data: translations}
		}
		return results
	}
}
