package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

func TestMapUserPointsByProject(t *testing.T) {
	rows := []*sqlc.GetUserPointsByProjectRow{
		{ProjectID: "PR2", ProjectName: "Sommercamp 2026", Points: 0},
		{ProjectID: "PR1", ProjectName: "Den siste lunsjen", Points: 25531},
	}

	got := mapUserPointsByProject(rows)

	require.Len(t, got, 2)
	// Order is the query's (most recently active first) and must be preserved:
	// re-sorting here would silently override the ORDER BY.
	assert.Equal(t, "PR2", got[0].ProjectID)
	assert.Equal(t, "Sommercamp 2026", got[0].ProjectName)
	assert.Equal(t, 0, got[0].Points)
	assert.Equal(t, 25531, got[1].Points)
}

// A project the user scored in but netted to zero still belongs in the list.
// "0 poeng" against a project someone participated in is a real answer on a
// support call; omitting it would suggest they were never in that project.
func TestMapUserPointsByProjectKeepsZeroTotals(t *testing.T) {
	got := mapUserPointsByProject([]*sqlc.GetUserPointsByProjectRow{
		{ProjectID: "PR1", ProjectName: "Nullsum", Points: 0},
	})

	require.Len(t, got, 1)
	assert.Equal(t, 0, got[0].Points)
}

func TestMapUserPointsByProjectHandlesNegativeTotals(t *testing.T) {
	// Manual adjustments can be negative, so a net total can be too.
	got := mapUserPointsByProject([]*sqlc.GetUserPointsByProjectRow{
		{ProjectID: "PR1", ProjectName: "Minus", Points: -184},
	})

	require.Len(t, got, 1)
	assert.Equal(t, -184, got[0].Points)
}

func TestMapUserPointsByProjectEmpty(t *testing.T) {
	// A user with no scoring history must yield an empty slice, not nil: the
	// field is `[UserProjectPoints!]!`, and gqlgen marshals nil as `null`.
	got := mapUserPointsByProject(nil)

	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestMapUserPointsByProjectSkipsNilRows(t *testing.T) {
	got := mapUserPointsByProject([]*sqlc.GetUserPointsByProjectRow{
		nil,
		{ProjectID: "PR1", ProjectName: "Ok", Points: 5},
	})

	require.Len(t, got, 1)
	assert.Equal(t, "PR1", got[0].ProjectID)
}
