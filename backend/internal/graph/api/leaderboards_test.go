package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPersonLeaderboardsRespectPaginationRegardlessOfSize(t *testing.T) {
	const projectID = "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const eventID = "EV01ARZ3NDEKTSV4RRFFQ69G5FAV"

	for _, source := range []string{"project", "event", "project config", "event config"} {
		t.Run(source, func(t *testing.T) {
			for _, tt := range []struct {
				name       string
				totalCount int
				first      *int
				after      *string
				last       *int
				before     *string
				startRank  int
				count      int
			}{
				{name: "small board", totalCount: 19, first: intPtr(100), startRank: 1, count: 19},
				{name: "medium board", totalCount: 20, first: intPtr(100), startRank: 1, count: 20},
				{name: "large board", totalCount: 50, first: intPtr(100), startRank: 1, count: 50},
				{name: "requested page size", totalCount: 50, first: intPtr(5), startRank: 1, count: 5},
				{name: "after old small cap", totalCount: 19, first: intPtr(5), after: stringPtr("3"), startRank: 4, count: 5},
				{name: "after old medium cap", totalCount: 20, first: intPtr(5), after: stringPtr("10"), startRank: 11, count: 5},
				{name: "after old large cap", totalCount: 50, first: intPtr(5), after: stringPtr("20"), startRank: 21, count: 5},
				{name: "backward beyond old cap", totalCount: 50, last: intPtr(5), before: stringPtr("30"), startRank: 25, count: 5},
				{name: "default page size", totalCount: 19, startRank: 1, count: 10},
			} {
				t.Run(tt.name, func(t *testing.T) {
					queries := mocks.NewMockLeaderboardQuerier(t)
					c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
					require.NoError(t, err)
					t.Cleanup(c.Close)

					projectRows := make([]*sqlc.GetFullProjectPersonLeaderboardRow, tt.totalCount)
					eventRows := make([]*sqlc.GetFullEventPersonLeaderboardRow, tt.totalCount)
					for i := range projectRows {
						projectRows[i] = &sqlc.GetFullProjectPersonLeaderboardRow{
							EntityID: fmt.Sprintf("US%026d", i+1),
							Name:     fmt.Sprintf("User %d", i+1),
							Score:    int64(tt.totalCount - i),
							Rank:     int64(i + 1),
						}
						eventRows[i] = (*sqlc.GetFullEventPersonLeaderboardRow)(projectRows[i])
					}
					isEvent := source == "event" || source == "event config"
					if isEvent {
						queries.On("GetFullEventPersonLeaderboard", mock.Anything, mock.Anything).Return(eventRows, nil).Once()
					} else {
						queries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).Return(projectRows, nil).Once()
					}

					userID := projectRows[tt.totalCount-1].EntityID
					ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)
					r := &Resolver{LeaderboardService: services.NewLeaderboardService(queries, c, nil)}
					var connection *model.LeaderboardConnection
					switch source {
					case "project":
						connection, err = r.Project().Leaderboard(ctx, &model.Project{ID: projectID}, model.LeaderboardEntityTypePersons, nil, tt.first, tt.after, tt.last, tt.before)
					case "event":
						connection, err = r.Event().Leaderboard(ctx, &model.Event{ID: eventID, ProjectID: projectID}, model.LeaderboardEntityTypePersons, nil, tt.first, tt.after, tt.last, tt.before)
					default:
						config := &model.LeaderboardConfig{ProjectID: projectID, EntityType: model.LeaderboardEntityTypePersons}
						if isEvent {
							config.EventID = stringPtr(eventID)
						}
						connection, err = r.LeaderboardConfig().Leaderboard(ctx, config, tt.first, tt.after, tt.last, tt.before)
					}
					require.NoError(t, err)
					count := tt.count
					if tt.name == "default page size" && (source == "project config" || source == "event config") {
						count = tt.totalCount
					}
					require.Len(t, connection.Edges, count)
					assert.Equal(t, tt.totalCount, connection.TotalCount)
					for i, edge := range connection.Edges {
						assert.Equal(t, tt.startRank+i, *edge.Node.Rank)
					}
					assert.Equal(t, fmt.Sprint(tt.startRank), *connection.PageInfo.StartCursor)
					assert.Equal(t, fmt.Sprint(tt.startRank+count-1), *connection.PageInfo.EndCursor)
					require.NotNil(t, connection.Me)
					assert.Equal(t, userID, connection.Me.ID)
					assert.Equal(t, tt.totalCount, *connection.Me.Rank)
				})
			}
		})
	}
}

func TestMarshalLeaderboardFilter_Nil(t *testing.T) {
	b, err := marshalLeaderboardFilter(nil)
	require.NoError(t, err)
	assert.Nil(t, b)
}

func TestMarshalLeaderboardFilter_ValidFilter(t *testing.T) {
	minScore := 10
	churchID := "CH01ARZ3NDEKTSV4RRFFQ69G5FAV"
	filter := &model.LeaderboardFilter{
		MinScore: &minScore,
		ChurchID: &churchID,
	}

	b, err := marshalLeaderboardFilter(filter)
	require.NoError(t, err)
	require.NotNil(t, b)
	assert.Contains(t, string(b), `"minScore":10`)
	assert.Contains(t, string(b), churchID)
}

func TestFilterConfigsByVisibility_NonAdminSeesOnlyActive(t *testing.T) {
	configs := []*model.LeaderboardConfig{
		{ID: "LC1", IsActive: true},
		{ID: "LC2", IsActive: false},
		{ID: "LC3", IsActive: true},
	}

	result := filterConfigsByVisibility(configs, false)

	require.Len(t, result, 2)
	assert.Equal(t, "LC1", result[0].ID)
	assert.Equal(t, "LC3", result[1].ID)
}

func TestFilterConfigsByVisibility_AdminSeesAll(t *testing.T) {
	configs := []*model.LeaderboardConfig{
		{ID: "LC1", IsActive: true},
		{ID: "LC2", IsActive: false},
	}

	result := filterConfigsByVisibility(configs, true)

	require.Len(t, result, 2)
}

func TestFilterConfigsByVisibility_EmptyInput(t *testing.T) {
	result := filterConfigsByVisibility(nil, false)
	assert.Empty(t, result)
}

func TestBuildLeaderboardParamsFromConfig_ProjectScoped(t *testing.T) {
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	config := &model.LeaderboardConfig{
		ProjectID:  projectID,
		EventID:    nil,
		EntityType: model.LeaderboardEntityTypePersons,
	}

	params, isEvent := buildLeaderboardParamsFromConfig(config, nil, nil, nil, nil, "US01ARZ3NDEKTSV4RRFFQ69G5FAV")

	assert.False(t, isEvent)
	assert.Equal(t, projectID, params.ContextID)
	assert.Equal(t, model.LeaderboardEntityTypePersons, params.EntityType)
	assert.Equal(t, "US01ARZ3NDEKTSV4RRFFQ69G5FAV", params.UserID)
	assert.Nil(t, params.Filter)
}

func TestBuildLeaderboardParamsFromConfig_EventScoped(t *testing.T) {
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	eventID := "EV01ARZ3NDEKTSV4RRFFQ69G5FAV"
	config := &model.LeaderboardConfig{
		ProjectID:  projectID,
		EventID:    &eventID,
		EntityType: model.LeaderboardEntityTypeTeams,
	}

	params, isEvent := buildLeaderboardParamsFromConfig(config, nil, nil, nil, nil, "US01ARZ3NDEKTSV4RRFFQ69G5FAV")

	assert.True(t, isEvent)
	assert.Equal(t, eventID, params.ContextID)
}

func TestBuildLeaderboardParamsFromConfig_ConvertsFilterView(t *testing.T) {
	minScore := 42
	config := &model.LeaderboardConfig{
		ProjectID:  "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		EntityType: model.LeaderboardEntityTypePersons,
		Filter:     &model.LeaderboardFilterView{MinScore: &minScore},
	}

	params, _ := buildLeaderboardParamsFromConfig(config, nil, nil, nil, nil, "US01ARZ3NDEKTSV4RRFFQ69G5FAV")

	require.NotNil(t, params.Filter)
	require.NotNil(t, params.Filter.MinScore)
	assert.Equal(t, 42, *params.Filter.MinScore)
}

func TestFilterViewToFilter_Nil(t *testing.T) {
	assert.Nil(t, filterViewToFilter(nil))
}

func TestFilterViewToFilter_ConvertsAgeRange(t *testing.T) {
	view := &model.LeaderboardFilterView{
		AgeRange: &model.AgeRange{Min: 10, Max: 20},
	}

	filter := filterViewToFilter(view)

	require.NotNil(t, filter.AgeRange)
	assert.Equal(t, 10, filter.AgeRange.Min)
	assert.Equal(t, 20, filter.AgeRange.Max)
}

func TestBuildLeaderboardConfigFilterParamsCursor_AppliesFilterFields(t *testing.T) {
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	isActive := true
	filter := &model.LeaderboardConfigFilter{
		ProjectID: &projectID,
		IsActive:  &isActive,
	}

	params, err := buildLeaderboardConfigFilterParamsCursor(filter, nil, nil, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, projectID, params.Projectid)
	require.NotNil(t, params.Isactive)
	assert.True(t, *params.Isactive)
	assert.Equal(t, int32(11), params.Querylimit) // default page size + 1
	assert.False(t, params.Isbackward)
}

func TestBuildLeaderboardConfigFilterParamsCursor_FirstAndLastMutuallyExclusive(t *testing.T) {
	first := 5
	last := 5

	_, err := buildLeaderboardConfigFilterParamsCursor(nil, &first, nil, &last, nil)

	assert.Error(t, err)
}

func TestBuildLeaderboardConfigFilterParamsCursor_BackwardPagination(t *testing.T) {
	last := 5

	params, err := buildLeaderboardConfigFilterParamsCursor(nil, nil, nil, &last, nil)

	require.NoError(t, err)
	assert.True(t, params.Isbackward)
	assert.Equal(t, int32(6), params.Querylimit)
}

func TestBuildLeaderboardConfigFilterParamsCursor_EmptyIdsTreatedAsNoFilter(t *testing.T) {
	filter := &model.LeaderboardConfigFilter{
		Ids: []string{},
	}

	params, err := buildLeaderboardConfigFilterParamsCursor(filter, nil, nil, nil, nil)

	require.NoError(t, err)
	assert.Nil(t, params.Ids, "an explicitly empty ids slice should not filter out every row")
}

func TestBuildCountLeaderboardConfigsFilterParams_NilFilter(t *testing.T) {
	params := buildCountLeaderboardConfigsFilterParams(nil)
	assert.Equal(t, "", params.Projectid)
	assert.Nil(t, params.Isactive)
}

func TestBuildCountLeaderboardConfigsFilterParams_WithFilter(t *testing.T) {
	eventID := "EV01ARZ3NDEKTSV4RRFFQ69G5FAV"
	filter := &model.LeaderboardConfigFilter{
		EventID: &eventID,
		Ids:     []string{"LC1", "LC2"},
	}

	params := buildCountLeaderboardConfigsFilterParams(filter)

	assert.Equal(t, eventID, params.Eventid)
	assert.Equal(t, []string{"LC1", "LC2"}, params.Ids)
}

func TestBuildCountLeaderboardConfigsFilterParams_EmptyIdsTreatedAsNoFilter(t *testing.T) {
	filter := &model.LeaderboardConfigFilter{
		Ids: []string{},
	}

	params := buildCountLeaderboardConfigsFilterParams(filter)

	assert.Nil(t, params.Ids, "an explicitly empty ids slice should not filter out every row")
}

func TestLeaderboardMaxEntriesToDB(t *testing.T) {
	for _, value := range []*int{nil, intPtr(1), intPtr(20), intPtr(2147483647)} {
		result, err := leaderboardMaxEntriesToDB(value)
		require.NoError(t, err)
		if value == nil {
			assert.Nil(t, result)
		} else {
			require.NotNil(t, result)
			assert.Equal(t, *value, int(*result))
		}
	}
	for _, value := range []int{0, -1, 2147483648} {
		_, err := leaderboardMaxEntriesToDB(&value)
		assert.Error(t, err)
	}
}

func TestPaginateConfiguredLeaderboard(t *testing.T) {
	entries := make([]services.LeaderboardEntry, 5)
	for i := range entries {
		entries[i].Rank = int64(i + 1)
	}
	for _, tt := range []struct {
		name                    string
		first, last             *int
		after, before           *string
		count, start            int
		previous, next, invalid bool
	}{
		{name: "whole board", first: intPtr(100), count: 5, start: 1},
		{name: "first page", first: intPtr(2), count: 2, start: 1, next: true},
		{name: "middle page", first: intPtr(2), after: stringPtr("2"), count: 2, start: 3, previous: true, next: true},
		{name: "last page", first: intPtr(2), after: stringPtr("4"), count: 1, start: 5, previous: true},
		{name: "at cap", first: intPtr(100), after: stringPtr("5"), previous: true},
		{name: "beyond cap", first: intPtr(100), after: stringPtr("99"), previous: true},
		{name: "backward", last: intPtr(2), count: 2, start: 4, previous: true},
		{name: "backward before", last: intPtr(2), before: stringPtr("4"), count: 2, start: 2, previous: true, next: true},
		{name: "before beyond cap", last: intPtr(2), before: stringPtr("99"), count: 2, start: 4, previous: true},
		{name: "both bounds", first: intPtr(10), after: stringPtr("2"), before: stringPtr("5"), count: 2, start: 3, previous: true, next: true},
		{name: "zero page", first: intPtr(0), next: true},
		{name: "negative first", first: intPtr(-1), invalid: true},
		{name: "negative last", last: intPtr(-1), invalid: true},
		{name: "both page sizes", first: intPtr(1), last: intPtr(1), invalid: true},
		{name: "invalid after", first: intPtr(1), after: stringPtr("garbage"), invalid: true},
		{name: "invalid before", last: intPtr(1), before: stringPtr("0"), invalid: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			page, previous, next, err := paginateConfiguredLeaderboard(entries, tt.first, tt.after, tt.last, tt.before)
			if tt.invalid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, page, tt.count)
			assert.Equal(t, tt.previous, previous)
			assert.Equal(t, tt.next, next)
			if tt.count > 0 {
				assert.Equal(t, int64(tt.start), page[0].Rank)
			}
		})
	}
}

func TestConfiguredLeaderboardMaxEntries(t *testing.T) {
	for _, isEvent := range []bool{false, true} {
		for _, maxEntries := range []*int{nil, intPtr(3), intPtr(150)} {
			t.Run(fmt.Sprintf("event=%v/limit=%v", isEvent, maxEntries), func(t *testing.T) {
				queries := mocks.NewMockLeaderboardQuerier(t)
				c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
				require.NoError(t, err)
				t.Cleanup(c.Close)
				rows := make([]*sqlc.GetFullProjectPersonLeaderboardRow, 200)
				eventRows := make([]*sqlc.GetFullEventPersonLeaderboardRow, len(rows))
				for i := range rows {
					rows[i] = &sqlc.GetFullProjectPersonLeaderboardRow{EntityID: fmt.Sprintf("US%026d", i+1), Rank: int64(i + 1), Score: int64(200 - i), ChurchID: "CH1"}
					eventRows[i] = (*sqlc.GetFullEventPersonLeaderboardRow)(rows[i])
				}
				if isEvent {
					queries.On("GetFullEventPersonLeaderboard", mock.Anything, mock.Anything).Return(eventRows, nil)
				} else {
					queries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).Return(rows, nil)
				}
				config := &model.LeaderboardConfig{ProjectID: "PR1", EntityType: model.LeaderboardEntityTypePersons, MaxEntries: maxEntries}
				if isEvent {
					config.EventID = stringPtr("EV1")
				}
				userID := rows[199].EntityID
				ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)
				r := &Resolver{LeaderboardService: services.NewLeaderboardService(queries, c, nil)}
				board, err := r.getLeaderboardForConfig(ctx, config, nil, nil, nil, nil)
				require.NoError(t, err)
				count := 100
				if maxEntries != nil {
					count = *maxEntries
				}
				require.Len(t, board.Edges, count)
				assert.Equal(t, 200, board.TotalCount)
				assert.Equal(t, maxEntries == nil, board.PageInfo.HasNextPage)
				require.NotNil(t, board.Me)
				assert.Equal(t, userID, board.Me.ID)
				rivals, err := r.LeaderboardConnection().NearestChurchRivals(ctx, board, intPtr(3))
				require.NoError(t, err)
				require.Len(t, rivals, 3)
				assert.Equal(t, 199, *rivals[0].Rank)
			})
		}
	}
}
