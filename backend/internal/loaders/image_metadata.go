package loaders

import (
	"context"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// imageMetadataByURLBatchFunc batches image metadata loading by URLs
func imageMetadataByURLBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Image] {
	return func(ctx context.Context, urls []string) []*dataloader.Result[*model.Image] {
		// Check cache first for each URL
		imageMap := make(map[string]*model.Image)
		missingURLs := []string{}

		for _, url := range urls {
			cacheKey := cache.ImageMetadataByURLKey(url)
			if cached, ok := c.Get(cacheKey); ok {
				if image, ok := cached.(*model.Image); ok {
					imageMap[url] = image
					continue
				}
			}
			missingURLs = append(missingURLs, url)
		}

		// Query database only for cache misses
		if len(missingURLs) > 0 {
			rows, err := db.Queries.GetFileUploadsByURLs(ctx, missingURLs)
			if err != nil {
				// Don't fail completely - just return images with URLs only for missing ones
				// This handles cases where the image wasn't uploaded through our system
				for _, url := range missingURLs {
					imageMap[url] = &model.Image{URL: url}
				}
			} else {
				// Track which URLs were found in the database
				foundURLs := make(map[string]bool)

				// Convert to GraphQL model and populate cache
				for _, row := range rows {
					var width, height *int
					if row.Width != nil {
						w := int(*row.Width)
						width = &w
					}
					if row.Height != nil {
						h := int(*row.Height)
						height = &h
					}

					image := &model.Image{
						URL:      row.PublicUrl,
						Width:    width,
						Height:   height,
						Blurhash: row.Blurhash,
					}

					imageMap[row.PublicUrl] = image
					foundURLs[row.PublicUrl] = true
					// Image metadata is stable, use longer TTL (1 hour)
					c.SetWithTTL(cache.ImageMetadataByURLKey(row.PublicUrl), image, 1*time.Hour)
				}

				// For URLs not found in file_uploads, create Image with just the URL
				for _, url := range missingURLs {
					if !foundURLs[url] {
						image := &model.Image{URL: url}
						imageMap[url] = image
						// Cache the "not found" result for a shorter time (5 minutes)
						c.SetWithTTL(cache.ImageMetadataByURLKey(url), image, 5*time.Minute)
					}
				}
			}
		}

		// Return results in same order as input URLs
		results := make([]*dataloader.Result[*model.Image], len(urls))
		for i, url := range urls {
			if image, ok := imageMap[url]; ok {
				results[i] = &dataloader.Result[*model.Image]{Data: image}
			} else {
				// This shouldn't happen, but handle it gracefully
				results[i] = &dataloader.Result[*model.Image]{Data: &model.Image{URL: url}}
			}
		}
		return results
	}
}
