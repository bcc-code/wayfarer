package services

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// LeaderboardQuerier defines the database operations needed for leaderboards
type LeaderboardQuerier interface {
	// Project leaderboards
	GetProjectPersonLeaderboard(ctx context.Context, params sqlc.GetProjectPersonLeaderboardParams) ([]*sqlc.GetProjectPersonLeaderboardRow, error)
	FindMyProjectPersonPosition(ctx context.Context, params sqlc.FindMyProjectPersonPositionParams) (*sqlc.FindMyProjectPersonPositionRow, error)
	CountProjectPersonLeaderboard(ctx context.Context, params sqlc.CountProjectPersonLeaderboardParams) (int64, error)

	GetProjectTeamLeaderboard(ctx context.Context, params sqlc.GetProjectTeamLeaderboardParams) ([]*sqlc.GetProjectTeamLeaderboardRow, error)
	FindMyProjectTeamPosition(ctx context.Context, params sqlc.FindMyProjectTeamPositionParams) (*sqlc.FindMyProjectTeamPositionRow, error)
	CountProjectTeamLeaderboard(ctx context.Context, params sqlc.CountProjectTeamLeaderboardParams) (int64, error)

	GetProjectSuperTeamLeaderboard(ctx context.Context, params sqlc.GetProjectSuperTeamLeaderboardParams) ([]*sqlc.GetProjectSuperTeamLeaderboardRow, error)
	FindMyProjectSuperTeamPosition(ctx context.Context, params sqlc.FindMyProjectSuperTeamPositionParams) (*sqlc.FindMyProjectSuperTeamPositionRow, error)
	CountProjectSuperTeamLeaderboard(ctx context.Context, projectid string) (int64, error)

	GetProjectChurchLeaderboard(ctx context.Context, params sqlc.GetProjectChurchLeaderboardParams) ([]*sqlc.GetProjectChurchLeaderboardRow, error)
	FindMyProjectChurchPosition(ctx context.Context, params sqlc.FindMyProjectChurchPositionParams) (*sqlc.FindMyProjectChurchPositionRow, error)
	CountProjectChurchLeaderboard(ctx context.Context, params sqlc.CountProjectChurchLeaderboardParams) (int64, error)

	// Event leaderboards
	GetEventPersonLeaderboard(ctx context.Context, params sqlc.GetEventPersonLeaderboardParams) ([]*sqlc.GetEventPersonLeaderboardRow, error)
	FindMyEventPersonPosition(ctx context.Context, params sqlc.FindMyEventPersonPositionParams) (*sqlc.FindMyEventPersonPositionRow, error)
	CountEventPersonLeaderboard(ctx context.Context, params sqlc.CountEventPersonLeaderboardParams) (int64, error)

	GetEventTeamLeaderboard(ctx context.Context, params sqlc.GetEventTeamLeaderboardParams) ([]*sqlc.GetEventTeamLeaderboardRow, error)
	FindMyEventTeamPosition(ctx context.Context, params sqlc.FindMyEventTeamPositionParams) (*sqlc.FindMyEventTeamPositionRow, error)
	CountEventTeamLeaderboard(ctx context.Context, eventid string) (int64, error)

	GetEventSuperTeamLeaderboard(ctx context.Context, params sqlc.GetEventSuperTeamLeaderboardParams) ([]*sqlc.GetEventSuperTeamLeaderboardRow, error)
	FindMyEventSuperTeamPosition(ctx context.Context, params sqlc.FindMyEventSuperTeamPositionParams) (*sqlc.FindMyEventSuperTeamPositionRow, error)
	CountEventSuperTeamLeaderboard(ctx context.Context, eventid string) (int64, error)

	GetEventChurchLeaderboard(ctx context.Context, params sqlc.GetEventChurchLeaderboardParams) ([]*sqlc.GetEventChurchLeaderboardRow, error)
	FindMyEventChurchPosition(ctx context.Context, params sqlc.FindMyEventChurchPositionParams) (*sqlc.FindMyEventChurchPositionRow, error)
	CountEventChurchLeaderboard(ctx context.Context, params sqlc.CountEventChurchLeaderboardParams) (int64, error)
}

// LeaderboardService provides leaderboard functionality with caching
type LeaderboardService struct {
	queries LeaderboardQuerier
	cache   *cache.Cache
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(queries LeaderboardQuerier, c *cache.Cache) *LeaderboardService {
	return &LeaderboardService{
		queries: queries,
		cache:   c,
	}
}

// LeaderboardParams contains common parameters for leaderboard requests
type LeaderboardParams struct {
	ContextID  string                  // Project ID or Event ID
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
	EntityID string
	Name     string
	Image    *string
	Score    int
	Rank     int64
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
	// Build query params
	queryParams := s.buildProjectPersonParams(params)

	// Execute query
	rows, err := s.queries.GetProjectPersonLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get project person leaderboard: %w", err)
	}

	// Get total count
	countParams := s.buildProjectPersonCountParams(params)
	total, err := s.queries.CountProjectPersonLeaderboard(ctx, countParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count project person leaderboard: %w", err)
	}

	// Convert to entries
	entries := make([]LeaderboardEntry, len(rows))
	for i, row := range rows {
		if row != nil {
			entries[i] = LeaderboardEntry{
				EntityID: row.EntityID,
				Name:     row.Name,
				Image:    row.Image,
				Score:    int(row.Score),
				Rank:     row.Rank,
			}
		}
	}

	// Get "me" position
	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		posParams := s.buildProjectPersonPositionParams(params)
		myPos, err := s.queries.FindMyProjectPersonPosition(ctx, posParams)
		if err == nil && myPos != nil {
			meEntry = &LeaderboardEntry{
				EntityID: myPos.EntityID,
				Name:     myPos.Name,
				Image:    myPos.Image,
				Score:    int(myPos.Score),
				Rank:     myPos.Rank,
			}
		}
		// If error (user not found), meEntry stays nil
	}

	return entries, meEntry, int(total), nil
}

func (s *LeaderboardService) getProjectTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	queryParams := s.buildProjectTeamParams(params)

	rows, err := s.queries.GetProjectTeamLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get project team leaderboard: %w", err)
	}

	countParams := s.buildProjectTeamCountParams(params)
	total, err := s.queries.CountProjectTeamLeaderboard(ctx, countParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count project team leaderboard: %w", err)
	}

	entries := make([]LeaderboardEntry, len(rows))
	for i, row := range rows {
		if row != nil {
			entries[i] = LeaderboardEntry{
				EntityID: row.EntityID,
				Name:     row.Name,
				Image:    row.Image,
				Score:    int(row.Score),
				Rank:     row.Rank,
			}
		}
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		posParams := s.buildProjectTeamPositionParams(params)
		myPos, err := s.queries.FindMyProjectTeamPosition(ctx, posParams)
		if err == nil && myPos != nil {
			meEntry = &LeaderboardEntry{
				EntityID: myPos.EntityID,
				Name:     myPos.Name,
				Image:    myPos.Image,
				Score:    int(myPos.Score),
				Rank:     myPos.Rank,
			}
		}
	}

	return entries, meEntry, int(total), nil
}

func (s *LeaderboardService) getProjectSuperTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	queryParams := s.buildProjectSuperTeamParams(params)

	rows, err := s.queries.GetProjectSuperTeamLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get project superteam leaderboard: %w", err)
	}

	countParams := s.buildProjectSuperTeamCountParams(params)
	total, err := s.queries.CountProjectSuperTeamLeaderboard(ctx, countParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count project superteam leaderboard: %w", err)
	}

	entries := make([]LeaderboardEntry, len(rows))
	for i, row := range rows {
		if row != nil {
			entries[i] = LeaderboardEntry{
				EntityID: row.EntityID,
				Name:     row.Name,
				Image:    row.Image,
				Score:    int(row.Score),
				Rank:     row.Rank,
			}
		}
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		posParams := s.buildProjectSuperTeamPositionParams(params)
		myPos, err := s.queries.FindMyProjectSuperTeamPosition(ctx, posParams)
		if err == nil && myPos != nil {
			meEntry = &LeaderboardEntry{
				EntityID: myPos.EntityID,
				Name:     myPos.Name,
				Image:    myPos.Image,
				Score:    int(myPos.Score),
				Rank:     myPos.Rank,
			}
		}
	}

	return entries, meEntry, int(total), nil
}

func (s *LeaderboardService) getProjectChurchLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	queryParams := s.buildProjectChurchParams(params)

	rows, err := s.queries.GetProjectChurchLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get project church leaderboard: %w", err)
	}

	countParams := s.buildProjectChurchCountParams(params)
	total, err := s.queries.CountProjectChurchLeaderboard(ctx, countParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count project church leaderboard: %w", err)
	}

	entries := make([]LeaderboardEntry, len(rows))
	for i, row := range rows {
		if row != nil {
			entries[i] = LeaderboardEntry{
				EntityID: row.EntityID,
				Name:     row.Name,
				Image:    row.Image,
				Score:    int(row.Score),
				Rank:     row.Rank,
			}
		}
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		posParams := s.buildProjectChurchPositionParams(params)
		myPos, err := s.queries.FindMyProjectChurchPosition(ctx, posParams)
		if err == nil && myPos != nil {
			meEntry = &LeaderboardEntry{
				EntityID: myPos.EntityID,
				Name:     myPos.Name,
				Image:    myPos.Image,
				Score:    int(myPos.Score),
				Rank:     myPos.Rank,
			}
		}
	}

	return entries, meEntry, int(total), nil
}

// Helper functions for event leaderboards

func (s *LeaderboardService) getEventPersonLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	queryParams := s.buildEventPersonParams(params)

	rows, err := s.queries.GetEventPersonLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get event person leaderboard: %w", err)
	}

	countParams := s.buildEventPersonCountParams(params)
	total, err := s.queries.CountEventPersonLeaderboard(ctx, countParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count event person leaderboard: %w", err)
	}

	entries := make([]LeaderboardEntry, len(rows))
	for i, row := range rows {
		if row != nil {
			entries[i] = LeaderboardEntry{
				EntityID: row.EntityID,
				Name:     row.Name,
				Image:    row.Image,
				Score:    int(row.Score.(int32)),
				Rank:     row.Rank,
			}
		}
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		posParams := s.buildEventPersonPositionParams(params)
		myPos, err := s.queries.FindMyEventPersonPosition(ctx, posParams)
		if err == nil && myPos != nil {
			meEntry = &LeaderboardEntry{
				EntityID: myPos.EntityID,
				Name:     myPos.Name,
				Image:    myPos.Image,
				Score:    int(myPos.Score.(int32)),
				Rank:     myPos.Rank,
			}
		}
	}

	return entries, meEntry, int(total), nil
}

func (s *LeaderboardService) getEventTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	queryParams := s.buildEventTeamParams(params)

	rows, err := s.queries.GetEventTeamLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get event team leaderboard: %w", err)
	}

	countParams := s.buildEventTeamCountParams(params)
	total, err := s.queries.CountEventTeamLeaderboard(ctx, countParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count event team leaderboard: %w", err)
	}

	entries := make([]LeaderboardEntry, len(rows))
	for i, row := range rows {
		if row != nil {
			entries[i] = LeaderboardEntry{
				EntityID: row.EntityID,
				Name:     row.Name,
				Image:    row.Image,
				Score:    int(row.Score.(int32)),
				Rank:     row.Rank,
			}
		}
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		posParams := s.buildEventTeamPositionParams(params)
		myPos, err := s.queries.FindMyEventTeamPosition(ctx, posParams)
		if err == nil && myPos != nil {
			meEntry = &LeaderboardEntry{
				EntityID: myPos.EntityID,
				Name:     myPos.Name,
				Image:    myPos.Image,
				Score:    int(myPos.Score.(int32)),
				Rank:     myPos.Rank,
			}
		}
	}

	return entries, meEntry, int(total), nil
}

func (s *LeaderboardService) getEventSuperTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	queryParams := s.buildEventSuperTeamParams(params)

	rows, err := s.queries.GetEventSuperTeamLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get event superteam leaderboard: %w", err)
	}

	countParams := s.buildEventSuperTeamCountParams(params)
	total, err := s.queries.CountEventSuperTeamLeaderboard(ctx, countParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count event superteam leaderboard: %w", err)
	}

	entries := make([]LeaderboardEntry, len(rows))
	for i, row := range rows {
		if row != nil {
			entries[i] = LeaderboardEntry{
				EntityID: row.EntityID,
				Name:     row.Name,
				Image:    row.Image,
				Score:    int(row.Score.(int32)),
				Rank:     row.Rank,
			}
		}
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		posParams := s.buildEventSuperTeamPositionParams(params)
		myPos, err := s.queries.FindMyEventSuperTeamPosition(ctx, posParams)
		if err == nil && myPos != nil {
			meEntry = &LeaderboardEntry{
				EntityID: myPos.EntityID,
				Name:     myPos.Name,
				Image:    myPos.Image,
				Score:    int(myPos.Score.(int32)),
				Rank:     myPos.Rank,
			}
		}
	}

	return entries, meEntry, int(total), nil
}

func (s *LeaderboardService) getEventChurchLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	queryParams := s.buildEventChurchParams(params)

	rows, err := s.queries.GetEventChurchLeaderboard(ctx, queryParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get event church leaderboard: %w", err)
	}

	countParams := s.buildEventChurchCountParams(params)
	total, err := s.queries.CountEventChurchLeaderboard(ctx, countParams)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count event church leaderboard: %w", err)
	}

	entries := make([]LeaderboardEntry, len(rows))
	for i, row := range rows {
		if row != nil {
			entries[i] = LeaderboardEntry{
				EntityID: row.EntityID,
				Name:     row.Name,
				Image:    row.Image,
				Score:    int(row.Score.(int32)),
				Rank:     row.Rank,
			}
		}
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		posParams := s.buildEventChurchPositionParams(params)
		myPos, err := s.queries.FindMyEventChurchPosition(ctx, posParams)
		if err == nil && myPos != nil {
			meEntry = &LeaderboardEntry{
				EntityID: myPos.EntityID,
				Name:     myPos.Name,
				Image:    myPos.Image,
				Score:    int(myPos.Score.(int32)),
				Rank:     myPos.Rank,
			}
		}
	}

	return entries, meEntry, int(total), nil
}

// Parameter builders for project queries

func (s *LeaderboardService) buildProjectPersonParams(params LeaderboardParams) sqlc.GetProjectPersonLeaderboardParams {
	limit := 10
	if params.First != nil {
		limit = *params.First + 1 // Request one extra to check hasNext
	} else if params.Last != nil {
		limit = *params.Last + 1
	}

	var afterRank *int64
	if params.After != nil && *params.After != "" {
		if rank, err := strconv.ParseInt(*params.After, 10, 64); err == nil {
			afterRank = &rank
		}
	}

	var beforeRank *int64
	if params.Before != nil && *params.Before != "" {
		if rank, err := strconv.ParseInt(*params.Before, 10, 64); err == nil {
			beforeRank = &rank
		}
	}

	return sqlc.GetProjectPersonLeaderboardParams{
		Projectid: params.ContextID,
		Churchid: getFilterString(params.Filter, "churchId"),
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
		Gender: getFilterString(params.Filter, "gender"),
		Minage: getFilterInt(params.Filter, "minAge"),
		Maxage: getFilterInt(params.Filter, "maxAge"),
		Teamid: getFilterString(params.Filter, "teamId"),
		Superteamid: getFilterString(params.Filter, "superTeamId"),
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
		Afterrank: afterRank,
		Beforerank: beforeRank,
		Querylimit: int32(limit),
	}
}

func (s *LeaderboardService) buildProjectPersonCountParams(params LeaderboardParams) sqlc.CountProjectPersonLeaderboardParams {
	return sqlc.CountProjectPersonLeaderboardParams{
		Projectid: params.ContextID,
		Churchid: getFilterString(params.Filter, "churchId"),
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
		Gender: getFilterString(params.Filter, "gender"),
		Minage: getFilterInt(params.Filter, "minAge"),
		Maxage: getFilterInt(params.Filter, "maxAge"),
		Teamid: getFilterString(params.Filter, "teamId"),
		Superteamid: getFilterString(params.Filter, "superTeamId"),
	}
}

func (s *LeaderboardService) buildProjectPersonPositionParams(params LeaderboardParams) sqlc.FindMyProjectPersonPositionParams {
	return sqlc.FindMyProjectPersonPositionParams{
		Projectid: params.ContextID,
		Userid: params.UserID,
		Churchid: getFilterString(params.Filter, "churchId"),
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
		Gender: getFilterString(params.Filter, "gender"),
		Minage: getFilterInt(params.Filter, "minAge"),
		Maxage: getFilterInt(params.Filter, "maxAge"),
		Teamid: getFilterString(params.Filter, "teamId"),
		Superteamid: getFilterString(params.Filter, "superTeamId"),
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
	}
}

func (s *LeaderboardService) buildProjectTeamParams(params LeaderboardParams) sqlc.GetProjectTeamLeaderboardParams {
	limit := 10
	if params.First != nil {
		limit = *params.First + 1
	} else if params.Last != nil {
		limit = *params.Last + 1
	}

	var afterRank *int64
	if params.After != nil && *params.After != "" {
		if rank, err := strconv.ParseInt(*params.After, 10, 64); err == nil {
			afterRank = &rank
		}
	}

	var beforeRank *int64
	if params.Before != nil && *params.Before != "" {
		if rank, err := strconv.ParseInt(*params.Before, 10, 64); err == nil {
			beforeRank = &rank
		}
	}

	return sqlc.GetProjectTeamLeaderboardParams{
		Projectid: params.ContextID,
		Superteamid: getFilterString(params.Filter, "superTeamId"),
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
		Afterrank: afterRank,
		Beforerank: beforeRank,
		Querylimit: int32(limit),
	}
}

func (s *LeaderboardService) buildProjectTeamCountParams(params LeaderboardParams) sqlc.CountProjectTeamLeaderboardParams {
	return sqlc.CountProjectTeamLeaderboardParams{
		Projectid: params.ContextID,
		Superteamid: getFilterString(params.Filter, "superTeamId"),
	}
}

func (s *LeaderboardService) buildProjectTeamPositionParams(params LeaderboardParams) sqlc.FindMyProjectTeamPositionParams {
	return sqlc.FindMyProjectTeamPositionParams{
		Projectid: params.ContextID,
		Userid: params.UserID,
		Superteamid: getFilterString(params.Filter, "superTeamId"),
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
	}
}

func (s *LeaderboardService) buildProjectSuperTeamParams(params LeaderboardParams) sqlc.GetProjectSuperTeamLeaderboardParams {
	limit := 10
	if params.First != nil {
		limit = *params.First + 1
	} else if params.Last != nil {
		limit = *params.Last + 1
	}

	var afterRank *int64
	if params.After != nil && *params.After != "" {
		if rank, err := strconv.ParseInt(*params.After, 10, 64); err == nil {
			afterRank = &rank
		}
	}

	var beforeRank *int64
	if params.Before != nil && *params.Before != "" {
		if rank, err := strconv.ParseInt(*params.Before, 10, 64); err == nil {
			beforeRank = &rank
		}
	}

	return sqlc.GetProjectSuperTeamLeaderboardParams{
		Projectid: params.ContextID,
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
		Afterrank: afterRank,
		Beforerank: beforeRank,
		Querylimit: int32(limit),
	}
}

func (s *LeaderboardService) buildProjectSuperTeamCountParams(params LeaderboardParams) string {
	return params.ContextID
}

func (s *LeaderboardService) buildProjectSuperTeamPositionParams(params LeaderboardParams) sqlc.FindMyProjectSuperTeamPositionParams {
	return sqlc.FindMyProjectSuperTeamPositionParams{
		Projectid: params.ContextID,
		Userid: params.UserID,
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
	}
}

func (s *LeaderboardService) buildProjectChurchParams(params LeaderboardParams) sqlc.GetProjectChurchLeaderboardParams {
	limit := 10
	if params.First != nil {
		limit = *params.First + 1
	} else if params.Last != nil {
		limit = *params.Last + 1
	}

	var afterRank *int64
	if params.After != nil && *params.After != "" {
		if rank, err := strconv.ParseInt(*params.After, 10, 64); err == nil {
			afterRank = &rank
		}
	}

	var beforeRank *int64
	if params.Before != nil && *params.Before != "" {
		if rank, err := strconv.ParseInt(*params.Before, 10, 64); err == nil {
			beforeRank = &rank
		}
	}

	return sqlc.GetProjectChurchLeaderboardParams{
		Projectid: params.ContextID,
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
		Afterrank: afterRank,
		Beforerank: beforeRank,
		Querylimit: int32(limit),
	}
}

func (s *LeaderboardService) buildProjectChurchCountParams(params LeaderboardParams) sqlc.CountProjectChurchLeaderboardParams {
	return sqlc.CountProjectChurchLeaderboardParams{
		Projectid: params.ContextID,
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
	}
}

func (s *LeaderboardService) buildProjectChurchPositionParams(params LeaderboardParams) sqlc.FindMyProjectChurchPositionParams {
	return sqlc.FindMyProjectChurchPositionParams{
		Projectid: params.ContextID,
		Userid: params.UserID,
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
	}
}

// Parameter builders for event queries

func (s *LeaderboardService) buildEventPersonParams(params LeaderboardParams) sqlc.GetEventPersonLeaderboardParams {
	limit := 10
	if params.First != nil {
		limit = *params.First + 1
	} else if params.Last != nil {
		limit = *params.Last + 1
	}

	var afterRank *int64
	if params.After != nil && *params.After != "" {
		if rank, err := strconv.ParseInt(*params.After, 10, 64); err == nil {
			afterRank = &rank
		}
	}

	var beforeRank *int64
	if params.Before != nil && *params.Before != "" {
		if rank, err := strconv.ParseInt(*params.Before, 10, 64); err == nil {
			beforeRank = &rank
		}
	}

	return sqlc.GetEventPersonLeaderboardParams{
		Eventid: params.ContextID,
		Churchid: getFilterString(params.Filter, "churchId"),
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
		Gender: getFilterString(params.Filter, "gender"),
		Minage: getFilterInt(params.Filter, "minAge"),
		Maxage: getFilterInt(params.Filter, "maxAge"),
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
		Afterrank: afterRank,
		Beforerank: beforeRank,
		Querylimit: int32(limit),
	}
}

func (s *LeaderboardService) buildEventPersonCountParams(params LeaderboardParams) sqlc.CountEventPersonLeaderboardParams {
	return sqlc.CountEventPersonLeaderboardParams{
		Eventid: params.ContextID,
		Churchid: getFilterString(params.Filter, "churchId"),
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
		Gender: getFilterString(params.Filter, "gender"),
		Minage: getFilterInt(params.Filter, "minAge"),
		Maxage: getFilterInt(params.Filter, "maxAge"),
	}
}

func (s *LeaderboardService) buildEventPersonPositionParams(params LeaderboardParams) sqlc.FindMyEventPersonPositionParams {
	return sqlc.FindMyEventPersonPositionParams{
		Eventid: params.ContextID,
		Userid: params.UserID,
		Churchid: getFilterString(params.Filter, "churchId"),
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
		Gender: getFilterString(params.Filter, "gender"),
		Minage: getFilterInt(params.Filter, "minAge"),
		Maxage: getFilterInt(params.Filter, "maxAge"),
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
	}
}

func (s *LeaderboardService) buildEventTeamParams(params LeaderboardParams) sqlc.GetEventTeamLeaderboardParams {
	limit := 10
	if params.First != nil {
		limit = *params.First + 1
	} else if params.Last != nil {
		limit = *params.Last + 1
	}

	var afterRank *int64
	if params.After != nil && *params.After != "" {
		if rank, err := strconv.ParseInt(*params.After, 10, 64); err == nil {
			afterRank = &rank
		}
	}

	var beforeRank *int64
	if params.Before != nil && *params.Before != "" {
		if rank, err := strconv.ParseInt(*params.Before, 10, 64); err == nil {
			beforeRank = &rank
		}
	}

	return sqlc.GetEventTeamLeaderboardParams{
		Eventid: params.ContextID,
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
		Afterrank: afterRank,
		Beforerank: beforeRank,
		Querylimit: int32(limit),
	}
}

func (s *LeaderboardService) buildEventTeamCountParams(params LeaderboardParams) string {
	return params.ContextID
}

func (s *LeaderboardService) buildEventTeamPositionParams(params LeaderboardParams) sqlc.FindMyEventTeamPositionParams {
	return sqlc.FindMyEventTeamPositionParams{
		Eventid: params.ContextID,
		Userid: params.UserID,
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
	}
}

func (s *LeaderboardService) buildEventSuperTeamParams(params LeaderboardParams) sqlc.GetEventSuperTeamLeaderboardParams {
	limit := 10
	if params.First != nil {
		limit = *params.First + 1
	} else if params.Last != nil {
		limit = *params.Last + 1
	}

	var afterRank *int64
	if params.After != nil && *params.After != "" {
		if rank, err := strconv.ParseInt(*params.After, 10, 64); err == nil {
			afterRank = &rank
		}
	}

	var beforeRank *int64
	if params.Before != nil && *params.Before != "" {
		if rank, err := strconv.ParseInt(*params.Before, 10, 64); err == nil {
			beforeRank = &rank
		}
	}

	return sqlc.GetEventSuperTeamLeaderboardParams{
		Eventid: params.ContextID,
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
		Afterrank: afterRank,
		Beforerank: beforeRank,
		Querylimit: int32(limit),
	}
}

func (s *LeaderboardService) buildEventSuperTeamCountParams(params LeaderboardParams) string {
	return params.ContextID
}

func (s *LeaderboardService) buildEventSuperTeamPositionParams(params LeaderboardParams) sqlc.FindMyEventSuperTeamPositionParams {
	return sqlc.FindMyEventSuperTeamPositionParams{
		Eventid: params.ContextID,
		Userid: params.UserID,
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
	}
}

func (s *LeaderboardService) buildEventChurchParams(params LeaderboardParams) sqlc.GetEventChurchLeaderboardParams {
	limit := 10
	if params.First != nil {
		limit = *params.First + 1
	} else if params.Last != nil {
		limit = *params.Last + 1
	}

	var afterRank *int64
	if params.After != nil && *params.After != "" {
		if rank, err := strconv.ParseInt(*params.After, 10, 64); err == nil {
			afterRank = &rank
		}
	}

	var beforeRank *int64
	if params.Before != nil && *params.Before != "" {
		if rank, err := strconv.ParseInt(*params.Before, 10, 64); err == nil {
			beforeRank = &rank
		}
	}

	return sqlc.GetEventChurchLeaderboardParams{
		Eventid: params.ContextID,
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
		Minscore: getFilterInt(params.Filter, "minScore"),
		Maxscore: getFilterInt(params.Filter, "maxScore"),
		Afterrank: afterRank,
		Beforerank: beforeRank,
		Querylimit: int32(limit),
	}
}

func (s *LeaderboardService) buildEventChurchCountParams(params LeaderboardParams) sqlc.CountEventChurchLeaderboardParams {
	return sqlc.CountEventChurchLeaderboardParams{
		Eventid: params.ContextID,
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
	}
}

func (s *LeaderboardService) buildEventChurchPositionParams(params LeaderboardParams) sqlc.FindMyEventChurchPositionParams {
	return sqlc.FindMyEventChurchPositionParams{
		Eventid: params.ContextID,
		Userid: params.UserID,
		Country: getFilterString(params.Filter, "country"),
		Churchcategory: getFilterString(params.Filter, "churchCategory"),
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
