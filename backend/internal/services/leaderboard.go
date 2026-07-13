package services

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/utils"
	"golang.org/x/sync/singleflight"
)

// leaderboardFetchTimeout bounds the detached singleflight fetch so an
// abandoned request can't hold a DB connection indefinitely.
const leaderboardFetchTimeout = 30 * time.Second

// LeaderboardQuerier defines the database operations needed for leaderboards
type LeaderboardQuerier interface {
	// Project leaderboards
	GetProjectPersonLeaderboard(ctx context.Context, params sqlc.GetProjectPersonLeaderboardParams) ([]*sqlc.GetProjectPersonLeaderboardRow, error)
	FindMyProjectPersonPosition(ctx context.Context, params sqlc.FindMyProjectPersonPositionParams) (*sqlc.FindMyProjectPersonPositionRow, error)
	CountProjectPersonLeaderboard(ctx context.Context, params sqlc.CountProjectPersonLeaderboardParams) (int64, error)
	GetFullProjectPersonLeaderboard(ctx context.Context, params sqlc.GetFullProjectPersonLeaderboardParams) ([]*sqlc.GetFullProjectPersonLeaderboardRow, error)

	GetProjectTeamLeaderboard(ctx context.Context, params sqlc.GetProjectTeamLeaderboardParams) ([]*sqlc.GetProjectTeamLeaderboardRow, error)
	FindMyProjectTeamPosition(ctx context.Context, params sqlc.FindMyProjectTeamPositionParams) (*sqlc.FindMyProjectTeamPositionRow, error)
	CountProjectTeamLeaderboard(ctx context.Context, arg sqlc.CountProjectTeamLeaderboardParams) (int64, error)
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
	CountEventTeamLeaderboard(ctx context.Context, arg sqlc.CountEventTeamLeaderboardParams) (int64, error)
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
	cache   *cache.CacheWithRegistry
	loaders *loaders.Loaders
	flight  singleflight.Group
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(queries LeaderboardQuerier, c *cache.CacheWithRegistry, l *loaders.Loaders) *LeaderboardService {
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

// cachedLeaderboard is a decoded full leaderboard as stored in Ristretto.
// Entries and IndexByEntityID are shared across requests — treat as read-only.
type cachedLeaderboard struct {
	Entries         []LeaderboardEntry
	IndexByEntityID map[string]int
}

// findMe returns a copy of the entry for entityID, or nil if absent.
func (c *cachedLeaderboard) findMe(entityID string) *LeaderboardEntry {
	if entityID == "" {
		return nil
	}
	if i, ok := c.IndexByEntityID[entityID]; ok {
		entry := c.Entries[i]
		return &entry
	}
	return nil
}

// getFullLeaderboardCached returns the full leaderboard for cacheKey, serving
// the decoded form from cache when possible. Concurrent misses for the same
// key share a single fetch via singleflight, so an expired hot key does not
// stampede the database.
func (s *LeaderboardService) getFullLeaderboardCached(
	ctx context.Context,
	cacheKey string,
	fetch func(context.Context) ([]LeaderboardEntry, error),
) (*cachedLeaderboard, error) {
	if cachedData, found := s.cache.Get(cacheKey); found {
		if board, ok := cachedData.(*cachedLeaderboard); ok {
			return board, nil
		}
	}

	v, err, _ := s.flight.Do(cacheKey, func() (interface{}, error) {
		// Another flight may have populated the cache while we queued
		if cachedData, found := s.cache.Get(cacheKey); found {
			if board, ok := cachedData.(*cachedLeaderboard); ok {
				return board, nil
			}
		}

		// Detach from the leading request's cancellation: the result is
		// shared by every request waiting on this key. Still bound by a
		// service-owned timeout so an abandoned fetch can't hold a DB
		// connection indefinitely.
		fetchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), leaderboardFetchTimeout)
		defer cancel()
		entries, err := fetch(fetchCtx)
		if err != nil {
			return nil, err
		}

		board := &cachedLeaderboard{
			Entries:         entries,
			IndexByEntityID: make(map[string]int, len(entries)),
		}
		for i, e := range entries {
			board.IndexByEntityID[e.EntityID] = i
		}

		// Cache the full leaderboard with 5 minute TTL
		s.cache.SetWithTTL(cacheKey, board, 5*time.Minute)
		return board, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*cachedLeaderboard), nil
}

// meTeamIDInProject resolves the user's team ID within a project ("" if none)
func (s *LeaderboardService) meTeamIDInProject(ctx context.Context, userID, projectID string) string {
	teams, err := s.loaders.TeamsByUserLoader.Load(ctx, userID)()
	if err != nil {
		return ""
	}
	for _, team := range teams {
		if team.ProjectID == projectID {
			return team.ID
		}
	}
	return ""
}

// meSuperTeamIDInProject resolves the user's superteam ID within a project ("" if none)
func (s *LeaderboardService) meSuperTeamIDInProject(ctx context.Context, userID, projectID string) string {
	superTeams, err := s.loaders.SuperTeamsByUserLoader.Load(ctx, userID)()
	if err != nil {
		return ""
	}
	for _, superTeam := range superTeams {
		if superTeam.ProjectID == projectID {
			return superTeam.ID
		}
	}
	return ""
}

// meChurchID resolves the user's church ID ("" if the user cannot be loaded)
func (s *LeaderboardService) meChurchID(ctx context.Context, userID string) string {
	user, err := s.loaders.UserByIDLoader.Load(ctx, userID)()
	if err != nil || user == nil {
		return ""
	}
	return user.ChurchID
}

// eventProjectID resolves the project an event belongs to ("" if the event cannot be loaded)
func (s *LeaderboardService) eventProjectID(ctx context.Context, eventID string) string {
	event, err := s.loaders.EventByIDLoader.Load(ctx, eventID)()
	if err != nil || event == nil {
		return ""
	}
	return event.ProjectID
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
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("project", params.ContextID, "persons", filterParams)

	board, err := s.getFullLeaderboardCached(ctx, cacheKey, func(ctx context.Context) ([]LeaderboardEntry, error) {
		rows, err := s.queries.GetFullProjectPersonLeaderboard(ctx, s.buildFullProjectPersonParams(params))
		if err != nil {
			return nil, fmt.Errorf("failed to get full project person leaderboard: %w", err)
		}
		entries := make([]LeaderboardEntry, 0, len(rows))
		for _, row := range rows {
			if row != nil {
				entries = append(entries, LeaderboardEntry{
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
		return entries, nil
	})
	if err != nil {
		return nil, nil, 0, err
	}

	meEntry := board.findMe(params.UserID)
	paginated := paginateLeaderboard(board.Entries, params.First, params.After, params.Last, params.Before)
	return paginated, meEntry, len(board.Entries), nil
}

func (s *LeaderboardService) getProjectTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("project", params.ContextID, "teams", filterParams)

	board, err := s.getFullLeaderboardCached(ctx, cacheKey, func(ctx context.Context) ([]LeaderboardEntry, error) {
		rows, err := s.queries.GetFullProjectTeamLeaderboard(ctx, s.buildFullProjectTeamParams(params))
		if err != nil {
			return nil, fmt.Errorf("failed to get full project team leaderboard: %w", err)
		}
		entries := make([]LeaderboardEntry, 0, len(rows))
		for _, row := range rows {
			if row != nil {
				entries = append(entries, LeaderboardEntry{
					EntityID:    row.EntityID,
					Name:        row.Name,
					Image:       row.Image,
					Score:       int(row.Score),
					Rank:        row.Rank,
					LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
				})
			}
		}
		return entries, nil
	})
	if err != nil {
		return nil, nil, 0, err
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		meEntry = board.findMe(s.meTeamIDInProject(ctx, params.UserID, params.ContextID))
	}
	paginated := paginateLeaderboard(board.Entries, params.First, params.After, params.Last, params.Before)
	return paginated, meEntry, len(board.Entries), nil
}

func (s *LeaderboardService) getProjectSuperTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("project", params.ContextID, "superteams", filterParams)

	board, err := s.getFullLeaderboardCached(ctx, cacheKey, func(ctx context.Context) ([]LeaderboardEntry, error) {
		rows, err := s.queries.GetFullProjectSuperTeamLeaderboard(ctx, s.buildFullProjectSuperTeamParams(params))
		if err != nil {
			return nil, fmt.Errorf("failed to get full project superteam leaderboard: %w", err)
		}
		entries := make([]LeaderboardEntry, 0, len(rows))
		for _, row := range rows {
			if row != nil {
				entries = append(entries, LeaderboardEntry{
					EntityID:    row.EntityID,
					Name:        row.Name,
					Image:       row.Image,
					Score:       int(row.Score),
					Rank:        row.Rank,
					LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
				})
			}
		}
		return entries, nil
	})
	if err != nil {
		return nil, nil, 0, err
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		meEntry = board.findMe(s.meSuperTeamIDInProject(ctx, params.UserID, params.ContextID))
	}
	paginated := paginateLeaderboard(board.Entries, params.First, params.After, params.Last, params.Before)
	return paginated, meEntry, len(board.Entries), nil
}

func (s *LeaderboardService) getProjectChurchLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("project", params.ContextID, "churches", filterParams)

	board, err := s.getFullLeaderboardCached(ctx, cacheKey, func(ctx context.Context) ([]LeaderboardEntry, error) {
		rows, err := s.queries.GetFullProjectChurchLeaderboard(ctx, s.buildFullProjectChurchParams(params))
		if err != nil {
			return nil, fmt.Errorf("failed to get full project church leaderboard: %w", err)
		}
		entries := make([]LeaderboardEntry, 0, len(rows))
		for _, row := range rows {
			if row != nil {
				entries = append(entries, LeaderboardEntry{
					EntityID:    row.EntityID,
					Name:        row.Name,
					Image:       row.Image,
					Score:       int(row.Score),
					Rank:        row.Rank,
					LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
				})
			}
		}
		return entries, nil
	})
	if err != nil {
		return nil, nil, 0, err
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		meEntry = board.findMe(s.meChurchID(ctx, params.UserID))
	}
	paginated := paginateLeaderboard(board.Entries, params.First, params.After, params.Last, params.Before)
	return paginated, meEntry, len(board.Entries), nil
}

// Helper functions for event leaderboards

func (s *LeaderboardService) getEventPersonLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("event", params.ContextID, "persons", filterParams)

	board, err := s.getFullLeaderboardCached(ctx, cacheKey, func(ctx context.Context) ([]LeaderboardEntry, error) {
		rows, err := s.queries.GetFullEventPersonLeaderboard(ctx, s.buildFullEventPersonParams(params))
		if err != nil {
			return nil, fmt.Errorf("failed to get full event person leaderboard: %w", err)
		}
		entries := make([]LeaderboardEntry, 0, len(rows))
		for _, row := range rows {
			if row != nil {
				entries = append(entries, LeaderboardEntry{
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
		return entries, nil
	})
	if err != nil {
		return nil, nil, 0, err
	}

	meEntry := board.findMe(params.UserID)
	paginated := paginateLeaderboard(board.Entries, params.First, params.After, params.Last, params.Before)
	return paginated, meEntry, len(board.Entries), nil
}

func (s *LeaderboardService) getEventTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("event", params.ContextID, "teams", filterParams)

	board, err := s.getFullLeaderboardCached(ctx, cacheKey, func(ctx context.Context) ([]LeaderboardEntry, error) {
		rows, err := s.queries.GetFullEventTeamLeaderboard(ctx, s.buildFullEventTeamParams(params))
		if err != nil {
			return nil, fmt.Errorf("failed to get full event team leaderboard: %w", err)
		}
		entries := make([]LeaderboardEntry, 0, len(rows))
		for _, row := range rows {
			if row != nil {
				entries = append(entries, LeaderboardEntry{
					EntityID:    row.EntityID,
					Name:        row.Name,
					Image:       row.Image,
					Score:       int(row.Score),
					Rank:        row.Rank,
					LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
				})
			}
		}
		return entries, nil
	})
	if err != nil {
		return nil, nil, 0, err
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		if projectID := s.eventProjectID(ctx, params.ContextID); projectID != "" {
			meEntry = board.findMe(s.meTeamIDInProject(ctx, params.UserID, projectID))
		}
	}
	paginated := paginateLeaderboard(board.Entries, params.First, params.After, params.Last, params.Before)
	return paginated, meEntry, len(board.Entries), nil
}

func (s *LeaderboardService) getEventSuperTeamLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("event", params.ContextID, "superteams", filterParams)

	board, err := s.getFullLeaderboardCached(ctx, cacheKey, func(ctx context.Context) ([]LeaderboardEntry, error) {
		rows, err := s.queries.GetFullEventSuperTeamLeaderboard(ctx, s.buildFullEventSuperTeamParams(params))
		if err != nil {
			return nil, fmt.Errorf("failed to get full event superteam leaderboard: %w", err)
		}
		entries := make([]LeaderboardEntry, 0, len(rows))
		for _, row := range rows {
			if row != nil {
				entries = append(entries, LeaderboardEntry{
					EntityID:    row.EntityID,
					Name:        row.Name,
					Image:       row.Image,
					Score:       int(row.Score),
					Rank:        row.Rank,
					LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
				})
			}
		}
		return entries, nil
	})
	if err != nil {
		return nil, nil, 0, err
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		if projectID := s.eventProjectID(ctx, params.ContextID); projectID != "" {
			meEntry = board.findMe(s.meSuperTeamIDInProject(ctx, params.UserID, projectID))
		}
	}
	paginated := paginateLeaderboard(board.Entries, params.First, params.After, params.Last, params.Before)
	return paginated, meEntry, len(board.Entries), nil
}

func (s *LeaderboardService) getEventChurchLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, *LeaderboardEntry, int, error) {
	filterParams := buildFilterParamsMap(params.Filter)
	cacheKey := cache.FullLeaderboardKey("event", params.ContextID, "churches", filterParams)

	board, err := s.getFullLeaderboardCached(ctx, cacheKey, func(ctx context.Context) ([]LeaderboardEntry, error) {
		rows, err := s.queries.GetFullEventChurchLeaderboard(ctx, s.buildFullEventChurchParams(params))
		if err != nil {
			return nil, fmt.Errorf("failed to get full event church leaderboard: %w", err)
		}
		entries := make([]LeaderboardEntry, 0, len(rows))
		for _, row := range rows {
			if row != nil {
				entries = append(entries, LeaderboardEntry{
					EntityID:    row.EntityID,
					Name:        row.Name,
					Image:       row.Image,
					Score:       int(row.Score),
					Rank:        row.Rank,
					LastScoreAt: utils.TimestamptzToPtr(row.LastScoreAt),
				})
			}
		}
		return entries, nil
	})
	if err != nil {
		return nil, nil, 0, err
	}

	var meEntry *LeaderboardEntry
	if params.UserID != "" {
		meEntry = board.findMe(s.meChurchID(ctx, params.UserID))
	}
	paginated := paginateLeaderboard(board.Entries, params.First, params.After, params.Last, params.Before)
	return paginated, meEntry, len(board.Entries), nil
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
		Churchid:  getFilterString(params.Filter, "churchId"),
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
		Churchid: getFilterString(params.Filter, "churchId"),
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
