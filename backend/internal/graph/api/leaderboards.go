package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/pagination"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/jackc/pgx/v5/pgtype"
)

// defaultLeaderboardConfigsPageSize is the page size used by the admin leaderboardConfigs
// cursor query when neither first nor last is specified. Kept as a single constant so the
// two call sites (query-limit calculation, hasMore trimming) can't drift out of sync.
const defaultLeaderboardConfigsPageSize = 10

// marshalLeaderboardFilter serializes a LeaderboardFilter input into the JSON bytes
// stored in the leaderboard_configs.filter JSONB column. Returns nil for a nil filter.
func marshalLeaderboardFilter(filter *model.LeaderboardFilter) ([]byte, error) {
	if filter == nil {
		return nil, nil
	}
	b, err := json.Marshal(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize filter: %w", err)
	}
	return b, nil
}

// getVisibleLeaderboardConfigsByProject loads all leaderboard configs for a project via the
// dataloader and filters out inactive ones unless the requesting user is an admin/superadmin.
func (r *Resolver) getVisibleLeaderboardConfigsByProject(ctx context.Context, projectID string) ([]model.LeaderboardConfig, error) {
	thunk := r.Loaders.LeaderboardConfigsByProjectLoader.Load(ctx, projectID)
	configs, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load leaderboard configs: %w", err)
	}
	return r.filterVisibleLeaderboardConfigs(ctx, configs), nil
}

// getVisibleLeaderboardConfigsByEvent loads all leaderboard configs for an event via the
// dataloader and filters out inactive ones unless the requesting user is an admin/superadmin.
func (r *Resolver) getVisibleLeaderboardConfigsByEvent(ctx context.Context, eventID string) ([]model.LeaderboardConfig, error) {
	thunk := r.Loaders.LeaderboardConfigsByEventLoader.Load(ctx, eventID)
	configs, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load leaderboard configs: %w", err)
	}
	return r.filterVisibleLeaderboardConfigs(ctx, configs), nil
}

// filterVisibleLeaderboardConfigs drops inactive configs for non-admin viewers.
func (r *Resolver) filterVisibleLeaderboardConfigs(ctx context.Context, configs []*model.LeaderboardConfig) []model.LeaderboardConfig {
	isAdmin := false
	if userID, ok := middleware.GetUserID(ctx); ok && userID != "" {
		isAdmin = r.RoleService.IsAdmin(ctx, userID)
	}
	return filterConfigsByVisibility(configs, isAdmin)
}

// filterConfigsByVisibility drops inactive configs unless the viewer is an admin.
func filterConfigsByVisibility(configs []*model.LeaderboardConfig, isAdmin bool) []model.LeaderboardConfig {
	result := make([]model.LeaderboardConfig, 0, len(configs))
	for _, config := range configs {
		if config.IsActive || isAdmin {
			result = append(result, *config)
		}
	}
	return result
}

// buildLeaderboardParamsFromConfig adapts a persisted LeaderboardConfig plus pagination
// args into services.LeaderboardParams, and reports whether it's an event-scoped config
// (so the caller knows to call GetEventLeaderboard instead of GetProjectLeaderboard).
func buildLeaderboardParamsFromConfig(obj *model.LeaderboardConfig, first *int, after *string, last *int, before *string, userID string) (params services.LeaderboardParams, isEvent bool, err error) {
	var filter *model.LeaderboardFilter
	if obj.Filter != nil {
		if err := json.Unmarshal([]byte(*obj.Filter), &filter); err != nil {
			return services.LeaderboardParams{}, false, fmt.Errorf("failed to parse leaderboard config filter: %w", err)
		}
	}

	contextID := obj.ProjectID
	isEvent = obj.EventID != nil
	if isEvent {
		contextID = *obj.EventID
	}

	return services.LeaderboardParams{
		ContextID:  contextID,
		EntityType: obj.EntityType,
		Filter:     filter,
		First:      first,
		After:      after,
		Last:       last,
		Before:     before,
		UserID:     userID,
	}, isEvent, nil
}

// getLeaderboardForConfig computes the finished, paginated leaderboard for a persisted
// config by adapting it into services.LeaderboardParams and reusing the same leaderboard
// engine (caching, pagination, nearestChurchRivals) as the ad-hoc Project/Event.leaderboard fields.
func (r *Resolver) getLeaderboardForConfig(ctx context.Context, obj *model.LeaderboardConfig, first *int, after *string, last *int, before *string) (*model.LeaderboardConnection, error) {
	currentUserID, ok := middleware.GetUserID(ctx)
	if !ok || currentUserID == "" {
		return nil, fmt.Errorf("user not authenticated")
	}

	params, isEvent, err := buildLeaderboardParamsFromConfig(obj, first, after, last, before, currentUserID)
	if err != nil {
		return nil, err
	}

	var entries []services.LeaderboardEntry
	var meEntry *services.LeaderboardEntry
	var totalCount int
	var rivals []services.LeaderboardEntry
	if isEvent {
		entries, meEntry, totalCount, rivals, err = r.LeaderboardService.GetEventLeaderboard(ctx, params)
	} else {
		entries, meEntry, totalCount, rivals, err = r.LeaderboardService.GetProjectLeaderboard(ctx, params)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	if obj.EntityType == model.LeaderboardEntityTypePersons {
		result := FilterPersonLeaderboardEntries(entries, totalCount, first, after)
		entries = result.Entries
		first = result.AdjustedFirst
	}

	connection, err := buildLeaderboardConnection(ctx, entries, meEntry, totalCount, rivals, currentUserID, obj.EntityType, obj.ProjectID, r.Loaders, first, last, after, before)
	if err != nil {
		return nil, fmt.Errorf("failed to build leaderboard connection: %w", err)
	}

	return connection, nil
}

// getFilteredLeaderboardConfigs handles the admin-facing leaderboardConfigs cursor query.
func (r *Resolver) getFilteredLeaderboardConfigs(ctx context.Context, filter *model.LeaderboardConfigFilter, first *int, after *string, last *int, before *string) (*model.LeaderboardConfigConnection, error) {
	var afterCursor, beforeCursor *pagination.LeaderboardConfigCursor
	if after != nil && *after != "" {
		decoded, err := pagination.DecodeLeaderboardConfigCursor(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid after cursor: %w", err)
		}
		afterCursor = &decoded
	}
	if before != nil && *before != "" {
		decoded, err := pagination.DecodeLeaderboardConfigCursor(*before)
		if err != nil {
			return nil, fmt.Errorf("invalid before cursor: %w", err)
		}
		beforeCursor = &decoded
	}

	params, err := buildLeaderboardConfigFilterParamsCursor(filter, first, afterCursor, last, beforeCursor)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Queries.GetLeaderboardConfigsFilteredCursor(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard configs: %w", err)
	}

	totalCount, err := r.DB.Queries.CountLeaderboardConfigsFiltered(ctx, buildCountLeaderboardConfigsFilterParams(filter))
	if err != nil {
		return nil, fmt.Errorf("failed to count leaderboard configs: %w", err)
	}

	requestedLimit := defaultLeaderboardConfigsPageSize
	if first != nil {
		requestedLimit = *first
	} else if last != nil {
		requestedLimit = *last
	}

	hasMore := len(rows) > requestedLimit
	configRows := rows
	if hasMore {
		configRows = rows[:requestedLimit]
	}

	if last != nil {
		for i, j := 0, len(configRows)-1; i < j; i, j = i+1, j-1 {
			configRows[i], configRows[j] = configRows[j], configRows[i]
		}
	}

	configs := make([]*model.LeaderboardConfig, len(configRows))
	for i, row := range configRows {
		configs[i] = loaders.ConvertRowToLeaderboardConfig(row)
	}

	connection := pagination.BuildLeaderboardConfigConnection(pagination.BuildLeaderboardConfigConnectionParams{
		Configs:         configs,
		RequestedFirst:  first,
		RequestedLast:   last,
		RequestedAfter:  after,
		RequestedBefore: before,
		TotalCount:      int(totalCount),
		HasMore:         hasMore,
	})

	return connection, nil
}

// buildLeaderboardConfigFilterParamsCursor converts a GraphQL filter to cursor query parameters
func buildLeaderboardConfigFilterParamsCursor(filter *model.LeaderboardConfigFilter, first *int, afterCursor *pagination.LeaderboardConfigCursor, last *int, beforeCursor *pagination.LeaderboardConfigCursor) (sqlc.GetLeaderboardConfigsFilteredCursorParams, error) {
	params := sqlc.GetLeaderboardConfigsFilteredCursorParams{}

	if filter != nil {
		if len(filter.Ids) > 0 {
			params.Ids = filter.Ids
		}
		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}
		if filter.EventID != nil {
			params.Eventid = *filter.EventID
		}
		params.Isactive = filter.IsActive
	}

	isBackward := false
	var limit int

	if first != nil && last != nil {
		return params, fmt.Errorf("cannot specify both first and last")
	}

	if first != nil {
		limit = *first + 1
		isBackward = false
	} else if last != nil {
		limit = *last + 1
		isBackward = true
	} else {
		limit = defaultLeaderboardConfigsPageSize + 1
		isBackward = false
	}

	params.Querylimit = int32(limit)
	params.Isbackward = isBackward

	if afterCursor != nil && afterCursor.ID != "" {
		params.Aftercursorcreatedat = pgtype.Timestamptz{Time: afterCursor.CreatedAt, Valid: true}
		params.Aftercursorid = afterCursor.ID
	}

	if beforeCursor != nil && beforeCursor.ID != "" {
		params.Beforecursorcreatedat = pgtype.Timestamptz{Time: beforeCursor.CreatedAt, Valid: true}
		params.Beforecursorid = beforeCursor.ID
	}

	return params, nil
}

// buildCountLeaderboardConfigsFilterParams converts a GraphQL filter to count query parameters
func buildCountLeaderboardConfigsFilterParams(filter *model.LeaderboardConfigFilter) sqlc.CountLeaderboardConfigsFilteredParams {
	params := sqlc.CountLeaderboardConfigsFilteredParams{}

	if filter != nil {
		if len(filter.Ids) > 0 {
			params.Ids = filter.Ids
		}
		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}
		if filter.EventID != nil {
			params.Eventid = *filter.EventID
		}
		params.Isactive = filter.IsActive
	}

	return params
}
