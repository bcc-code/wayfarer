package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
)

// board builds a cachedLeaderboard from (entityID, rank, churchID) triples,
// in the given order (index 0 = first in Entries).
func board(rows ...[3]string) *cachedLeaderboard {
	entries := make([]LeaderboardEntry, len(rows))
	index := make(map[string]int, len(rows))
	for i, r := range rows {
		var rank int64
		_, err := fmt.Sscanf(r[1], "%d", &rank)
		if err != nil {
			panic(err)
		}
		entries[i] = LeaderboardEntry{EntityID: r[0], Rank: rank, ChurchID: r[2]}
		index[r[0]] = i
	}
	return &cachedLeaderboard{Entries: entries, IndexByEntityID: index}
}

func TestFindNearestChurchRivals(t *testing.T) {
	t.Run("returns nearest N same-church entries above me, nearest first", func(t *testing.T) {
		b := board(
			[3]string{"top", "1", "CHOTHER"},
			[3]string{"a", "2", "CHMINE"},
			[3]string{"b", "3", "CHOTHER"},
			[3]string{"c", "4", "CHMINE"},
			[3]string{"me", "5", "CHMINE"},
		)
		got := b.findNearestChurchRivals("me", 3)
		require.Len(t, got, 2)
		assert.Equal(t, "c", got[0].EntityID)
		assert.Equal(t, "a", got[1].EntityID)
	})

	t.Run("returns fewer than N when near the top of own church's ranks", func(t *testing.T) {
		b := board(
			[3]string{"a", "1", "CHMINE"},
			[3]string{"top", "2", "CHOTHER"},
			[3]string{"me", "3", "CHMINE"},
		)
		got := b.findNearestChurchRivals("me", 5)
		require.Len(t, got, 1)
		assert.Equal(t, "a", got[0].EntityID)
	})

	t.Run("returns nil when no same-church entries are above me", func(t *testing.T) {
		b := board(
			[3]string{"top", "1", "CHOTHER"},
			[3]string{"me", "2", "CHMINE"},
		)
		assert.Empty(t, b.findNearestChurchRivals("me", 3))
	})

	t.Run("N=0 returns nil", func(t *testing.T) {
		b := board(
			[3]string{"a", "1", "CHMINE"},
			[3]string{"me", "2", "CHMINE"},
		)
		assert.Empty(t, b.findNearestChurchRivals("me", 0))
	})

	t.Run("negative N returns nil", func(t *testing.T) {
		b := board(
			[3]string{"a", "1", "CHMINE"},
			[3]string{"me", "2", "CHMINE"},
		)
		assert.Empty(t, b.findNearestChurchRivals("me", -1))
	})

	t.Run("me not on board returns nil", func(t *testing.T) {
		b := board([3]string{"a", "1", "CHMINE"})
		assert.Empty(t, b.findNearestChurchRivals("missing", 3))
	})

	t.Run("me already at the top (index 0) returns nil", func(t *testing.T) {
		b := board(
			[3]string{"me", "1", "CHMINE"},
			[3]string{"a", "2", "CHMINE"},
		)
		assert.Empty(t, b.findNearestChurchRivals("me", 3))
	})

	t.Run("empty churchID for me returns nil", func(t *testing.T) {
		b := board(
			[3]string{"a", "1", ""},
			[3]string{"me", "2", ""},
		)
		assert.Empty(t, b.findNearestChurchRivals("me", 3))
	})

	t.Run("same-rank ties are not counted as above me", func(t *testing.T) {
		// DENSE_RANK ties: "tied" shares rank 2 with "me" but sorts before it
		// on the last_score_at/name tiebreak. It must not count as a rival.
		b := board(
			[3]string{"a", "1", "CHMINE"},
			[3]string{"tied", "2", "CHMINE"},
			[3]string{"me", "2", "CHMINE"},
		)
		got := b.findNearestChurchRivals("me", 5)
		require.Len(t, got, 1)
		assert.Equal(t, "a", got[0].EntityID)
	})
}

func TestNearestChurchRivalsUsesCachedBoard(t *testing.T) {
	mockQueries := mocks.NewMockLeaderboardQuerier(t)
	service := NewLeaderboardService(mockQueries, newTestCache(), nil)

	rows := make([]*sqlc.GetFullProjectPersonLeaderboardRow, 0, 25)
	for i := int64(1); i <= 25; i++ {
		rows = append(rows, personRowWithChurch(
			fmt.Sprintf("US01ARZ3NDEKTSV4RRFFQ69G5%03d", i), i, 1000-i, "CHMINE"))
	}
	rows = append(rows, personRowWithChurch("US01ARZ3NDEKTSV4RRFFQ69G5ME0", 26, 1000-26, "CHMINE"))
	// .Once() proves the second call below is served from cache, not a new DB query.
	mockQueries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).Return(rows, nil).Once()

	ctx := context.Background()
	params := LeaderboardParams{
		ContextID: "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		UserID:    "US01ARZ3NDEKTSV4RRFFQ69G5ME0",
	}

	// Requesting more than maxNearestChurchRivals is capped server-side even
	// though 25 same-church candidates rank above "me".
	rivals, err := service.NearestChurchRivals(ctx, params, false, 20)
	require.NoError(t, err)
	assert.Len(t, rivals, maxNearestChurchRivals)
	assert.Equal(t, "US01ARZ3NDEKTSV4RRFFQ69G5025", rivals[0].EntityID, "nearest rival (rank 25) comes first")

	service.cache.Wait()

	// Same contextID+filter must hit the warm cache, not re-query the DB.
	rivals, err = service.NearestChurchRivals(ctx, params, false, 2)
	require.NoError(t, err)
	assert.Len(t, rivals, 2)
}

func TestNearestChurchRivalsGuardsShortCircuitBeforeAnyFetch(t *testing.T) {
	// No .On(...) expectations set: mockery fails the test if the DB query
	// is called, proving these guards short-circuit before any board fetch.
	mockQueries := mocks.NewMockLeaderboardQuerier(t)
	service := NewLeaderboardService(mockQueries, newTestCache(), nil)
	ctx := context.Background()
	contextID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"

	rivals, err := service.NearestChurchRivals(ctx, LeaderboardParams{ContextID: contextID, UserID: ""}, false, 5)
	require.NoError(t, err)
	assert.Empty(t, rivals, "empty userID must short-circuit")

	withUser := LeaderboardParams{ContextID: contextID, UserID: "US01ARZ3NDEKTSV4RRFFQ69G5ME0"}
	rivals, err = service.NearestChurchRivals(ctx, withUser, false, 0)
	require.NoError(t, err)
	assert.Empty(t, rivals, "first<=0 must short-circuit")

	rivals, err = service.NearestChurchRivals(ctx, withUser, false, -1)
	require.NoError(t, err)
	assert.Empty(t, rivals, "negative first must short-circuit")
}
