package api

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestBuildLeaderboardConnectionRivalsLookup checks that buildLeaderboardConnection
// stashes the nearestChurchRivals lookup context only for PERSONS leaderboards -
// it's meaningless for team/superteam/church boards, so those must leave it
// zero-valued rather than triggering a pointless board fetch later.
func TestBuildLeaderboardConnectionRivalsLookup(t *testing.T) {
	filter := &model.LeaderboardFilter{}

	t.Run("PERSONS populates the lookup context", func(t *testing.T) {
		connection, err := buildLeaderboardConnection(
			context.Background(),
			nil,
			&services.LeaderboardEntry{EntityID: "US01ARZ3NDEKTSV4RRFFQ69G5FA0", Rank: 340},
			1,
			"US01ARZ3NDEKTSV4RRFFQ69G5FA0",
			model.LeaderboardEntityTypePersons,
			"PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
			nil, // Loaders unused for PERSONS tag computation
			nil, nil, nil, nil,
			"PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
			false,
			filter,
		)
		require.NoError(t, err)
		assert.Equal(t, "PR01ARZ3NDEKTSV4RRFFQ69G5FAV", connection.RivalsContextID)
		assert.False(t, connection.RivalsIsEvent)
		assert.Same(t, filter, connection.RivalsFilter)
		assert.Equal(t, "US01ARZ3NDEKTSV4RRFFQ69G5FA0", connection.RivalsUserID)
	})
}

// TestNearestChurchRivalsWiring exercises the full path from
// buildLeaderboardConnection's lookup-context fields through the resolver
// and into a real LeaderboardService (backed by a mocked querier) - the link
// the pure-function and clamp-only tests don't cover on their own.
func TestNearestChurchRivalsWiring(t *testing.T) {
	mockQueries := mocks.NewMockLeaderboardQuerier(t)
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	svc := services.NewLeaderboardService(mockQueries, c, nil)

	rows := []*sqlc.GetFullProjectPersonLeaderboardRow{
		{EntityID: "US01ARZ3NDEKTSV4RRFFQ69G5FA1", Name: "Per", ChurchID: "CH1", ChurchName: "Church", Score: 331, Rank: 1},
		{EntityID: "US01ARZ3NDEKTSV4RRFFQ69G5FA2", Name: "Kari", ChurchID: "CH1", ChurchName: "Church", Score: 338, Rank: 2},
		{EntityID: "US01ARZ3NDEKTSV4RRFFQ69G5FA0", Name: "Me", ChurchID: "CH1", ChurchName: "Church", Score: 300, Rank: 3},
	}
	mockQueries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).Return(rows, nil).Once()

	connection, err := buildLeaderboardConnection(
		context.Background(),
		nil,
		&services.LeaderboardEntry{EntityID: "US01ARZ3NDEKTSV4RRFFQ69G5FA0", Rank: 3},
		3,
		"US01ARZ3NDEKTSV4RRFFQ69G5FA0",
		model.LeaderboardEntityTypePersons,
		"PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		nil,
		nil, nil, nil, nil,
		"PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		false,
		nil,
	)
	require.NoError(t, err)

	resolver := &leaderboardConnectionResolver{&Resolver{LeaderboardService: svc}}

	first := 1
	got, err := resolver.NearestChurchRivals(context.Background(), connection, &first)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Kari", got[0].Name, "nearest rival (rank 2, closer to me than rank 1) comes first")
}

// TestNearestChurchRivalsWiringSkipsNonPersons confirms a TEAMS-flavored
// connection (RivalsUserID left empty by buildLeaderboardConnection) never
// reaches the querier at all.
func TestNearestChurchRivalsWiringSkipsNonPersons(t *testing.T) {
	mockQueries := mocks.NewMockLeaderboardQuerier(t) // no .On(...): must not be called
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	svc := services.NewLeaderboardService(mockQueries, c, nil)

	connection := &model.LeaderboardConnection{}
	resolver := &leaderboardConnectionResolver{&Resolver{LeaderboardService: svc}}

	got, err := resolver.NearestChurchRivals(context.Background(), connection, nil)
	require.NoError(t, err)
	assert.Empty(t, got)
}
