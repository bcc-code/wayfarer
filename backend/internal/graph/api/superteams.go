package api

import (
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// buildSuperTeamFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildSuperTeamFilterParamsCursor(filter *model.SuperTeamFilter, first *int, after *string, last *int, before *string) (sqlc.GetSuperTeamsFilteredCursorParams, error) {
	params := sqlc.GetSuperTeamsFilteredCursorParams{}

	// Apply filters if provided
	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.MinTeams != nil {
			params.Minteams = int32(*filter.MinTeams)
		}

		if filter.MaxTeams != nil {
			params.Maxteams = int32(*filter.MaxTeams)
		}

		if filter.MinMembers != nil {
			params.Minmembers = int32(*filter.MinMembers)
		}

		if filter.MaxMembers != nil {
			params.Maxmembers = int32(*filter.MaxMembers)
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

// buildCountSuperTeamsFilterParams converts GraphQL filter to database count query parameters
func buildCountSuperTeamsFilterParams(filter *model.SuperTeamFilter) sqlc.CountSuperTeamsFilteredParams {
	params := sqlc.CountSuperTeamsFilteredParams{}

	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.MinTeams != nil {
			params.Minteams = int32(*filter.MinTeams)
		}

		if filter.MaxTeams != nil {
			params.Maxteams = int32(*filter.MaxTeams)
		}

		if filter.MinMembers != nil {
			params.Minmembers = int32(*filter.MinMembers)
		}

		if filter.MaxMembers != nil {
			params.Maxmembers = int32(*filter.MaxMembers)
		}
	}

	return params
}

// buildSuperTeamCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildSuperTeamCacheKeyParams(filter *model.SuperTeamFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if len(filter.Ids) > 0 {
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
		}
		if filter.ProjectID != nil {
			params["projectid"] = *filter.ProjectID
		}
		if filter.MinTeams != nil {
			params["minteams"] = fmt.Sprintf("%d", *filter.MinTeams)
		}
		if filter.MaxTeams != nil {
			params["maxteams"] = fmt.Sprintf("%d", *filter.MaxTeams)
		}
		if filter.MinMembers != nil {
			params["minmembers"] = fmt.Sprintf("%d", *filter.MinMembers)
		}
		if filter.MaxMembers != nil {
			params["maxmembers"] = fmt.Sprintf("%d", *filter.MaxMembers)
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
