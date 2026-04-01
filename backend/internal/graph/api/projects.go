package api

import (
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/jackc/pgx/v5/pgtype"
)

// buildProjectFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildProjectFilterParamsCursor(filter *model.ProjectFilter, first *int, after *string, last *int, before *string) (sqlc.GetProjectsFilteredCursorParams, error) {
	params := sqlc.GetProjectsFilteredCursorParams{}

	// Apply filters if provided
	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		params.Archived = filter.Archived

		if filter.StartDateAfter != nil {
			params.Startdateafter = pgtype.Timestamptz{
				Time:  filter.StartDateAfter.Time,
				Valid: true,
			}
		}

		if filter.StartDateBefore != nil {
			params.Startdatebefore = pgtype.Timestamptz{
				Time:  filter.StartDateBefore.Time,
				Valid: true,
			}
		}

		if filter.EndDateAfter != nil {
			params.Enddateafter = pgtype.Timestamptz{
				Time:  filter.EndDateAfter.Time,
				Valid: true,
			}
		}

		if filter.EndDateBefore != nil {
			params.Enddatebefore = pgtype.Timestamptz{
				Time:  filter.EndDateBefore.Time,
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

// buildCountProjectsFilterParams converts GraphQL filter to database count query parameters
func buildCountProjectsFilterParams(filter *model.ProjectFilter) sqlc.CountProjectsFilteredParams {
	params := sqlc.CountProjectsFilteredParams{}

	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		params.Archived = filter.Archived

		if filter.StartDateAfter != nil {
			params.Startdateafter = pgtype.Timestamptz{
				Time:  filter.StartDateAfter.Time,
				Valid: true,
			}
		}

		if filter.StartDateBefore != nil {
			params.Startdatebefore = pgtype.Timestamptz{
				Time:  filter.StartDateBefore.Time,
				Valid: true,
			}
		}

		if filter.EndDateAfter != nil {
			params.Enddateafter = pgtype.Timestamptz{
				Time:  filter.EndDateAfter.Time,
				Valid: true,
			}
		}

		if filter.EndDateBefore != nil {
			params.Enddatebefore = pgtype.Timestamptz{
				Time:  filter.EndDateBefore.Time,
				Valid: true,
			}
		}
	}

	return params
}

// buildProjectCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildProjectCacheKeyParams(filter *model.ProjectFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if len(filter.Ids) > 0 {
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
		}
		if filter.Archived != nil {
			params["archived"] = fmt.Sprintf("%t", *filter.Archived)
		}
		if filter.StartDateAfter != nil {
			params["startdateafter"] = filter.StartDateAfter.Format("2006-01-02T15:04:05Z07:00")
		}
		if filter.StartDateBefore != nil {
			params["startdatebefore"] = filter.StartDateBefore.Format("2006-01-02T15:04:05Z07:00")
		}
		if filter.EndDateAfter != nil {
			params["enddateafter"] = filter.EndDateAfter.Format("2006-01-02T15:04:05Z07:00")
		}
		if filter.EndDateBefore != nil {
			params["enddatebefore"] = filter.EndDateBefore.Format("2006-01-02T15:04:05Z07:00")
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
