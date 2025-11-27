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

// projectByIDBatchFunc batches loading projects by IDs
func projectByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Project] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.Project] {
		// Check cache first for each ID
		projectMap := make(map[string]*model.Project)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.ProjectKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if project, ok := cached.(*model.Project); ok {
					projectMap[id] = project
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetProjectsByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.Project], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.Project]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				var logo *string
				if row.LogoUrl != nil {
					logo = row.LogoUrl
				}

				project := &model.Project{
					ID:          row.ID,
					Name:        row.Name,
					Description: row.Description,
					StartDate:   scalars.DateTime{Time: row.StartDate.Time},
					EndDate:     scalars.DateTime{Time: row.EndDate.Time},
					Branding: &model.Branding{
						Logo: logo,
						Colors: &model.Colors{
							Light: &model.ColorSet{
								Accent:            row.ColorLightAccent,
								AccentContrast:    row.ColorLightAccentContrast,
								OnAccent:          row.ColorLightOnAccent,
								BackgroundDefault: row.ColorLightBackgroundDefault,
								BackgroundRaised:  row.ColorLightBackgroundRaised,
								BackgroundIndent:  row.ColorLightBackgroundIndent,
								TextDefault:       row.ColorLightTextDefault,
								TextMuted:         row.ColorLightTextMuted,
								TextHint:          row.ColorLightTextHint,
								ShadowDefault:     row.ColorLightShadowDefault,
								ShadowBlank:       row.ColorLightShadowBlank,
								BorderDefault:     row.ColorLightBorderDefault,
							},
							Dark: &model.ColorSet{
								Accent:            row.ColorDarkAccent,
								AccentContrast:    row.ColorDarkAccentContrast,
								OnAccent:          row.ColorDarkOnAccent,
								BackgroundDefault: row.ColorDarkBackgroundDefault,
								BackgroundRaised:  row.ColorDarkBackgroundRaised,
								BackgroundIndent:  row.ColorDarkBackgroundIndent,
								TextDefault:       row.ColorDarkTextDefault,
								TextMuted:         row.ColorDarkTextMuted,
								TextHint:          row.ColorDarkTextHint,
								ShadowDefault:     row.ColorDarkShadowDefault,
								ShadowBlank:       row.ColorDarkShadowBlank,
								BorderDefault:     row.ColorDarkBorderDefault,
							},
						},
						Rounding: int(row.Rounding),
					},
					ArchivedAt: row.Archived,
				}

				projectMap[row.ID] = project
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.ProjectKey(row.ID), project)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.Project], len(ids))
		for i, id := range ids {
			if project, ok := projectMap[id]; ok {
				results[i] = &dataloader.Result[*model.Project]{Data: project}
			} else {
				results[i] = &dataloader.Result[*model.Project]{
					Error: fmt.Errorf("project not found: %s", id),
				}
			}
		}
		return results
	}
}
