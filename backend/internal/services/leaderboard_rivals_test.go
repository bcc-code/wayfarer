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

func TestGetProjectLeaderboardNearestChurchRivalsIntegration(t *testing.T) {
	mockQueries := mocks.NewMockLeaderboardQuerier(t)
	service := NewLeaderboardService(mockQueries, newTestCache(), nil)

	rows := make([]*sqlc.GetFullProjectPersonLeaderboardRow, 0, 25)
	for i := int64(1); i <= 25; i++ {
		rows = append(rows, personRowWithChurch(
			fmt.Sprintf("US01ARZ3NDEKTSV4RRFFQ69G5%03d", i), i, 1000-i, "CHMINE"))
	}
	rows = append(rows, personRowWithChurch("US01ARZ3NDEKTSV4RRFFQ69G5ME0", 26, 1000-26, "CHMINE"))
	mockQueries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.Anything).Return(rows, nil).Once()

	params := personLeaderboardParams("US01ARZ3NDEKTSV4RRFFQ69G5ME0")
	_, _, _, rivals, err := service.GetProjectLeaderboard(context.Background(), params)
	require.NoError(t, err)

	// Capped server-side at maxNearestChurchRivals even though 25 same-church
	// candidates rank above "me".
	assert.Len(t, rivals, maxNearestChurchRivals)
	assert.Equal(t, "US01ARZ3NDEKTSV4RRFFQ69G5025", rivals[0].EntityID, "nearest rival (rank 25) comes first")
}
