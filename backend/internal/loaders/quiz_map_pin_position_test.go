package loaders

import (
	"math"
	"testing"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestMapPinPosition(t *testing.T) {
	t.Run("reads a stored pin in whole pixels", func(t *testing.T) {
		assert.Equal(t, &model.MapPoint{X: 3, Y: 4}, MapPinPosition("QR1", pgPoint(3, 4)))
	})

	t.Run("rounds to the nearest pixel", func(t *testing.T) {
		assert.Equal(t, &model.MapPoint{X: 3, Y: 5}, MapPinPosition("QR1", pgPoint(2.6, 4.5)))
	})

	t.Run("a missing or broken pin is nil, not (0, 0)", func(t *testing.T) {
		assert.Nil(t, MapPinPosition("QR1", pgtype.Point{}), "NULL")
		assert.Nil(t, MapPinPosition("QR1", pgPoint(math.NaN(), 1)), "NaN")
		assert.Nil(t, MapPinPosition("QR1", pgPoint(1, math.Inf(1))), "infinite")
	})
}
