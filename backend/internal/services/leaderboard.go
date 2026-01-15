package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/utils"
)

// LeaderboardQuerier defines the database operations needed for leaderboards
type LeaderboardQuerier interface {
	// Project leaderboards
	GetProjectPersonLeaderboard(ctx context.Context, params sqlc.GetProjectPersonLeaderboardParams) ([]*sqlc.GetProjectPersonLeaderboardRow, error)
	FindMyProjectPersonPosition(ctx context.Context, params sqlc.FindMyProjectPersonPositionParams) (*sqlc.FindMyProjectPersonPositionRow, error)
	CountProjectPersonLeaderboard(ctx context.Context, params sqlc.CountProjectPersonLeaderboardParams) (int64, error)
	GetFullProjectPersonLeaderboard(ctx context.Context, params sqlc.GetFullProjectPersonLeaderboardParams) ([]*sqlc.GetFullProjectPersonLeaderboardRow, error)

	GetProjectTeamLeaderboard(ctx context.Context, params sqlc.GetProjectTeamLeaderboardParams) ([]*sqlc.GetProjectTeamLeaderboardRow, error)
	FindMyProjectTeamPosition(ctx context.Context, params sqlc.FindMyProjectTeamPositionParams) (*sqlc.FindMyProjectTeamPositionRow, error)
	CountProjectTeamLeaderboard(ctx context.Context, params sqlc.CountProjectTeamLeaderboardParams) (int64, error)
	GetFullProjectTeamLeaderboard(ctx context.Context, params sqlc.GetFullProjectTeamLeaderboardParams) ([]*sqlc.GetFullProjectTeamLeaderboardRow, error)

	GetProjectSuperTeamLeaderboard(ctx context.Context, params sqlc.GetProjectSuperTeamLeaderboardParams) ([]*sqlc.GetProjectSuperTeamLeaderboardRow, error)
	FindMyProjectSuperTeamPosition(ctx context.Context, params sqlc.FindMyProjectSuperTeamPositionParams) (*sqlc.FindMyProjectSuperTeamPositionRow, error)
	CountProjectSuperTeamLeaderboard(ctx context.Context, projectid string) (int64, error)
	GetFullProjectSuperTeamLeaderboard(ctx context.Context, params sqlc.GetFullProjectSuperTeamLeaderboardParams) ([]*sqlc.GetFullProjectSuperTeamLeaderboardRow, error)

	GetProjectChurchLeaderboard(ctx context.Context, params sqlc.GetProjectChurchLeaderboardParams) ([]*sqlc.GetProjectChurchLeaderboardRow, error)
	FindMyProjectChurchPosition(ctx context.Context, params sqlc.FindMyProjectChurchPositionParams) (*sqlc.FindMyProjectChurchPositionRow, error)
	CountProjectChurchLeaderboard(ctx context.Context, params sqlc.CountProjectChurchLeaderboardParams) (int64, error)
	GetFullProjectChurchLeaderboard(ctx context.Context, params sqlc.GetFullProjectChurchLeaderboardParams) ([]*sqlc.GetFullProjectChurchLeaderboardRow, error)

	// Event leaderboards
	GetEventPersonLeaderboard(ctx context.Context, params sqlc.GetEventPersonLeaderboardParams) ([]*sqlc.GetEventPersonLeaderboardRow, error)
	FindMyEventPersonPosition(ctx context.Context, params sqlc.FindMyEventPersonPositionParams) (*sqlc.FindMyEventPersonPositionRow, error)
	CountEventPersonLeaderboard(ctx context.Context, params sqlc.CountEventPersonLeaderboardParams) (int64, error)
	GetFullEventPersonLeaderboard(ctx context.Context, params sqlc.GetFullEventPersonLeaderboardParams) ([]*sqlc.GetFullEventPersonLeaderboardRow, error)

	GetEventTeamLeaderboard(ctx context.Context, params sqlc.GetEventTeamLeaderboardParams) ([]*sqlc.GetEventTeamLeaderboardRow, error)
	FindMyEventTeamPosition(ctx context.Context, params sqlc.FindMyEventTeamPositionParams) (*sqlc.FindMyEventTeamPositionRow, error)
	CountEventTeamLeaderboard(ctx context.Context, eventid string) (int64, error)
	GetFullEventTeamLeaderboard(ctx context.Context, params sqlc.GetFullEventTeamLeaderboardParams) ([]*sqlc.GetFullEventTeamLeaderboardRow, error)

	GetEventSuperTeamLeaderboard(ctx context.Context, params sqlc.GetEventSuperTeamLeaderboardParams) ([]*sqlc.GetEventSuperTeamLeaderboardRow, error)
	FindMyEventSuperTeamPosition(ctx context.Context, params sqlc.FindMyEventSuperTeamPositionParams) (*sqlc.FindMyEventSuperTeamPositionRow, error)
	CountEventSuperTeamLeaderboard(ctx context.Context, eventid string) (int64, error)
	GetFullEventSuperTeamLeaderboard(ctx context.Context, params sqlc.GetFullEventSuperTeamLeaderboardParams) ([]*sqlc.GetFullEventSuperTeamLeaderboardRow, error)

	GetEventChurchLeaderboard(ctx context.Context, params sqlc.GetEventChurchLeaderboardParams) ([]*sqlc.GetEventChurchLeaderboardRow, error)
	FindMyEventChurchPosition(ctx context.Context, params sqlc.FindMyEventChurchPositionParams) (*sqlc.FindMyEventChurchPositionRow, error)
	CountEventChurchLeaderboard(ctx context.Context, params sqlc.CountEventChurchLeaderboardParams) (int64, error)
	GetFullEventChurchLeaderboard(ctx context.Context, params sqlc.GetFullEventChurchLeaderboardParams) ([]*sqlc.GetFullEventChurchLeaderboardRow, error)
}

// LeaderboardService provides leaderboard functionality with caching
type LeaderboardService struct {
	queries LeaderboardQuerier
	cache   *cache.Cache
	loaders *loaders.Loaders
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(queries LeaderboardQuerier, c *cache.Cache, l *loaders.Loaders) *LeaderboardService {
	return &LeaderboardService{
		queries: queries,
		cache:   c,
		loaders: l,
	}
}

// LeaderboardParams contains common parameters for leaderboard requests
type LeaderboardParams struct {
	ContextID  string // Project ID or Event ID
	EntityType model.LeaderboardEntityType
	Filter     *model.LeaderboardFilter
	First      *int
	After      *string
	Last       *int
	Before     *string
	UserID     string // Current user ID for "me" lookups
}

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	EntityID    string
	Name        string
	Description string
	Image       *string
	Score       int
	Rank        int64
	LastScoreAt *time.Time
}

// GetProjectLeaderboard retrieves leaderboard for a project
func (s *LeaderboardService) GetProjectLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	switch params.EntityType {
	case model.LeaderboardEntityTypePersons:
		return s.getProjectPersonLeaderboard(ctx, params)
	case model.LeaderboardEntityTypeTeams:
		return s.getProjectTeamLeaderboard(ctx, params)
	case model.LeaderboardEntityTypeSuperteams:
		return s.getProjectSuperTeamLeaderboard(ctx, params)
	case model.LeaderboardEntityTypeChurches:
		return s.getProjectChurchLeaderboard(ctx, params)
	default:
		return nil, nil, 0, fmt.Errorf("invalid entity type: %s", params.EntityType)
	}
}

// GetEventLeaderboard retrieves leaderboard for an event
func (s *LeaderboardService) GetEventLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	switch params.EntityType {
	case model.LeaderboardEntityTypePersons:
		return s.getEventPersonLeaderboard(ctx, params)
	case model.LeaderboardEntityTypeTeams:
		return s.getEventTeamLeaderboard(ctx, params)
	case model.LeaderboardEntityTypeSuperteams:
		return s.getEventSuperTeamLeaderboard(ctx, params)
	case model.LeaderboardEntityTypeChurches:
		return s.getEventChurchLeaderboard(ctx, params)
	default:
		return nil, nil, 0, fmt.Errorf("invalid entity type: %s", params.EntityType)
	}
}

// Helper functions for project leaderboards

func (s *LeaderboardService) getProjectPersonLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	// Build cache key
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("project", params.ContextID, "persons", filterParams)

	// Try to get from cache
	var fullLeaderboard []LeaderboardEntry
	cachedData, found := s.cache.Get(cacheKey)
	if found {
		// Unmarshal from cache
		if cachedBytes, ok := cachedData.([]byte); ok {
			if err := json.Unmarshal(cachedBytes, &fullLeaderboard); err == nil {
				// Successfully got from cache
				total := len(fullLeaderboard)

				// Find "me" in cached results
				var meEntry *LeaderboardEntry
				if params.UserID != "" {
					meEntry = findMeInLeaderboard(fullLeaderboard, params.UserID)
				}

				// Paginate in-memory
				paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

				return paginated, meEntry, total, nil
			}
		}
	}

	// Cache miss or unmarshal error - query database
	queryParams := s.buildFullProjectPersonParams(params)
	rows, err := s.queries.GetFullProjectPersonLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get full project person leaderboard: %w", err)
	}

	// Convert to entries
	fullLeaderboard = make([]LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			fullLeaderboard = append(fullLeaderboard, LeaderboardEntry{
				EntityID:    row.EntityID,
				Name:        row.Name,
				Description: row.ChurchName,
				Image:       row.Image,
				Score:       int(row.Score),
				Rank:        row.Rank,
				LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
			})
		}
	}

	// Cache the full leaderboard with 5 minute TTL
	if marshaledData, err := json.Marshal(fullLeaderboard); err == nil {
		s.cache.SetWithTTL(cacheKey, marshaledData, 5*time.Minute)
	}

	total := len(fullLeaderboard)

	// Find "me" in results
	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		meEntry = findMeInLeaderboard(fullLeaderboard, params.UserID)
	}

	// Paginate in-memory
	paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

	return paginated, meEntry, total, nil
}

func (s *LeaderboardService) getProjectTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	// Build cache key
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("project", params.ContextID, "teams", filterParams)

	// Try to get from cache
	var fullLeaderboard []LeaderboardEntry
	cachedData, found := s.cache.Get(cacheKey)
	if found {
		// Unmarshal from cache
		if cachedBytes, ok := cachedData.([]byte); ok {
			if err := json.Unmarshal(cachedBytes, &fullLeaderboard); err == nil {
				// Successfully got from cache
				total := len(fullLeaderboard)

				// Find "me" in cached results
				var meEntry *LeaderboardEntry
				if params.UserID != "" {
					// Use loader to get user's teams
					thunk := s.loaders.TeamsByUserLoader.Load(ctx, params.UserID)
					teams, err := thunk()
					if err == nil && teams != nil {
						// Find team in this project
						for _, team := range teams {
							if team.ProjectID == params.ContextID {
								meEntry = findMeInLeaderboard(fullLeaderboard, team.ID)
								break
							}
						}
					}
				}

				// Paginate in-memory
				paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

				return paginated, meEntry, total, nil
			}
		}
	}

	// Cache miss or unmarshal error - query database
	queryParams := s.buildFullProjectTeamParams(params)
	rows, err := s.queries.GetFullProjectTeamLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get full project team leaderboard: %w", err)
	}

	// Convert to entries
	fullLeaderboard = make([]LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			fullLeaderboard = append(fullLeaderboard, LeaderboardEntry{
				EntityID:    row.EntityID,
				Name:        row.Name,
				Image:       row.Image,
				Score:       int(row.Score),
				Rank:        row.Rank,
				LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
			})
		}
	}

	// Cache the full leaderboard with 5 minute TTL
	if marshaledData, err := json.Marshal(fullLeaderboard); err == nil {
		s.cache.SetWithTTL(cacheKey, marshaledData, 5*time.Minute)
	}

	total := len(fullLeaderboard)

	// Find "me" in results
	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		// Use loader to get user's teams
		thunk := s.loaders.TeamsByUserLoader.Load(ctx, params.UserID)
		teams, err := thunk()
		if err == nil && teams != nil {
			// Find team in this project
			for _, team := range teams {
				if team.ProjectID == params.ContextID {
					meEntry = findMeInLeaderboard(fullLeaderboard, team.ID)
					break
				}
			}
		}
	}

	// Paginate in-memory
	paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

	return paginated, meEntry, total, nil
}

func (s *LeaderboardService) getProjectSuperTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	// Build cache key
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("project", params.ContextID, "superteams", filterParams)

	// Try to get from cache
	var fullLeaderboard []LeaderboardEntry
	cachedData, found := s.cache.Get(cacheKey)
	if found {
		// Unmarshal from cache
		if cachedBytes, ok := cachedData.([]byte); ok {
			if err := json.Unmarshal(cachedBytes, &fullLeaderboard); err == nil {
				// Successfully got from cache
				total := len(fullLeaderboard)

				// Find "me" in cached results
				var meEntry *LeaderboardEntry
				if params.UserID != "" {
					// Use loader to get user's superteams
					thunk := s.loaders.SuperTeamsByUserLoader.Load(ctx, params.UserID)
					superTeams, err := thunk()
					if err == nil && superTeams != nil {
						// Find superteam in this project
						for _, superTeam := range superTeams {
							if superTeam.ProjectID == params.ContextID {
								meEntry = findMeInLeaderboard(fullLeaderboard, superTeam.ID)
								break
							}
						}
					}
				}

				// Paginate in-memory
				paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

				return paginated, meEntry, total, nil
			}
		}
	}

	// Cache miss or unmarshal error - query database
	queryParams := s.buildFullProjectSuperTeamParams(params)
	rows, err := s.queries.GetFullProjectSuperTeamLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get full project superteam leaderboard: %w", err)
	}

	// Convert to entries
	fullLeaderboard = make([]LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			fullLeaderboard = append(fullLeaderboard, LeaderboardEntry{
				EntityID:    row.EntityID,
				Name:        row.Name,
				Image:       row.Image,
				Score:       int(row.Score),
				Rank:        row.Rank,
				LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
			})
		}
	}

	// Cache the full leaderboard with 5 minute TTL
	if marshaledData, err := json.Marshal(fullLeaderboard); err == nil {
		s.cache.SetWithTTL(cacheKey, marshaledData, 5*time.Minute)
	}

	total := len(fullLeaderboard)

	// Find "me" in results
	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		// Use loader to get user's superteams
		thunk := s.loaders.SuperTeamsByUserLoader.Load(ctx, params.UserID)
		superTeams, err := thunk()
		if err == nil && superTeams != nil {
			// Find superteam in this project
			for _, superTeam := range superTeams {
				if superTeam.ProjectID == params.ContextID {
					meEntry = findMeInLeaderboard(fullLeaderboard, superTeam.ID)
					break
				}
			}
		}
	}

	// Paginate in-memory
	paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

	return paginated, meEntry, total, nil
}

func (s *LeaderboardService) getProjectChurchLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	// Build cache key
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("project", params.ContextID, "churches", filterParams)

	// Try to get from cache
	var fullLeaderboard []LeaderboardEntry
	cachedData, found := s.cache.Get(cacheKey)
	if found {
		// Unmarshal from cache
		if cachedBytes, ok := cachedData.([]byte); ok {
			if err := json.Unmarshal(cachedBytes, &fullLeaderboard); err == nil {
				// Successfully got from cache
				total := len(fullLeaderboard)

				// Find "me" in cached results
				var meEntry *LeaderboardEntry
				if params.UserID != "" {
					// Use loader to get user (includes ChurchID)
					thunk := s.loaders.UserByIDLoader.Load(ctx, params.UserID)
					user, err := thunk()
					if err == nil && user != nil {
						meEntry = findMeInLeaderboard(fullLeaderboard, user.ChurchID)
					}
				}

				// Paginate in-memory
				paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

				return paginated, meEntry, total, nil
			}
		}
	}

	// Cache miss or unmarshal error - query database
	queryParams := s.buildFullProjectChurchParams(params)
	rows, err := s.queries.GetFullProjectChurchLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get full project church leaderboard: %w", err)
	}

	// Convert to entries
	fullLeaderboard = make([]LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			fullLeaderboard = append(fullLeaderboard, LeaderboardEntry{
				EntityID:    row.EntityID,
				Name:        row.Name,
				Image:       row.Image,
				Score:       int(row.Score),
				Rank:        row.Rank,
				LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
			})
		}
	}

	// Cache the full leaderboard with 5 minute TTL
	if marshaledData, err := json.Marshal(fullLeaderboard); err == nil {
		s.cache.SetWithTTL(cacheKey, marshaledData, 5*time.Minute)
	}

	total := len(fullLeaderboard)

	// Find "me" in results
	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		// Use loader to get user (includes ChurchID)
		thunk := s.loaders.UserByIDLoader.Load(ctx, params.UserID)
		user, err := thunk()
		if err == nil && user != nil {
			meEntry = findMeInLeaderboard(fullLeaderboard, user.ChurchID)
		}
	}

	// Paginate in-memory
	paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

	return paginated, meEntry, total, nil
}

// Helper functions for event leaderboards

func (s *LeaderboardService) getEventPersonLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	// Build cache key
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("event", params.ContextID, "persons", filterParams)

	// Try to get from cache
	var fullLeaderboard []LeaderboardEntry
	cachedData, found := s.cache.Get(cacheKey)
	if found {
		// Unmarshal from cache
		if cachedBytes, ok := cachedData.([]byte); ok {
			if err := json.Unmarshal(cachedBytes, &fullLeaderboard); err == nil {
				// Successfully got from cache
				total := len(fullLeaderboard)

				// Find "me" in cached results
				var meEntry *LeaderboardEntry
				if params.UserID != "" {
					meEntry = findMeInLeaderboard(fullLeaderboard, params.UserID)
				}

				// Paginate in-memory
				paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

				return paginated, meEntry, total, nil
			}
		}
	}

	// Cache miss or unmarshal error - query database
	queryParams := s.buildFullEventPersonParams(params)
	rows, err := s.queries.GetFullEventPersonLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get full event person leaderboard: %w", err)
	}

	// Convert to entries
	fullLeaderboard = make([]LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			fullLeaderboard = append(fullLeaderboard, LeaderboardEntry{
				EntityID:    row.EntityID,
				Name:        row.Name,
				Description: row.ChurchName,
				Image:       row.Image,
				Score:       int(row.Score),
				Rank:        row.Rank,
				LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
			})
		}
	}

	// Cache the full leaderboard with 5 minute TTL
	if marshaledData, err := json.Marshal(fullLeaderboard); err == nil {
		s.cache.SetWithTTL(cacheKey, marshaledData, 5*time.Minute)
	}

	total := len(fullLeaderboard)

	// Find "me" in results
	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		meEntry = findMeInLeaderboard(fullLeaderboard, params.UserID)
	}

	// Paginate in-memory
	paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

	return paginated, meEntry, total, nil
}

func (s *LeaderboardService) getEventTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	// Build cache key
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("event", params.ContextID, "teams", filterParams)

	// Try to get from cache
	var fullLeaderboard []LeaderboardEntry
	cachedData, found := s.cache.Get(cacheKey)
	if found {
		// Unmarshal from cache
		if cachedBytes, ok := cachedData.([]byte); ok {
			if err := json.Unmarshal(cachedBytes, &fullLeaderboard); err == nil {
				// Successfully got from cache
				total := len(fullLeaderboard)

				// Find "me" in cached results
				var meEntry *LeaderboardEntry
				if params.UserID != "" {
					// Get event to find its project, then get user's teams
					eventThunk := s.loaders.EventByIDLoader.Load(ctx, params.ContextID)
					event, err := eventThunk()
					if err == nil && event != nil {
						teamsThunk := s.loaders.TeamsByUserLoader.Load(ctx, params.UserID)
						teams, err := teamsThunk()
						if err == nil && teams != nil {
							// Find team in this event's project
							for _, team := range teams {
								if team.ProjectID == event.ProjectID {
									meEntry = findMeInLeaderboard(fullLeaderboard, team.ID)
									break
								}
							}
						}
					}
				}

				// Paginate in-memory
				paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

				return paginated, meEntry, total, nil
			}
		}
	}

	// Cache miss or unmarshal error - query database
	queryParams := s.buildFullEventTeamParams(params)
	rows, err := s.queries.GetFullEventTeamLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get full event team leaderboard: %w", err)
	}

	// Convert to entries
	fullLeaderboard = make([]LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			fullLeaderboard = append(fullLeaderboard, LeaderboardEntry{
				EntityID:    row.EntityID,
				Name:        row.Name,
				Image:       row.Image,
				Score:       int(row.Score),
				Rank:        row.Rank,
				LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
			})
		}
	}

	// Cache the full leaderboard with 5 minute TTL
	if marshaledData, err := json.Marshal(fullLeaderboard); err == nil {
		s.cache.SetWithTTL(cacheKey, marshaledData, 5*time.Minute)
	}

	total := len(fullLeaderboard)

	// Find "me" in results
	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		// Get event to find its project, then get user's teams
		eventThunk := s.loaders.EventByIDLoader.Load(ctx, params.ContextID)
		event, err := eventThunk()
		if err == nil && event != nil {
			teamsThunk := s.loaders.TeamsByUserLoader.Load(ctx, params.UserID)
			teams, err := teamsThunk()
			if err == nil && teams != nil {
				// Find team in this event's project
				for _, team := range teams {
					if team.ProjectID == event.ProjectID {
						meEntry = findMeInLeaderboard(fullLeaderboard, team.ID)
						break
					}
				}
			}
		}
	}

	// Paginate in-memory
	paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

	return paginated, meEntry, total, nil
}

func (s *LeaderboardService) getEventSuperTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	// Build cache key
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("event", params.ContextID, "superteams", filterParams)

	// Try to get from cache
	var fullLeaderboard []LeaderboardEntry
	cachedData, found := s.cache.Get(cacheKey)
	if found {
		// Unmarshal from cache
		if cachedBytes, ok := cachedData.([]byte); ok {
			if err := json.Unmarshal(cachedBytes, &fullLeaderboard); err == nil {
				// Successfully got from cache
				total := len(fullLeaderboard)

				// Find "me" in cached results
				var meEntry *LeaderboardEntry
				if params.UserID != "" {
					// Get event to find its project, then get user's superteams
					eventThunk := s.loaders.EventByIDLoader.Load(ctx, params.ContextID)
					event, err := eventThunk()
					if err == nil && event != nil {
						superTeamsThunk := s.loaders.SuperTeamsByUserLoader.Load(ctx, params.UserID)
						superTeams, err := superTeamsThunk()
						if err == nil && superTeams != nil {
							// Find superteam in this event's project
							for _, superTeam := range superTeams {
								if superTeam.ProjectID == event.ProjectID {
									meEntry = findMeInLeaderboard(fullLeaderboard, superTeam.ID)
									break
								}
							}
						}
					}
				}

				// Paginate in-memory
				paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

				return paginated, meEntry, total, nil
			}
		}
	}

	// Cache miss or unmarshal error - query database
	queryParams := s.buildFullEventSuperTeamParams(params)
	rows, err := s.queries.GetFullEventSuperTeamLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get full event superteam leaderboard: %w", err)
	}

	// Convert to entries
	fullLeaderboard = make([]LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			fullLeaderboard = append(fullLeaderboard, LeaderboardEntry{
				EntityID:    row.EntityID,
				Name:        row.Name,
				Image:       row.Image,
				Score:       int(row.Score),
				Rank:        row.Rank,
				LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
			})
		}
	}

	// Cache the full leaderboard with 5 minute TTL
	if marshaledData, err := json.Marshal(fullLeaderboard); err == nil {
		s.cache.SetWithTTL(cacheKey, marshaledData, 5*time.Minute)
	}

	total := len(fullLeaderboard)

	// Find "me" in results
	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		// Get event to find its project, then get user's superteams
		eventThunk := s.loaders.EventByIDLoader.Load(ctx, params.ContextID)
		event, err := eventThunk()
		if err == nil && event != nil {
			superTeamsThunk := s.loaders.SuperTeamsByUserLoader.Load(ctx, params.UserID)
			superTeams, err := superTeamsThunk()
			if err == nil && superTeams != nil {
				// Find superteam in this event's project
				for _, superTeam := range superTeams {
					if superTeam.ProjectID == event.ProjectID {
						meEntry = findMeInLeaderboard(fullLeaderboard, superTeam.ID)
						break
					}
				}
			}
		}
	}

	// Paginate in-memory
	paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

	return paginated, meEntry, total, nil
}

func (s *LeaderboardService) getEventChurchLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	// Build cache key
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("event", params.ContextID, "churches", filterParams)

	// Try to get from cache
	var fullLeaderboard []LeaderboardEntry
	cachedData, found := s.cache.Get(cacheKey)
	if found {
		// Unmarshal from cache
		if cachedBytes, ok := cachedData.([]byte); ok {
			if err := json.Unmarshal(cachedBytes, &fullLeaderboard); err == nil {
				// Successfully got from cache
				total := len(fullLeaderboard)

				// Find "me" in cached results
				var meEntry *LeaderboardEntry
				if params.UserID != "" {
					// Use loader to get user (includes ChurchID)
					thunk := s.loaders.UserByIDLoader.Load(ctx, params.UserID)
					user, err := thunk()
					if err == nil && user != nil {
						meEntry = findMeInLeaderboard(fullLeaderboard, user.ChurchID)
					}
				}

				// Paginate in-memory
				paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

				return paginated, meEntry, total, nil
			}
		}
	}

	// Cache miss or unmarshal error - query database
	queryParams := s.buildFullEventChurchParams(params)
	rows, err := s.queries.GetFullEventChurchLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get full event church leaderboard: %w", err)
	}

	// Convert to entries
	fullLeaderboard = make([]LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			fullLeaderboard = append(fullLeaderboard, LeaderboardEntry{
				EntityID:    row.EntityID,
				Name:        row.Name,
				Image:       row.Image,
				Score:       int(row.Score),
				Rank:        row.Rank,
				LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
			})
		}
	}

	// Cache the full leaderboard with 5 minute TTL
	if marshaledData, err := json.Marshal(fullLeaderboard); err == nil {
		s.cache.SetWithTTL(cacheKey, marshaledData, 5*time.Minute)
	}

	total := len(fullLeaderboard)

	// Find "me" in results
	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		// Use loader to get user (includes ChurchID)
		thunk := s.loaders.UserByIDLoader.Load(ctx, params.UserID)
		user, err := thunk()
		if err == nil && user != nil {
			meEntry = findMeInLeaderboard(fullLeaderboard, user.ChurchID)
		}
	}

	// Paginate in-memory
	paginated := paginateLeaderboard(fullLeaderboard, params.First, params.After, params.Last, params.Before)

	return paginated, meEntry, total, nil
}

// Parameter builders for project queries

func (s *LeaderboardService) buildFullProjectPersonParams(params LeaderboardParams) sqlc.GetFullProjectPersonLeaderboardParams {
	return sqlc.GetFullProjectPersonLeaderboardParams{
		Projectid:   params.ContextID,
		Churchid:    getFilterString(params.Filter, "churchId"),
		Minscore:    getFilterInt(params.Filter, "minScore"),
		Maxscore:    getFilterInt(params.Filter, "maxScore"),
		Minage:      getFilterInt(params.Filter, "minAge"),
		Maxage:      getFilterInt(params.Filter, "maxAge"),
		Teamid:      getFilterString(params.Filter, "teamId"),
		Superteamid: getFilterString(params.Filter, "superTeamId"),
	}
}

func (s *LeaderboardService) buildFullProjectTeamParams(params LeaderboardParams) sqlc.GetFullProjectTeamLeaderboardParams {
	return sqlc.GetFullProjectTeamLeaderboardParams{
		Projectid: params.ContextID,
		Minscore:  getFilterInt(params.Filter, "minScore"),
		Maxscore:  getFilterInt(params.Filter, "maxScore"),
	}
}

func (s *LeaderboardService) buildFullProjectSuperTeamParams(params LeaderboardParams) sqlc.GetFullProjectSuperTeamLeaderboardParams {
	return sqlc.GetFullProjectSuperTeamLeaderboardParams{
		Projectid: params.ContextID,
		Minscore:  getFilterInt(params.Filter, "minScore"),
		Maxscore:  getFilterInt(params.Filter, "maxScore"),
	}
}

func (s *LeaderboardService) buildFullProjectChurchParams(params LeaderboardParams) sqlc.GetFullProjectChurchLeaderboardParams {
	return sqlc.GetFullProjectChurchLeaderboardParams{
		Projectid: params.ContextID,
		Minscore:  getFilterInt(params.Filter, "minScore"),
		Maxscore:  getFilterInt(params.Filter, "maxScore"),
	}
}

// Parameter builders for event queries

func (s *LeaderboardService) buildFullEventPersonParams(params LeaderboardParams) sqlc.GetFullEventPersonLeaderboardParams {
	return sqlc.GetFullEventPersonLeaderboardParams{
		Eventid:     params.ContextID,
		Churchid:    getFilterString(params.Filter, "churchId"),
		Minscore:    getFilterInt(params.Filter, "minScore"),
		Maxscore:    getFilterInt(params.Filter, "maxScore"),
		Minage:      getFilterInt(params.Filter, "minAge"),
		Maxage:      getFilterInt(params.Filter, "maxAge"),
		Teamid:      getFilterString(params.Filter, "teamId"),
		Superteamid: getFilterString(params.Filter, "superTeamId"),
	}
}

func (s *LeaderboardService) buildFullEventTeamParams(params LeaderboardParams) sqlc.GetFullEventTeamLeaderboardParams {
	return sqlc.GetFullEventTeamLeaderboardParams{
		Eventid:  params.ContextID,
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
	}
}

func (s *LeaderboardService) buildFullEventSuperTeamParams(params LeaderboardParams) sqlc.GetFullEventSuperTeamLeaderboardParams {
	return sqlc.GetFullEventSuperTeamLeaderboardParams{
		Eventid:  params.ContextID,
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
	}
}

func (s *LeaderboardService) buildFullEventChurchParams(params LeaderboardParams) sqlc.GetFullEventChurchLeaderboardParams {
	return sqlc.GetFullEventChurchLeaderboardParams{
		Eventid:  params.ContextID,
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
	}
}

// Helper functions

func getFilterString(filter *model.LeaderboardFilter, field string) string {
	if filter == nil {
		return ""
	}

	switch field {
	case "churchId":
		if filter.ChurchID != nil {
			return *filter.ChurchID
		}
	case "country":
		if filter.Country != nil {
			return *filter.Country
		}
	case "churchCategory":
		if filter.ChurchCategory != nil {
			return string(*filter.ChurchCategory)
		}
	case "gender":
		if filter.Gender != nil {
			return string(*filter.Gender)
		}
	case "teamId":
		if filter.TeamID != nil {
			return *filter.TeamID
		}
	case "superTeamId":
		if filter.SuperTeamID != nil {
			return *filter.SuperTeamID
		}
	}

	return ""
}

func getFilterInt(filter *model.LeaderboardFilter, field string) int32 {
	if filter == nil {
		switch field {
		case "minScore", "minAge":
			return math.MinInt32
		case "maxScore", "maxAge":
			return math.MaxInt32
		}
	}

	switch field {
	case "minScore":
		if filter.MinScore != nil {
			return int32(*filter.MinScore)
		}
		return math.MinInt32
	case "maxScore":
		if filter.MaxScore != nil {
			return int32(*filter.MaxScore)
		}
		return math.MaxInt32
	case "minAge":
		if filter.AgeRange != nil {
			return int32(filter.AgeRange.Min)
		}
		return math.MinInt32
	case "maxAge":
		if filter.AgeRange != nil {
			return int32(filter.AgeRange.Max)
		}
		return math.MaxInt32
	}

	return 0
}

// buildFilterParamsMap builds a map of filter parameters for cache key generation
func buildFilterParamsMap(filter *model.LeaderboardFilter) map[string]string {
	params := make(map[string]string)

	if filter == nil {
		return params
	}

	if filter.ChurchID != nil {
		params["churchId"] = *filter.ChurchID
	}
	if filter.Country != nil {
		params["country"] = *filter.Country
	}
	if filter.ChurchCategory != nil {
		params["churchCategory"] = string(*filter.ChurchCategory)
	}
	if filter.Gender != nil {
		params["gender"] = string(*filter.Gender)
	}
	if filter.TeamID != nil {
		params["teamId"] = *filter.TeamID
	}
	if filter.SuperTeamID != nil {
		params["superTeamId"] = *filter.SuperTeamID
	}
	if filter.MinScore != nil {
		params["minScore"] = strconv.Itoa(*filter.MinScore)
	}
	if filter.MaxScore != nil {
		params["maxScore"] = strconv.Itoa(*filter.MaxScore)
	}
	if filter.AgeRange != nil {
		params["minAge"] = strconv.Itoa(filter.AgeRange.Min)
		params["maxAge"] = strconv.Itoa(filter.AgeRange.Max)
	}

	return params
}

// paginateLeaderboard performs in-memory pagination on a leaderboard slice
func paginateLeaderboard(entries []LeaderboardEntry, first *int, after *string, last *int, before *string) []LeaderboardEntry {
	if len(entries) == 0 {
		return entries
	}

	// Handle forward pagination (first/after)
	if first != nil {
		startIdx := 0
		if after != nil && *after != "" {
			// after is a rank, find the index after that rank
			afterRank, err := strconv.ParseInt(*after, 10, 64)
			if err == nil {
				for i, entry := range entries {
					if entry.Rank == afterRank {
						startIdx = i + 1
						break
					}
				}
			}
		}

		endIdx := startIdx + *first
		if endIdx > len(entries) {
			endIdx = len(entries)
		}

		return entries[startIdx:endIdx]
	}

	// Handle backward pagination (last/before)
	if last != nil {
		endIdx := len(entries)
		if before != nil && *before != "" {
			// before is a rank, find the index before that rank
			beforeRank, err := strconv.ParseInt(*before, 10, 64)
			if err == nil {
				for i, entry := range entries {
					if entry.Rank == beforeRank {
						endIdx = i
						break
					}
				}
			}
		}

		startIdx := endIdx - *last
		if startIdx < 0 {
			startIdx = 0
		}

		return entries[startIdx:endIdx]
	}

	// No pagination params, return first 10
	if len(entries) > 10 {
		return entries[:10]
	}
	return entries
}

// findMeInLeaderboard finds the current user's entry in the leaderboard
// For persons: matches by entity_id == userID
// For teams/superteams/churches: needs to check membership (handled by caller)
func findMeInLeaderboard(entries []LeaderboardEntry, entityID string) *LeaderboardEntry {
	for _, entry := range entries {
		if entry.EntityID == entityID {
			return &entry
		}
	}
	return nil
}
