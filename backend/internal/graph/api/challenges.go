package api

import (
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/jackc/pgx/v5/pgtype"
)

// buildChallengeFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildChallengeFilterParamsCursor(filter *model.ChallengeFilter, first *int, after *string, last *int, before *string) (sqlc.GetChallengesFilteredCursorParams, error) {
	params := sqlc.GetChallengesFilteredCursorParams{}

	// Apply filters if provided
	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.EventID != nil {
			params.Eventid = *filter.EventID
		}

		if filter.PublishedAfter != nil {
			params.Publishedafter = pgtype.Timestamptz{
				Time:  filter.PublishedAfter.Time,
				Valid: true,
			}
		}

		if filter.PublishedBefore != nil {
			params.Publishedbefore = pgtype.Timestamptz{
				Time:  filter.PublishedBefore.Time,
				Valid: true,
			}
		}
	}

	// Handle cursor pagination
	isBackward := false
	var limit int

	if first != nil && last != nil {
		return params, fmt.Errorf("cannot specify both first and last")
	}

	if first != nil {
		limit = *first + 1 // Fetch one extra to determine hasNextPage
		isBackward = false
	} else if last != nil {
		limit = *last + 1 // Fetch one extra to determine hasPreviousPage
		isBackward = true
	} else {
		// Default page size
		limit = 11 // 10 items + 1 to check for next page
		isBackward = false
	}

	params.Querylimit = int32(limit)
	params.Isbackward = isBackward

	// Set cursors
	if after != nil && *after != "" {
		params.Aftercursor = *after
	}

	if before != nil && *before != "" {
		params.Beforecursor = *before
	}

	return params, nil
}

// buildCountChallengesFilterParams converts GraphQL filter to database count query parameters
func buildCountChallengesFilterParams(filter *model.ChallengeFilter) sqlc.CountChallengesFilteredParams {
	params := sqlc.CountChallengesFilteredParams{}

	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.EventID != nil {
			params.Eventid = *filter.EventID
		}

		if filter.PublishedAfter != nil {
			params.Publishedafter = pgtype.Timestamptz{
				Time:  filter.PublishedAfter.Time,
				Valid: true,
			}
		}

		if filter.PublishedBefore != nil {
			params.Publishedbefore = pgtype.Timestamptz{
				Time:  filter.PublishedBefore.Time,
				Valid: true,
			}
		}
	}

	return params
}

// buildChallengeCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildChallengeCacheKeyParams(filter *model.ChallengeFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if len(filter.Ids) > 0 {
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
		}
		if filter.ProjectID != nil {
			params["projectid"] = *filter.ProjectID
		}
		if filter.EventID != nil {
			params["eventid"] = *filter.EventID
		}
		if filter.PublishedAfter != nil {
			params["publishedafter"] = filter.PublishedAfter.Format(time.RFC3339)
		}
		if filter.PublishedBefore != nil {
			params["publishedbefore"] = filter.PublishedBefore.Format(time.RFC3339)
		}
	}

	// Add pagination parameters
	if first != nil {
		params["first"] = fmt.Sprintf("%d", *first)
	}
	if after != nil && *after != "" {
		params["after"] = *after
	}
	if last != nil {
		params["last"] = fmt.Sprintf("%d", *last)
	}
	if before != nil && *before != "" {
		params["before"] = *before
	}

	return params
}

// convertRowToChallenge converts a database row to a Challenge model
func convertRowToChallenge(row *sqlc.Challenge) *model.Challenge {
	challenge := &model.Challenge{
		ID:          row.ID,
		Name:        row.Name,
		Description: scalars.HTML(row.Description),
		Image:       row.ImageUrl,
		URL:         row.Url,
		ButtonText:  row.ButtonText,
		ProjectID:   row.ProjectID,
		EventID:     row.EventID,
	}

	if row.PublishedAt.Valid {
		challenge.PublishedAt = scalars.DateTime{Time: row.PublishedAt.Time}
	}

	if row.EndTime.Valid {
		endTime := scalars.DateTime{Time: row.EndTime.Time}
		challenge.EndTime = &endTime
	}

	return challenge
}
