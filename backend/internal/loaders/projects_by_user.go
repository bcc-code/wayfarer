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
				var logo *string
				if row.LogoUrl != nil {
					logo = row.LogoUrl
				}
				var banner *string
				if row.BannerUrl != nil {
					banner = row.BannerUrl
				}

				project := &model.Project{
					ID:                 row.ID,
					Name:               row.Name,
					Description:        row.Description,
					RulesRaw:           row.Rules,
					InfoMessageRaw:     row.InfoMessage,
					InfoMessageStart:   scalars.ToDateTimePointer(row.InfoMessageStart),
					InfoMessageEnd:     scalars.ToDateTimePointer(row.InfoMessageEnd),
					StartDate:          scalars.DateTime{Time: row.StartDate.Time},
					EndDate:            scalars.DateTime{Time: row.EndDate.Time},
					Branding: &model.Branding{
						Logo:   logo,
						Banner: banner,
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
