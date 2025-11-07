package api

import (
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// buildChurchFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildChurchFilterParamsCursor(filter *model.ChurchFilter, first *int, after *string, last *int, before *string) (sqlc.GetChurchesFilteredCursorParams, error) {
	params := sqlc.GetChurchesFilteredCursorParams{}

	// Apply filters if provided
	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.Country != nil {
			params.Country = *filter.Country
		}

		if filter.Category != nil {
			params.Category = string(*filter.Category)
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

// buildCountChurchesFilterParams converts GraphQL filter to database count query parameters
func buildCountChurchesFilterParams(filter *model.ChurchFilter) sqlc.CountChurchesFilteredParams {
	params := sqlc.CountChurchesFilteredParams{}

	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.Country != nil {
			params.Country = *filter.Country
		}

		if filter.Category != nil {
			params.Category = string(*filter.Category)
		}
	}

	return params
}

// buildChurchCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildChurchCacheKeyParams(filter *model.ChurchFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if len(filter.Ids) > 0 {
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
		}
		if filter.Country != nil {
			params["country"] = *filter.Country
		}
		if filter.Category != nil {
			params["category"] = string(*filter.Category)
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
