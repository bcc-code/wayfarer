package api

import (
	"context"
	"fmt"

	"github.com/graph-gophers/dataloader/v7"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
)

// UserLoaderInterface defines the interface for loading users
type UserLoaderInterface interface {
	Load(ctx context.Context, key string) dataloader.Thunk[*model.User]
}

// RoleServiceInterface defines the interface for role and permission checking
type RoleServiceInterface interface {
	IsAdmin(ctx context.Context, userID string) bool
	CanManageChurch(ctx context.Context, userID, churchID string) bool
	CanManageProject(ctx context.Context, userID, projectID string) bool
	CanManageTeam(ctx context.Context, userID, teamID string) bool
}

// authenticatedUserInfo holds information about the authenticated user
type authenticatedUserInfo struct {
	UserID   string
	ChurchID string
}

// validateUserAccess validates that the user is authenticated and retrieves their basic info
func validateUserAccess(ctx context.Context, userLoader UserLoaderInterface) (*authenticatedUserInfo, error) {
	// Get current user ID from context
	currentUserID, ok := middleware.GetUserID(ctx)
	if !ok || currentUserID == "" {
		return nil, fmt.Errorf("user not authenticated")
	}

	// Load current user to get their church ID
	currentUserThunk := userLoader.Load(ctx, currentUserID)
	currentUser, err := currentUserThunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load current user: %w", err)
	}

	return &authenticatedUserInfo{
		UserID:   currentUserID,
		ChurchID: currentUser.ChurchID,
	}, nil
}

// permissionContext holds information about what the user has permission to access
type permissionContext struct {
	IsAdmin        bool
	IsChurchAdmin  bool
	IsProjectAdmin bool
	IsTeamLead     bool
	ChurchID       string
	ProjectID      string
	TeamID         string
}

// checkUserPermissions determines what level of access the user has
func checkUserPermissions(
	ctx context.Context,
	roleService RoleServiceInterface,
	userInfo *authenticatedUserInfo,
	filter *model.UserFilter,
) (*permissionContext, error) {
	perms := &permissionContext{
		ChurchID: userInfo.ChurchID,
	}

	// Check if current user has admin-level permissions
	perms.IsAdmin = roleService.IsAdmin(ctx, userInfo.UserID)

	// Check for other admin roles if not a global admin
	if !perms.IsAdmin {
		// Check if user has church admin role
		perms.IsChurchAdmin = roleService.CanManageChurch(ctx, userInfo.UserID, userInfo.ChurchID)

		// Check if user has project admin role (if projectId filter is provided)
		if filter != nil && filter.ProjectID != nil {
			perms.ProjectID = *filter.ProjectID
			perms.IsProjectAdmin = roleService.CanManageProject(ctx, userInfo.UserID, perms.ProjectID)
		}

		// Check if user has team lead role (if teamId filter is provided)
		if filter != nil && filter.TeamID != nil {
			perms.TeamID = *filter.TeamID
			perms.IsTeamLead = roleService.CanManageTeam(ctx, userInfo.UserID, perms.TeamID)
		}
	}

	// Only allow admin roles to access this query
	if !perms.IsAdmin && !perms.IsChurchAdmin && !perms.IsProjectAdmin && !perms.IsTeamLead {
		return nil, fmt.Errorf("permission denied: insufficient privileges to list users")
	}

	return perms, nil
}

// applyPermissionFilters modifies the filter based on the user's permissions
func applyPermissionFilters(filter *model.UserFilter, perms *permissionContext) *model.UserFilter {
	// Initialize filter if nil
	if filter == nil {
		filter = &model.UserFilter{}
	}

	// Apply role-based filter restrictions
	if !perms.IsAdmin {
		// Church admins can only see users from their church
		if perms.IsChurchAdmin && !perms.IsProjectAdmin && !perms.IsTeamLead {
			filter.ChurchID = &perms.ChurchID
		}

		// Project admins can only see users from their project
		if perms.IsProjectAdmin && filter.ProjectID == nil {
			filter.ProjectID = &perms.ProjectID
		}

		// Team leads can only see users from their team
		if perms.IsTeamLead && !perms.IsProjectAdmin && filter.TeamID == nil {
			filter.TeamID = &perms.TeamID
		}
	}

	return filter
}

// buildUserFilterParams converts GraphQL filter to database query parameters
func buildUserFilterParams(filter *model.UserFilter, limit *int, offset *int) sqlc.GetUsersFilteredParams {
	params := sqlc.GetUsersFilteredParams{}

	// Apply filters
	if filter.ChurchID != nil {
		params.Churchid = *filter.ChurchID
	}
	if filter.Gender != nil {
		params.Gender = string(*filter.Gender)
	}
	if filter.MinAge != nil {
		params.Minage = int32(*filter.MinAge)
	}
	if filter.MaxAge != nil {
		params.Maxage = int32(*filter.MaxAge)
	}
	if filter.ProjectID != nil {
		params.Projectid = *filter.ProjectID
	}
	if filter.EventID != nil {
		params.Eventid = *filter.EventID
	}
	if filter.TeamID != nil {
		params.Teamid = *filter.TeamID
	}
	if filter.Ids != nil {
		params.Ids = filter.Ids
	}

	// Apply pagination
	if limit != nil {
		params.Querylimit = int32(*limit)
	}
	if offset != nil {
		params.Queryoffset = int32(*offset)
	}

	return params
}

// buildUserFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildUserFilterParamsCursor(filter *model.UserFilter, first *int, after *string, last *int, before *string) (sqlc.GetUsersFilteredCursorParams, error) {
	params := sqlc.GetUsersFilteredCursorParams{}

	// Initialize filter if nil
	if filter == nil {
		filter = &model.UserFilter{}
	}

	// Apply filters
	if filter.Query != nil && *filter.Query != "" {
		params.Query = *filter.Query
	}

	if filter.ChurchID != nil {
		params.Churchid = *filter.ChurchID
	}

	if filter.Gender != nil {
		params.Gender = string(*filter.Gender)
	}

	if filter.MinAge != nil {
		params.Minage = int32(*filter.MinAge)
	}

	if filter.MaxAge != nil {
		params.Maxage = int32(*filter.MaxAge)
	} else {
		params.Maxage = 1000
	}

	if filter.ProjectID != nil {
		params.Projectid = *filter.ProjectID
	}

	if filter.EventID != nil {
		params.Eventid = *filter.EventID
	}

	if filter.TeamID != nil {
		params.Teamid = *filter.TeamID
	}

	if filter.Ids != nil {
		params.Ids = filter.Ids
	}

	// Handle cursor pagination
	// Determine pagination direction and limits
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

	// Handle cursors
	if after != nil && *after != "" {
		params.Aftercursor = *after
	}
	if before != nil && *before != "" {
		params.Beforecursor = *before
	}

	return params, nil
}

// buildCountFilterParams converts GraphQL filter to count query parameters
func buildCountFilterParams(filter *model.UserFilter) sqlc.CountUsersFilteredParams {
	params := sqlc.CountUsersFilteredParams{}

	// Initialize filter if nil
	if filter == nil {
		filter = &model.UserFilter{}
	}

	// Apply filters
	if filter.Query != nil && *filter.Query != "" {
		params.Query = *filter.Query
	}
	if filter.ChurchID != nil {
		params.Churchid = *filter.ChurchID
	}
	if filter.Gender != nil {
		params.Gender = string(*filter.Gender)
	}
	if filter.MinAge != nil {
		params.Minage = int32(*filter.MinAge)
	}
	if filter.MaxAge != nil {
		params.Maxage = int32(*filter.MaxAge)
	} else {
		params.Maxage = 1000
	}
	if filter.ProjectID != nil {
		params.Projectid = *filter.ProjectID
	}
	if filter.EventID != nil {
		params.Eventid = *filter.EventID
	}
	if filter.TeamID != nil {
		params.Teamid = *filter.TeamID
	}
	if filter.Ids != nil {
		params.Ids = filter.Ids
	}

	return params
}

// buildCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildCacheKeyParams(filter *model.UserFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if filter.Query != nil && *filter.Query != "" {
			params["query"] = *filter.Query
		}
		if filter.ChurchID != nil && *filter.ChurchID != "" {
			params["churchid"] = *filter.ChurchID
		}
		if filter.Gender != nil {
			params["gender"] = string(*filter.Gender)
		}
		if filter.MinAge != nil {
			params["minage"] = fmt.Sprintf("%d", *filter.MinAge)
		}
		if filter.MaxAge != nil {
			params["maxage"] = fmt.Sprintf("%d", *filter.MaxAge)
		}
		if filter.ProjectID != nil && *filter.ProjectID != "" {
			params["projectid"] = *filter.ProjectID
		}
		if filter.EventID != nil && *filter.EventID != "" {
			params["eventid"] = *filter.EventID
		}
		if filter.TeamID != nil && *filter.TeamID != "" {
			params["teamid"] = *filter.TeamID
		}
		if len(filter.Ids) > 0 {
			// Join IDs with comma for deterministic key
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
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
