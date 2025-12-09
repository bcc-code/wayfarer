package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// ChallengeType constants
const (
	ChallengeTypeSimple   = "SIMPLE"
	ChallengeTypeQuiz     = "QUIZ"
	ChallengeTypeExternal = "EXTERNAL"
)

// challengeByIDBatchFunc batches loading challenges by IDs
func challengeByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[model.Challenge] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[model.Challenge] {
		// Check cache first for each ID
		challengeMap := make(map[string]model.Challenge)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.ChallengeKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if challenge, ok := cached.(model.Challenge); ok {
					challengeMap[id] = challenge
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetChallengesByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[model.Challenge], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[model.Challenge]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				challenge := convertRowToChallenge(row)
				challengeMap[row.ID] = challenge
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.ChallengeKey(row.ID), challenge)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[model.Challenge], len(ids))
		for i, id := range ids {
			if challenge, ok := challengeMap[id]; ok {
				results[i] = &dataloader.Result[model.Challenge]{Data: challenge}
			} else {
				results[i] = &dataloader.Result[model.Challenge]{
					Error: fmt.Errorf("challenge not found: %s", id),
				}
			}
		}
		return results
	}
}

// convertRowToChallenge converts a database row to the appropriate Challenge implementation
func convertRowToChallenge(row *sqlc.GetChallengesByIDsRow) model.Challenge {
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
