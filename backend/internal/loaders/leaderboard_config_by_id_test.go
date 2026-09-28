package loaders

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertRowToLeaderboardConfig(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	eventID := "EV01ARZ3NDEKTSV4RRFFQ69G5FAV"

	maxEntries := int32(20)
	row := &sqlc.LeaderboardConfig{
		ID:         "LC01ARZ3NDEKTSV4RRFFQ69G5FAV",
		ProjectID:  "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		EventID:    &eventID,
		Name:       "Top Churches",
		EntityType: "CHURCHES",
		Filter:     []byte(`{"minScore":5}`),
		SortOrder:  2,
		MaxEntries: &maxEntries,
		IsActive:   true,
		CreatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
	}

	result := ConvertRowToLeaderboardConfig(row)

	assert.Equal(t, row.ID, result.ID)
	assert.Equal(t, row.ProjectID, result.ProjectID)
	require.NotNil(t, result.EventID)
	assert.Equal(t, eventID, *result.EventID)
	assert.Equal(t, "Top Churches", result.Name)
	assert.Equal(t, model.LeaderboardEntityTypeChurches, result.EntityType)
	require.NotNil(t, result.Filter)
	require.NotNil(t, result.Filter.MinScore)
	assert.Equal(t, 5, *result.Filter.MinScore)
	require.NotNil(t, result.MaxEntries)
	assert.Equal(t, 20, *result.MaxEntries)
	assert.Equal(t, 2, result.SortOrder)
	assert.True(t, result.IsActive)
	assert.True(t, now.Equal(result.CreatedAt.Time))
}

func TestConvertRowToLeaderboardConfig_NilFilterAndEvent(t *testing.T) {
	row := &sqlc.LeaderboardConfig{
		ID:         "LC01ARZ3NDEKTSV4RRFFQ69G5FAV",
		ProjectID:  "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		EventID:    nil,
		Name:       "Global",
		EntityType: "PERSONS",
		Filter:     nil,
		SortOrder:  0,
		IsActive:   false,
	}

	result := ConvertRowToLeaderboardConfig(row)

	assert.Nil(t, result.EventID)
	assert.Nil(t, result.Filter)
	assert.Nil(t, result.MaxEntries)
	assert.False(t, result.IsActive)
}
