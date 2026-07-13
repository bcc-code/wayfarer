package services

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
)

func personRow(userID string, rank int64, score int64) *sqlc.GetFullProjectPersonLeaderboardRow {
	return &sqlc.GetFullProjectPersonLeaderboardRow{
		EntityID:   userID,
		Name:       fmt.Sprintf("User %d", rank),
		ChurchName: "Church",
		Score:      score,
		Rank:       rank,
	}
}

func personLeaderboardParams(userID string) LeaderboardParams {
	first := 10
	return LeaderboardParams{
		ContextID:  "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		EntityType: model.LeaderboardEntityTypePersons,
		First:      &first,
		UserID:     userID,
	}
}

func TestGetProjectLeaderboardCachesDecodedBoard(t *testing.T) {
	mockQueries := mocks.NewMockLeaderboardQuerier(t)
	service := NewLeaderboardService(mockQueries, newTestCache(), nil)

	rows := []*sqlc.GetFullProjectPersonLeaderboardRow{
		personRow("US01ARZ3NDEKTSV4RRFFQ69G5FA1", 1, 300),
		personRow("US01ARZ3NDEKTSV4RRFFQ69G5FA2", 2, 200),
		personRow("US01ARZ3NDEKTSV4RRFFQ69G5FA3", 3, 100),
	}
	// .Once() asserts the second request is served from cache without a query
	mockQueries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).Return(rows, nil).Once()

	params := personLeaderboardParams("US01ARZ3NDEKTSV4RRFFQ69G5FA2")

	entries, me, total, err := service.GetProjectLeaderboard(context.Background(), params)
	require.NoError(t, err)
	assert.Len(t, entries, 3)
	assert.Equal(t, 3, total)
	require.NotNil(t, me)
	assert.Equal(t, int64(2), me.Rank)
	assert.Equal(t, 200, me.Score)

	// Wait for ristretto to admit the entry, then hit the cache
	service.cache.Wait()

	entries, me, total, err = service.GetProjectLeaderboard(context.Background(), params)
	require.NoError(t, err)
	assert.Len(t, entries, 3)
	assert.Equal(t, 3, total)
	require.NotNil(t, me)
	assert.Equal(t, int64(2), me.Rank)
}

func TestGetProjectLeaderboardMeNotOnBoard(t *testing.T) {
	mockQueries := mocks.NewMockLeaderboardQuerier(t)
	service := NewLeaderboardService(mockQueries, newTestCache(), nil)

	rows := []*sqlc.GetFullProjectPersonLeaderboardRow{
		personRow("US01ARZ3NDEKTSV4RRFFQ69G5FA1", 1, 300),
	}
	mockQueries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).Return(rows, nil).Once()

	_, me, total, err := service.GetProjectLeaderboard(context.Background(), personLeaderboardParams("US01ARZ3NDEKTSV4RRFFQ69G5FA9"))
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Nil(t, me, "user not on the board should get no me entry")
}

func TestGetProjectLeaderboardPaginatesSharedBoard(t *testing.T) {
	mockQueries := mocks.NewMockLeaderboardQuerier(t)
	service := NewLeaderboardService(mockQueries, newTestCache(), nil)

	rows := make([]*sqlc.GetFullProjectPersonLeaderboardRow, 0, 5)
	for i := int64(1); i <= 5; i++ {
		rows = append(rows, personRow(fmt.Sprintf("US01ARZ3NDEKTSV4RRFFQ69G5FA%d", i), i, 600-i*100))
	}
	mockQueries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).Return(rows, nil).Once()

	first := 2
	params := personLeaderboardParams("")
	params.First = &first

	entries, _, total, err := service.GetProjectLeaderboard(context.Background(), params)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	require.Len(t, entries, 2)
	assert.Equal(t, int64(1), entries[0].Rank)
	assert.Equal(t, int64(2), entries[1].Rank)

	service.cache.Wait()

	// Second page from the cached board
	after := "2"
	params.After = &after
	entries, _, _, err = service.GetProjectLeaderboard(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, int64(3), entries[0].Rank)
	assert.Equal(t, int64(4), entries[1].Rank)
}

func TestGetProjectLeaderboardSingleflightSharesFetch(t *testing.T) {
	mockQueries := mocks.NewMockLeaderboardQuerier(t)
	service := NewLeaderboardService(mockQueries, newTestCache(), nil)

	rows := []*sqlc.GetFullProjectPersonLeaderboardRow{
		personRow("US01ARZ3NDEKTSV4RRFFQ69G5FA1", 1, 300),
	}
	// A slow query while many requests hit the same cold key: exactly one fetch
	mockQueries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { time.Sleep(50 * time.Millisecond) }).
		Return(rows, nil).Once()

	params := personLeaderboardParams("")

	const concurrency = 10
	var wg sync.WaitGroup
	errs := make([]error, concurrency)
	totals := make([]int, concurrency)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _, total, err := service.GetProjectLeaderboard(context.Background(), params)
			errs[i] = err
			totals[i] = total
		}(i)
	}
	wg.Wait()

	for i := 0; i < concurrency; i++ {
		require.NoError(t, errs[i])
		assert.Equal(t, 1, totals[i])
	}
}

func TestGetProjectLeaderboardFetchErrorPropagates(t *testing.T) {
	mockQueries := mocks.NewMockLeaderboardQuerier(t)
	service := NewLeaderboardService(mockQueries, newTestCache(), nil)

	mockQueries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("db down")).Once()

	_, _, _, err := service.GetProjectLeaderboard(context.Background(), personLeaderboardParams(""))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get full project person leaderboard")
}

func TestGetProjectLeaderboardCancelledCallerDoesNotPoisonFlight(t *testing.T) {
	mockQueries := mocks.NewMockLeaderboardQuerier(t)
	service := NewLeaderboardService(mockQueries, newTestCache(), nil)

	rows := []*sqlc.GetFullProjectPersonLeaderboardRow{
		personRow("US01ARZ3NDEKTSV4RRFFQ69G5FA1", 1, 300),
	}
	// The fetch runs with context.WithoutCancel, so the query context must not
	// be cancelled even though the caller's context already is. It must still
	// carry the service-owned fetch timeout.
	mockQueries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			queryCtx := args.Get(0).(context.Context)
			assert.NoError(t, queryCtx.Err(), "query context should not inherit caller cancellation")
			deadline, ok := queryCtx.Deadline()
			assert.True(t, ok, "query context should have the service-owned fetch timeout")
			assert.LessOrEqual(t, time.Until(deadline), leaderboardFetchTimeout)
		}).
		Return(rows, nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, total, err := service.GetProjectLeaderboard(ctx, personLeaderboardParams(""))
	require.NoError(t, err)
	assert.Equal(t, 1, total)
}
