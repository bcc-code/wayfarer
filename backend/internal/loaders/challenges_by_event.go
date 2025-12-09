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

// challengesByEventBatchFunc batches loading challenges by event IDs
func challengesByEventBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.Challenge] {
	return func(ctx context.Context, eventIDs []string) []*dataloader.Result[[]model.Challenge] {
		// Check cache first for each event ID
		challengesByEvent := make(map[string][]model.Challenge)
		missingEventIDs := []string{}

		for _, eventID := range eventIDs {
			cacheKey := cache.ChallengesByEventKey(eventID)
			if cached, ok := c.Get(cacheKey); ok {
				if challenges, ok := cached.([]model.Challenge); ok {
					challengesByEvent[eventID] = challenges
					continue
				}
			}
			missingEventIDs = append(missingEventIDs, eventID)
		}

		// Query database only for cache misses
		if len(missingEventIDs) > 0 {
			rows, err := db.Queries.GetChallengesByEventIDs(ctx, missingEventIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]model.Challenge], len(eventIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]model.Challenge]{Error: err}
				}
				return results
			}

			// Group challenges by event ID and convert to GraphQL model
			for _, row := range rows {
				challenge := convertEventChallengeRow(row)
				challengesByEvent[*row.EventID] = append(challengesByEvent[*row.EventID], challenge)
			}

			// Populate cache for each event, including empty results
			for _, eventID := range missingEventIDs {
				challenges := challengesByEvent[eventID]
				if challenges == nil {
					challenges = []model.Challenge{} // Empty slice, not nil
				}
				challengesByEvent[eventID] = challenges
				c.Set(cache.ChallengesByEventKey(eventID), challenges)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]model.Challenge], len(eventIDs))
		for i, eventID := range eventIDs {
			challenges := challengesByEvent[eventID]
			if challenges == nil {
				challenges = []model.Challenge{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]model.Challenge]{Data: challenges}
		}
		return results
	}
}

// convertEventChallengeRow converts a GetChallengesByEventIDsRow to the appropriate Challenge implementation
func convertEventChallengeRow(row *sqlc.GetChallengesByEventIDsRow) model.Challenge {
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
