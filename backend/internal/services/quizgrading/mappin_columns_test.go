package quizgrading

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapPinColumnsRoundTrip(t *testing.T) {
	t.Run("POINT_RADIUS stores the target and radii, no polygon", func(t *testing.T) {
		c := pointRadius()

		cols, err := c.Columns()
		require.NoError(t, err)
		require.NotNil(t, cols.Mode)
		assert.Equal(t, MapPinModePointRadius, *cols.Mode)
		assert.Equal(t, ptr(int32(1000)), cols.Width)
		assert.Equal(t, ptr(int32(1000)), cols.Height)
		assert.Equal(t, ptr(500.0), cols.TargetX)
		assert.Equal(t, ptr(300.0), cols.TargetY)
		assert.Equal(t, ptr(40.0), cols.FullRadius)
		assert.Equal(t, ptr(200.0), cols.ZeroRadius)
		assert.Nil(t, cols.Polygon, "a NULL polygon is what the CHECK requires for POINT_RADIUS")

		back, err := cols.Config()
		require.NoError(t, err)
		assert.Equal(t, c, back)
	})

	t.Run("POLYGON stores the vertices as JSON, no target or radii", func(t *testing.T) {
		c := polygonConfig(MapPoint{X: 100.25, Y: 100}, MapPoint{X: 300, Y: 100.5}, MapPoint{X: 200, Y: 1.0 / 3})

		cols, err := c.Columns()
		require.NoError(t, err)
		assert.Equal(t, MapPinModePolygon, *cols.Mode)
		assert.Nil(t, cols.TargetX)
		assert.Nil(t, cols.TargetY)
		assert.Nil(t, cols.FullRadius)
		assert.Nil(t, cols.ZeroRadius)
		assert.JSONEq(t, `[{"x":100.25,"y":100},{"x":300,"y":100.5},{"x":200,"y":0.3333333333333333}]`, string(cols.Polygon))

		back, err := cols.Config()
		require.NoError(t, err)
		assert.Equal(t, c, back, "floats must survive the JSON round trip exactly")
		assert.True(t, c.Equal(back))
	})

	t.Run("columns don't share memory with the config", func(t *testing.T) {
		c := pointRadius()
		cols, err := c.Columns()
		require.NoError(t, err)

		*cols.TargetX = 1
		*cols.FullRadius = 2
		assert.Equal(t, 500.0, c.Target.X)
		assert.Equal(t, 40.0, *c.FullRadius)
	})
}

func TestMapPinColumnsConfig(t *testing.T) {
	t.Run("all NULL gives the zero config", func(t *testing.T) {
		c, err := MapPinColumns{}.Config()
		require.NoError(t, err)
		assert.Equal(t, MapPinConfig{}, c)
	})

	t.Run("a target needs both coordinates", func(t *testing.T) {
		c, err := MapPinColumns{TargetX: ptr(1.0)}.Config()
		require.NoError(t, err)
		assert.Nil(t, c.Target)
	})

	t.Run("a polygon that isn't a list of points is an error and can't score", func(t *testing.T) {
		for name, polygon := range map[string]string{
			"numbers":    `[1, 2, 3]`,
			"object":     `{"x": 1, "y": 2}`,
			"bad values": `[{"x": "a", "y": 1}, {"x": 2, "y": 2}, {"x": 3, "y": 1}]`,
		} {
			t.Run(name, func(t *testing.T) {
				c, err := MapPinColumns{
					Mode:    ptr(MapPinModePolygon),
					Width:   ptr(int32(1000)),
					Height:  ptr(int32(1000)),
					Polygon: []byte(polygon),
				}.Config()
				require.Error(t, err)
				assert.Equal(t, MapPinModePolygon, c.Mode, "the rest of the config is still returned")
				assert.Equal(t, 1000, c.Width)
				assert.Nil(t, c.Polygon, "a partly decoded polygon must not be kept")

				_, _, scoreErr := MapPinScore(c, 10, MapPoint{X: 1, Y: 1})
				assert.Error(t, scoreErr, "a broken config must not award points")
			})
		}
	})
}

func TestMapPinConfigEqual(t *testing.T) {
	assert.True(t, pointRadius().Equal(pointRadius()))
	assert.True(t, polygonConfig(triangle...).Equal(polygonConfig(triangle...)))

	tests := []struct {
		name   string
		a      MapPinConfig
		modify func(*MapPinConfig)
	}{
		{name: "mode", a: pointRadius(), modify: func(c *MapPinConfig) { c.Mode = MapPinModePolygon }},
		{name: "width", a: pointRadius(), modify: func(c *MapPinConfig) { c.Width = 999 }},
		{name: "height", a: pointRadius(), modify: func(c *MapPinConfig) { c.Height = 999 }},
		{name: "target", a: pointRadius(), modify: func(c *MapPinConfig) { c.Target = &MapPoint{X: 500, Y: 301} }},
		{name: "missing target", a: pointRadius(), modify: func(c *MapPinConfig) { c.Target = nil }},
		{name: "full radius", a: pointRadius(), modify: func(c *MapPinConfig) { c.FullRadius = ptr(41.0) }},
		{name: "zero radius", a: pointRadius(), modify: func(c *MapPinConfig) { c.ZeroRadius = ptr(201.0) }},
		{name: "missing zero radius", a: pointRadius(), modify: func(c *MapPinConfig) { c.ZeroRadius = nil }},
		{name: "vertex moved", a: polygonConfig(triangle...), modify: func(c *MapPinConfig) {
			c.Polygon = []MapPoint{{100, 100}, {300, 100}, {200, 301}}
		}},
		{name: "vertex added", a: polygonConfig(triangle...), modify: func(c *MapPinConfig) {
			c.Polygon = append([]MapPoint{}, rectangle...)
		}},
		{name: "same vertices, other starting point", a: polygonConfig(triangle...), modify: func(c *MapPinConfig) {
			c.Polygon = []MapPoint{triangle[1], triangle[2], triangle[0]}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.a
			tt.modify(&b)
			assert.False(t, tt.a.Equal(b))
			assert.False(t, b.Equal(tt.a))
		})
	}

	t.Run("a closing vertex is no change once normalized", func(t *testing.T) {
		stored, err := ValidateMapPinConfig(polygonConfig(triangle...))
		require.NoError(t, err)
		resent, err := ValidateMapPinConfig(polygonConfig(append(append([]MapPoint{}, triangle...), triangle[0])...))
		require.NoError(t, err)
		assert.True(t, stored.Equal(resent))
	})
}
