package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services"
)

// getUserTeamInProject returns the user's team in a project, or nil if the user
// is not in any team. The user→team membership is batched per project through
// UserTeamIDInProjectLoader and cached per (user, project), including the
// negative result; team details come from TeamByIDLoader so team edits stay
// visible without touching the membership cache. Membership changes invalidate
// via InvalidateUser (all team mutations call it for affected users).
func (r *Resolver) getUserTeamInProject(ctx context.Context, userID, projectID string) (*model.Team, error) {
	key := loaders.UserProjectKey{UserID: userID, ProjectID: projectID}
	teamID, err := r.Loaders.UserTeamIDInProjectLoader.Load(ctx, key)()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user's team: %w", err)
	}
	if teamID == "" {
		return nil, nil // user not in any team in this project
	}

	team, err := r.Loaders.TeamByIDLoader.Load(ctx, teamID)()
	if err == nil && team != nil {
		return team, nil
	}

	// Cached team ID no longer resolves (e.g. team deleted without membership
	// invalidation); drop the stale entry and retry once from the DB.
	r.Cache.Delete(cache.UserTeamInProjectKey(userID, projectID))
	teamID, err = r.Loaders.UserTeamIDInProjectLoader.Load(ctx, key)()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user's team: %w", err)
	}
	if teamID == "" {
		return nil, nil
	}
	team, err = r.Loaders.TeamByIDLoader.Load(ctx, teamID)()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user's team: %w", err)
	}
	return team, nil
}

// teamUpdateRoleChecker is an interface for checking roles during team updates.
type teamUpdateRoleChecker interface {
	IsAdmin(ctx context.Context, userID string) bool
	CanManageProject(ctx context.Context, userID, projectID string) bool
	LoadUserRoles(ctx context.Context, userID string) ([]*sqlc.UserRole, error)
}

// canModifyAllTeamFields checks whether the user has permission to modify fields
// beyond just the team name (e.g. description, leaderboardExcluded).
// Team leads can only update the team name; admins, project admins, and church admins
// can update all fields.
func canModifyAllTeamFields(ctx context.Context, checker teamUpdateRoleChecker, userID, projectID string) bool {
	if checker.IsAdmin(ctx, userID) {
		return true
	}
	if checker.CanManageProject(ctx, userID, projectID) {
		return true
	}
	if roles, err := checker.LoadUserRoles(ctx, userID); err == nil {
		for _, role := range roles {
			if role.Role == string(services.RoleChurchAdmin) {
				return true
			}
		}
	}
	return false
}

// validateTeamUpdateInput checks whether the given input is permitted for the user's role.
// Team leads may only update the team name. If description or leaderboardExcluded is set,
// the user must have admin, project admin, or church admin privileges.
// Returns nil if the update is allowed, or an error describing the restriction.
func validateTeamUpdateInput(ctx context.Context, checker teamUpdateRoleChecker, userID, projectID string, input model.UpdateTeamInput) error {
	if input.Description != nil || input.LeaderboardExcluded != nil {
		if !canModifyAllTeamFields(ctx, checker, userID, projectID) {
			return fmt.Errorf("unauthorized: team leads can only update team name")
		}
	}
	return nil
}

// buildTeamFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildTeamFilterParamsCursor(filter *model.TeamFilter, first *int, after *string, last *int, before *string) (sqlc.GetTeamsFilteredCursorParams, error) {
	params := sqlc.GetTeamsFilteredCursorParams{}

	// Apply filters if provided
	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.SuperTeamID != nil {
			params.Superteamid = *filter.SuperTeamID
		}

		if filter.NoSuperTeam != nil {
			params.Nosuperteam = *filter.NoSuperTeam
		}

		if filter.MinMembers != nil {
			params.Minmembers = int32(*filter.MinMembers)
		}

		if filter.MaxMembers != nil {
			params.Maxmembers = int32(*filter.MaxMembers)
		}

		if filter.ChurchID != nil {
			params.Churchid = *filter.ChurchID
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

// buildCountTeamsFilterParams converts GraphQL filter to database count query parameters
func buildCountTeamsFilterParams(filter *model.TeamFilter) sqlc.CountTeamsFilteredParams {
	params := sqlc.CountTeamsFilteredParams{}

	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.SuperTeamID != nil {
			params.Superteamid = *filter.SuperTeamID
		}

		if filter.NoSuperTeam != nil {
			params.Nosuperteam = *filter.NoSuperTeam
		}

		if filter.MinMembers != nil {
			params.Minmembers = int32(*filter.MinMembers)
		}

		if filter.MaxMembers != nil {
			params.Maxmembers = int32(*filter.MaxMembers)
		}

		if filter.ChurchID != nil {
			params.Churchid = *filter.ChurchID
		}
	}

	return params
}

// buildTeamCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildTeamCacheKeyParams(filter *model.TeamFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if len(filter.Ids) > 0 {
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
		}
		if filter.ProjectID != nil {
			params["projectid"] = *filter.ProjectID
		}
		if filter.SuperTeamID != nil {
			params["superteamid"] = *filter.SuperTeamID
		}
		if filter.NoSuperTeam != nil {
			params["nosuperteam"] = fmt.Sprintf("%v", *filter.NoSuperTeam)
		}
		if filter.MinMembers != nil {
			params["minmembers"] = fmt.Sprintf("%d", *filter.MinMembers)
		}
		if filter.MaxMembers != nil {
			params["maxmembers"] = fmt.Sprintf("%d", *filter.MaxMembers)
		}
		if filter.ChurchID != nil {
			params["churchid"] = *filter.ChurchID
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
