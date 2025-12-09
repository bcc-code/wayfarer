package api

import (
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/jackc/pgx/v5/pgtype"
)

// buildQuizFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildQuizFilterParamsCursor(filter *model.QuizFilter, first *int, after *string, last *int, before *string) (sqlc.GetQuizzesFilteredCursorParams, error) {
	params := sqlc.GetQuizzesFilteredCursorParams{}

	// Apply filters if provided
	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.ChallengeID != nil {
			params.Challengeid = *filter.ChallengeID
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

// buildCountQuizzesFilterParams converts GraphQL filter to database count query parameters
func buildCountQuizzesFilterParams(filter *model.QuizFilter) sqlc.CountQuizzesFilteredParams {
	params := sqlc.CountQuizzesFilteredParams{}

	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.ChallengeID != nil {
			params.Challengeid = *filter.ChallengeID
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

// buildQuizCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildQuizCacheKeyParams(filter *model.QuizFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if len(filter.Ids) > 0 {
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
		}
		if filter.ProjectID != nil {
			params["projectid"] = *filter.ProjectID
		}
		if filter.ChallengeID != nil {
			params["challengeid"] = *filter.ChallengeID
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
