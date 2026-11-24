package loaders

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
)

func TestMapPinPosition(t *testing.T) {
	t.Run("reads a stored pin", func(t *testing.T) {
		assert.Equal(t, &model.MapPoint{X: 3, Y: 4}, MapPinPosition("QR1", []byte(`{"x": 3, "y": 4}`)))
	})

	t.Run("an unreadable pin is nil, not (0, 0)", func(t *testing.T) {
		assert.Nil(t, MapPinPosition("QR1", []byte(`{}`)))
		assert.Nil(t, MapPinPosition("QR1", nil))
		assert.Nil(t, MapPinPosition("QR1", []byte(`"oops"`)))
	})
}
