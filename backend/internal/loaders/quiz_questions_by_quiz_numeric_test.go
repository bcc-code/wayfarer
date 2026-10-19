package loaders

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNumericToFloat(t *testing.T) {
	t.Run("not set", func(t *testing.T) {
		assert.Nil(t, numericToFloat(pgtype.Numeric{}))
	})

	t.Run("value", func(t *testing.T) {
		var n pgtype.Numeric
		require.NoError(t, n.Scan("2.50"))
		got := numericToFloat(n)
		require.NotNil(t, got)
		assert.Equal(t, 2.5, *got)
	})

	t.Run("NaN is not exposed", func(t *testing.T) {
		assert.Nil(t, numericToFloat(pgtype.Numeric{Valid: true, NaN: true}))
	})

	t.Run("infinity is not exposed", func(t *testing.T) {
		assert.Nil(t, numericToFloat(pgtype.Numeric{Valid: true, InfinityModifier: pgtype.Infinity}))
	})
}
