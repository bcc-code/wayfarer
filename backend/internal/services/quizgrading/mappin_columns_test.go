package quizgrading

import (
	"math"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pgPoint(x, y float64) pgtype.Point {
	return pgtype.Point{P: pgtype.Vec2{X: x, Y: y}, Valid: true}
}

func TestMapPinColumnsRoundTrip(t *testing.T) {
	t.Run("POINT_RADIUS stores the target and radii, no polygon", func(t *testing.T) {
		c := pointRadius()

		cols := c.Columns()
		require.NotNil(t, cols.Mode)
		assert.Equal(t, MapPinModePointRadius, *cols.Mode)
		assert.Equal(t, ptr(int32(1000)), cols.Width)
		assert.Equal(t, ptr(int32(1000)), cols.Height)
		assert.Equal(t, pgPoint(500, 300), cols.Target)
		assert.Equal(t, ptr(int32(40)), cols.FullRadius)
		assert.Equal(t, ptr(int32(200)), cols.ZeroRadius)
		assert.Nil(t, cols.Polygon, "a NULL polygon is what the CHECK requires for POINT_RADIUS")

		back, err := cols.Config()
		require.NoError(t, err)
		assert.Equal(t, c, back)
	})

	t.Run("POLYGON stores the vertices as points, no target or radii", func(t *testing.T) {
		c := polygonConfig(MapPoint{X: 100, Y: 100}, MapPoint{X: 300, Y: 101}, MapPoint{X: 200, Y: 0})

		cols := c.Columns()
		assert.Equal(t, MapPinModePolygon, *cols.Mode)
		assert.False(t, cols.Target.Valid, "a NULL target")
		assert.Nil(t, cols.FullRadius)
		assert.Nil(t, cols.ZeroRadius)
		assert.Equal(t, []pgtype.Point{pgPoint(100, 100), pgPoint(300, 101), pgPoint(200, 0)}, cols.Polygon)

		back, err := cols.Config()
		require.NoError(t, err)
		assert.Equal(t, c, back)
		assert.True(t, c.Equal(back))
	})

	t.Run("columns don't share memory with the config", func(t *testing.T) {
		c := pointRadius()
		cols := c.Columns()

		*cols.Width = 1
		*cols.FullRadius = 2
		cols.Target.P.X = 3
		assert.Equal(t, 1000, c.Width)
		assert.Equal(t, 40, *c.FullRadius)
		assert.Equal(t, 500, c.Target.X)
	})
}

func TestMapPinColumnsConfig(t *testing.T) {
	t.Run("all NULL gives the zero config", func(t *testing.T) {
		c, err := MapPinColumns{}.Config()
		require.NoError(t, err)
		assert.Equal(t, MapPinConfig{}, c)
	})

	t.Run("points are rounded to whole pixels", func(t *testing.T) {
		c, err := MapPinColumns{
			Target:  pgPoint(10.4, 19.6),
			Polygon: []pgtype.Point{pgPoint(0.49, 0), pgPoint(9.5, 0), pgPoint(5, 7.51)},
		}.Config()
		require.NoError(t, err)
		assert.Equal(t, &MapPoint{X: 10, Y: 20}, c.Target)
		assert.Equal(t, []MapPoint{{0, 0}, {10, 0}, {5, 8}}, c.Polygon)
	})

	brokenPolygons := map[string][]pgtype.Point{
		"a NULL vertex":   {pgPoint(1, 1), {}, pgPoint(3, 1)},
		"a NaN vertex":    {pgPoint(1, 1), pgPoint(math.NaN(), 2), pgPoint(3, 1)},
		"an infinite one": {pgPoint(1, 1), pgPoint(2, 2), pgPoint(3, math.Inf(1))},
	}
	for name, polygon := range brokenPolygons {
		t.Run("a polygon with "+name+" is an error and can't score", func(t *testing.T) {
			c, err := MapPinColumns{
				Mode:    ptr(MapPinModePolygon),
				Width:   ptr(int32(1000)),
				Height:  ptr(int32(1000)),
				Polygon: polygon,
			}.Config()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "map_polygon[")
			assert.Equal(t, MapPinModePolygon, c.Mode, "the rest of the config is still returned")
			assert.Equal(t, 1000, c.Width)
			assert.Nil(t, c.Polygon, "a partly converted polygon must not be kept")

			_, _, scoreErr := MapPinScore(c, 10, MapPoint{X: 1, Y: 1})
			assert.Error(t, scoreErr, "a broken config must not award points")
		})
	}

	t.Run("a non-finite target is an error and can't score", func(t *testing.T) {
		c, err := MapPinColumns{
			Mode:       ptr(MapPinModePointRadius),
			Width:      ptr(int32(1000)),
			Height:     ptr(int32(1000)),
			Target:     pgPoint(math.NaN(), 1),
			FullRadius: ptr(int32(40)),
			ZeroRadius: ptr(int32(200)),
		}.Config()
		require.ErrorContains(t, err, "map_target")
		assert.Nil(t, c.Target)
		_, _, scoreErr := MapPinScore(c, 10, MapPoint{X: 1, Y: 1})
		assert.Error(t, scoreErr)
	})
}

func TestPointFromPG(t *testing.T) {
	t.Run("NULL is nil", func(t *testing.T) {
		p, err := PointFromPG(pgtype.Point{})
		require.NoError(t, err)
		assert.Nil(t, p)
	})

	t.Run("whole pixels are exact", func(t *testing.T) {
		p, err := PointFromPG(pgPoint(812, 430))
		require.NoError(t, err)
		assert.Equal(t, &MapPoint{X: 812, Y: 430}, p)
	})

	t.Run("anything else rounds to the nearest pixel", func(t *testing.T) {
		for _, tt := range []struct {
			x, y float64
			want MapPoint
		}{
			{0.4, 0.6, MapPoint{0, 1}},
			{2.5, 3.5, MapPoint{3, 4}},
			{99.999999, 0.000001, MapPoint{100, 0}},
			{-0.4, -0.6, MapPoint{0, -1}},
		} {
			p, err := PointFromPG(pgPoint(tt.x, tt.y))
			require.NoError(t, err)
			assert.Equal(t, &tt.want, p, "(%g, %g)", tt.x, tt.y)
		}
	})

	for name, pt := range map[string]pgtype.Point{
		"NaN x":  pgPoint(math.NaN(), 1),
		"NaN y":  pgPoint(1, math.NaN()),
		"+Inf x": pgPoint(math.Inf(1), 1),
		"-Inf y": pgPoint(1, math.Inf(-1)),
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			p, err := PointFromPG(pt)
			assert.ErrorContains(t, err, "not finite")
			assert.Nil(t, p)
		})
	}
}

func TestPointToPG(t *testing.T) {
	assert.Equal(t, pgtype.Point{}, PointToPG(nil), "nil is NULL")

	for _, p := range []MapPoint{{0, 0}, {812, 430}, {MapPinMaxSize, MapPinMaxSize}} {
		pg := PointToPG(&p)
		assert.True(t, pg.Valid)
		back, err := PointFromPG(pg)
		require.NoError(t, err)
		assert.Equal(t, &p, back, "round trip")
	}
}

func TestInt32PtrAndIntPtr(t *testing.T) {
	assert.Nil(t, Int32Ptr(nil))
	assert.Nil(t, IntPtr(nil))
	assert.Equal(t, ptr(int32(40)), Int32Ptr(ptr(40)))
	assert.Equal(t, ptr(40), IntPtr(ptr(int32(40))))

	v := 7
	p := Int32Ptr(&v)
	*p = 8
	assert.Equal(t, 7, v, "a copy, not an alias")
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
		{name: "full radius", a: pointRadius(), modify: func(c *MapPinConfig) { c.FullRadius = ptr(41) }},
		{name: "zero radius", a: pointRadius(), modify: func(c *MapPinConfig) { c.ZeroRadius = ptr(201) }},
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

func TestNewMapPinColumns(t *testing.T) {
	t.Run("puts each argument in its own field", func(t *testing.T) {
		mode := MapPinModePointRadius
		polygon := []pgtype.Point{pgPoint(1, 2)}

		cols := NewMapPinColumns(&mode, ptr(int32(10)), ptr(int32(20)),
			pgPoint(1, 2), ptr(int32(3)), ptr(int32(4)), polygon)

		assert.Equal(t, MapPinColumns{
			Mode:       &mode,
			Width:      ptr(int32(10)),
			Height:     ptr(int32(20)),
			Target:     pgPoint(1, 2),
			FullRadius: ptr(int32(3)),
			ZeroRadius: ptr(int32(4)),
			Polygon:    polygon,
		}, cols)
	})

	t.Run("all NULL for a question that isn't MAP_PIN", func(t *testing.T) {
		assert.Equal(t, MapPinColumns{}, NewMapPinColumns(nil, nil, nil, pgtype.Point{}, nil, nil, nil))
	})
}
