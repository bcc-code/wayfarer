package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// challengesByProjectBatchFunc batches loading challenges by project IDs
func challengesByProjectBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.Challenge] {
	return func(ctx context.Context, projectIDs []string) []*dataloader.Result[[]model.Challenge] {
		// Check cache first for each project ID
		challengesByProject := make(map[string][]model.Challenge)
		missingProjectIDs := []string{}

		for _, projectID := range projectIDs {
			cacheKey := cache.ChallengesByProjectKey(projectID)
			if cached, ok := c.Get(cacheKey); ok {
				if challenges, ok := cached.([]model.Challenge); ok {
					challengesByProject[projectID] = challenges
					continue
				}
			}
			missingProjectIDs = append(missingProjectIDs, projectID)
		}

		// Query database only for cache misses
		if len(missingProjectIDs) > 0 {
			rows, err := db.Queries.GetChallengesByProjectIDs(ctx, missingProjectIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]model.Challenge], len(projectIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]model.Challenge]{Error: err}
				}
				return results
			}

			// Group challenges by project ID and convert to GraphQL model
			for _, row := range rows {
				challenge := convertProjectChallengeRow(row)
				challengesByProject[row.ProjectID] = append(challengesByProject[row.ProjectID], challenge)
			}

			// Populate cache for each project, including empty results
			for _, projectID := range missingProjectIDs {
				challenges := challengesByProject[projectID]
				if challenges == nil {
					challenges = []model.Challenge{} // Empty slice, not nil
				}
				challengesByProject[projectID] = challenges
				c.Set(cache.ChallengesByProjectKey(projectID), challenges)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]model.Challenge], len(projectIDs))
		for i, projectID := range projectIDs {
			challenges := challengesByProject[projectID]
			if challenges == nil {
				challenges = []model.Challenge{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]model.Challenge]{Data: challenges}
		}
		return results
	}
}

// convertProjectChallengeRow converts a GetChallengesByProjectIDsRow to the appropriate Challenge implementation
func convertProjectChallengeRow(row *sqlc.GetChallengesByProjectIDsRow) model.Challenge {
	var publishedAt, visibleAt, startedAt, endTime *scalars.DateTime
	if row.PublishedAt.Valid {
		dt := scalars.DateTime{Time: row.PublishedAt.Time}
		publishedAt = &dt
	}
	if row.VisibleAt.Valid {
		dt := scalars.DateTime{Time: row.VisibleAt.Time}
		visibleAt = &dt
	}
	if row.StartedAt.Valid {
		dt := scalars.DateTime{Time: row.StartedAt.Time}
		startedAt = &dt
	}
	if row.EndTime.Valid {
		dt := scalars.DateTime{Time: row.EndTime.Time}
		endTime = &dt
	}

	switch row.ChallengeType {
	case ChallengeTypeQuiz:
		return &model.QuizChallenge{
			ID:                          row.ID,
			Name:                        row.Name,
			Description:                 scalars.HTML(row.Description),
			Image:                       row.ImageUrl,
			ButtonText:                  row.ButtonText,
			ProjectID:                   row.ProjectID,
			EventID:                     row.EventID,
			PublishedAt:                 publishedAt,
			VisibleAt:                   visibleAt,
			StartedAt:                   startedAt,
			EndTime:                     endTime,
			RequiresTeamMembership:      row.RequiresTeamMembership,
			RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		}
	case ChallengeTypeExternal:
		url := ""
		if row.Url != nil {
			url = *row.Url
		}
		return &model.ExternalChallenge{
			ID:                          row.ID,
			Name:                        row.Name,
			Description:                 scalars.HTML(row.Description),
			Image:                       row.ImageUrl,
			URL:                         url,
			ButtonText:                  row.ButtonText,
			ProjectID:                   row.ProjectID,
			EventID:                     row.EventID,
			PublishedAt:                 publishedAt,
			VisibleAt:                   visibleAt,
			StartedAt:                   startedAt,
			EndTime:                     endTime,
			RequiresTeamMembership:      row.RequiresTeamMembership,
			RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		}
	default: // SIMPLE
		return &model.SimpleChallenge{
			ID:                          row.ID,
			Name:                        row.Name,
			Description:                 scalars.HTML(row.Description),
			Image:                       row.ImageUrl,
			ButtonText:                  row.ButtonText,
			ProjectID:                   row.ProjectID,
			EventID:                     row.EventID,
			PublishedAt:                 publishedAt,
			VisibleAt:                   visibleAt,
			StartedAt:                   startedAt,
			EndTime:                     endTime,
			AllowSelfCompletion:         row.AllowSelfCompletion,
			RequiresTeamMembership:      row.RequiresTeamMembership,
			RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		}
	}
}
