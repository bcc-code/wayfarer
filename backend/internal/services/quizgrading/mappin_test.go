package quizgrading

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

// pointRadius is a 1000x1000 map with full points within 40 px of (500, 300)
// and none from 200 px on.
func pointRadius() MapPinConfig {
	return MapPinConfig{
		Mode:       MapPinModePointRadius,
		Width:      1000,
		Height:     1000,
		Target:     &MapPoint{X: 500, Y: 300},
		FullRadius: ptr(40.0),
		ZeroRadius: ptr(200.0),
	}
}

func polygonConfig(vertices ...MapPoint) MapPinConfig {
	return MapPinConfig{Mode: MapPinModePolygon, Width: 1000, Height: 1000, Polygon: vertices}
}

var (
	triangle  = []MapPoint{{100, 100}, {300, 100}, {200, 300}}
	rectangle = []MapPoint{{100, 100}, {400, 100}, {400, 300}, {100, 300}}
	// diamond has vertices left and right of its centre on y = 200, so a ray
	// from a point on that line passes exactly through a vertex.
	diamond = []MapPoint{{200, 100}, {300, 200}, {200, 300}, {100, 200}}
	// peak is a triangle whose apex (200, 100) is the only point a ray along
	// y = 100 meets.
	peak = []MapPoint{{200, 100}, {300, 300}, {100, 300}}
	// cShape is a 300x300 square with a notch cut from its right side:
	// x in (200, 400), y in (200, 300) is outside.
	cShape = []MapPoint{
		{100, 100}, {400, 100}, {400, 200}, {200, 200},
		{200, 300}, {400, 300}, {400, 400}, {100, 400},
	}
)

func TestMapPinScorePointRadius(t *testing.T) {
	tests := []struct {
		name      string
		pin       MapPoint
		full      float64
		maxPoints int
		want      int
		distance  float64
	}{
		{name: "exact hit", pin: MapPoint{500, 300}, full: 40, maxPoints: 100, want: 100, distance: 0},
		{name: "inside full radius", pin: MapPoint{506, 308}, full: 40, maxPoints: 100, want: 100, distance: 10},
		{name: "exactly on full radius", pin: MapPoint{540, 300}, full: 40, maxPoints: 100, want: 100, distance: 40},
		{name: "just beyond full radius rounds up to max", pin: MapPoint{540.5, 300}, full: 40, maxPoints: 100, want: 100, distance: 40.5},
		{name: "past the rounding band loses a point", pin: MapPoint{541, 300}, full: 40, maxPoints: 100, want: 99, distance: 41},
		{name: "midpoint of falloff", pin: MapPoint{500, 420}, full: 40, maxPoints: 100, want: 50, distance: 120},
		{name: "just inside zero radius still rounds up to 1", pin: MapPoint{699, 300}, full: 40, maxPoints: 100, want: 1, distance: 199},
		{name: "just inside zero radius rounds down to 0", pin: MapPoint{699.6, 300}, full: 40, maxPoints: 100, want: 0, distance: 199.6},
		{name: "exactly on zero radius", pin: MapPoint{500, 500}, full: 40, maxPoints: 100, want: 0, distance: 200},
		{name: "beyond zero radius", pin: MapPoint{900, 900}, full: 40, maxPoints: 100, want: 0, distance: math.Hypot(400, 600)},
		{name: "full 0: exact hit is max", pin: MapPoint{500, 300}, full: 0, maxPoints: 100, want: 100, distance: 0},
		{name: "full 0: plain linear", pin: MapPoint{550, 300}, full: 0, maxPoints: 100, want: 75, distance: 50},
		{name: "0.5 rounds up", pin: MapPoint{500, 480}, full: 0, maxPoints: 5, want: 1, distance: 180},
		{name: "1.5 rounds up", pin: MapPoint{500, 440}, full: 0, maxPoints: 5, want: 2, distance: 140},
		{name: "max 0 awards nothing", pin: MapPoint{500, 300}, full: 40, maxPoints: 0, want: 0, distance: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pointRadius()
			c.FullRadius = ptr(tt.full)

			got, distance, err := MapPinScore(c, tt.maxPoints, tt.pin)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			if assert.NotNil(t, distance) {
				assert.InDelta(t, tt.distance, *distance, 1e-9)
			}
		})
	}
}

func TestMapPinScorePolygon(t *testing.T) {
	c := polygonConfig(rectangle...)

	tests := []struct {
		name string
		pin  MapPoint
		want int
	}{
		{name: "inside", pin: MapPoint{250, 200}, want: 100},
		{name: "on an edge", pin: MapPoint{250, 100}, want: 100},
		{name: "on a vertex", pin: MapPoint{400, 300}, want: 100},
		{name: "one pixel outside", pin: MapPoint{401, 200}, want: 0},
		{name: "far outside", pin: MapPoint{900, 900}, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, distance, err := MapPinScore(c, 100, tt.pin)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Nil(t, distance, "POLYGON has no distance")
		})
	}
}

func TestMapPinScoreRejectsIncompleteConfig(t *testing.T) {
	missingRadius := pointRadius()
	missingRadius.ZeroRadius = nil

	tests := []struct {
		name   string
		config MapPinConfig
	}{
		{name: "POINT_RADIUS without zero radius", config: missingRadius},
		{name: "POLYGON without vertices", config: polygonConfig()},
		{name: "unknown mode", config: MapPinConfig{Mode: "CIRCLE", Width: 10, Height: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := MapPinScore(tt.config, 100, MapPoint{5, 5})
			assert.Error(t, err)
		})
	}
}

func TestMapPinScoreRejectsNegativeMaxPoints(t *testing.T) {
	tests := []struct {
		name   string
		config MapPinConfig
		pin    MapPoint
	}{
		{name: "POINT_RADIUS: exact hit", config: pointRadius(), pin: MapPoint{500, 300}},
		{name: "POINT_RADIUS: in the falloff", config: pointRadius(), pin: MapPoint{500, 420}},
		{name: "POINT_RADIUS: beyond zero radius", config: pointRadius(), pin: MapPoint{900, 900}},
		{name: "POLYGON: inside", config: polygonConfig(rectangle...), pin: MapPoint{250, 200}},
		{name: "POLYGON: outside", config: polygonConfig(rectangle...), pin: MapPoint{900, 900}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points, distance, err := MapPinScore(tt.config, -10, tt.pin)
			assert.ErrorContains(t, err, "maxPoints")
			assert.Zero(t, points)
			assert.Nil(t, distance)
		})
	}
}

func TestMapPinScoreRejectsInvalidPosition(t *testing.T) {
	configs := map[string]MapPinConfig{
		"POINT_RADIUS": pointRadius(),
		"POLYGON":      polygonConfig(rectangle...),
	}
	pins := []struct {
		name    string
		pin     MapPoint
		message string
	}{
		{name: "NaN x", pin: MapPoint{math.NaN(), 300}, message: "finite"},
		{name: "NaN y", pin: MapPoint{500, math.NaN()}, message: "finite"},
		{name: "+Inf", pin: MapPoint{math.Inf(1), 300}, message: "finite"},
		{name: "-Inf", pin: MapPoint{500, math.Inf(-1)}, message: "finite"},
		{name: "left of the map", pin: MapPoint{-1, 300}, message: "outside"},
		{name: "below the map", pin: MapPoint{500, 1000.5}, message: "outside"},
	}

	for mode, c := range configs {
		for _, tt := range pins {
			t.Run(mode+": "+tt.name, func(t *testing.T) {
				points, distance, err := MapPinScore(c, 100, tt.pin)
				assertMapPinError(t, err, "mapPosition", tt.message)
				assert.Zero(t, points)
				assert.Nil(t, distance)
			})
		}
	}

	t.Run("a pin on the map's edge is scored", func(t *testing.T) {
		points, distance, err := MapPinScore(pointRadius(), 100, MapPoint{500, 0})
		require.NoError(t, err)
		assert.Equal(t, 0, points, "300 px from the target is beyond the zero radius")
		if assert.NotNil(t, distance) {
			assert.InDelta(t, 300, *distance, 1e-9)
		}
	})
}

func TestPointInPolygon(t *testing.T) {
	tests := []struct {
		name    string
		polygon []MapPoint
		pin     MapPoint
		want    bool
	}{
		{name: "triangle: inside", polygon: triangle, pin: MapPoint{200, 150}, want: true},
		{name: "triangle: outside", polygon: triangle, pin: MapPoint{100, 300}, want: false},
		{name: "rectangle: inside", polygon: rectangle, pin: MapPoint{250, 200}, want: true},
		{name: "rectangle: outside", polygon: rectangle, pin: MapPoint{50, 200}, want: false},

		{name: "concave: inside an arm", polygon: cShape, pin: MapPoint{300, 150}, want: true},
		{name: "concave: inside the spine", polygon: cShape, pin: MapPoint{150, 250}, want: true},
		{name: "concave: inside the notch is outside", polygon: cShape, pin: MapPoint{300, 250}, want: false},
		{name: "concave: on the notch edge", polygon: cShape, pin: MapPoint{300, 200}, want: true},
		{name: "concave: on the notch's inner corner", polygon: cShape, pin: MapPoint{200, 300}, want: true},

		{name: "on a horizontal edge", polygon: rectangle, pin: MapPoint{250, 100}, want: true},
		{name: "on a vertical edge", polygon: rectangle, pin: MapPoint{400, 200}, want: true},
		{name: "on a diagonal edge", polygon: triangle, pin: MapPoint{250, 200}, want: true},
		{name: "on a vertex", polygon: rectangle, pin: MapPoint{100, 100}, want: true},
		{name: "within epsilon of an edge", polygon: rectangle, pin: MapPoint{250, 100 - MapPinEpsilon/2}, want: true},
		{name: "beyond epsilon of an edge", polygon: rectangle, pin: MapPoint{250, 100 - 10*MapPinEpsilon}, want: false},

		{name: "ray through a vertex, inside", polygon: diamond, pin: MapPoint{150, 200}, want: true},
		{name: "ray through two vertices, outside", polygon: diamond, pin: MapPoint{50, 200}, want: false},
		{name: "ray grazing an apex, outside", polygon: peak, pin: MapPoint{50, 100}, want: false},
		{name: "collinear with a horizontal edge, left of it", polygon: rectangle, pin: MapPoint{50, 100}, want: false},
		{name: "collinear with a horizontal edge, right of it", polygon: rectangle, pin: MapPoint{500, 100}, want: false},

		{name: "winding does not matter", polygon: []MapPoint{{100, 300}, {400, 300}, {400, 100}, {100, 100}}, pin: MapPoint{250, 200}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, PointInPolygon(tt.pin, tt.polygon))
		})
	}
}

func TestValidateMapPinConfigPolygon(t *testing.T) {
	t.Run("accepts a triangle, a rectangle and a concave polygon", func(t *testing.T) {
		for _, polygon := range [][]MapPoint{triangle, rectangle, cShape} {
			got, err := ValidateMapPinConfig(polygonConfig(polygon...))
			require.NoError(t, err)
			assert.Equal(t, polygon, got.Polygon)
		}
	})

	t.Run("strips a closing vertex", func(t *testing.T) {
		closed := append(append([]MapPoint{}, rectangle...), rectangle[0])
		got, err := ValidateMapPinConfig(polygonConfig(closed...))
		require.NoError(t, err)
		assert.Equal(t, rectangle, got.Polygon)
	})

	t.Run("accepts collinear vertices along an edge", func(t *testing.T) {
		polygon := []MapPoint{{100, 100}, {250, 100}, {400, 100}, {400, 300}, {100, 300}}
		_, err := ValidateMapPinConfig(polygonConfig(polygon...))
		assert.NoError(t, err)
	})

	t.Run("accepts vertices on the map's edges", func(t *testing.T) {
		_, err := ValidateMapPinConfig(polygonConfig(MapPoint{0, 0}, MapPoint{1000, 0}, MapPoint{1000, 1000}))
		assert.NoError(t, err)
	})

	t.Run("accepts exactly the maximum number of vertices", func(t *testing.T) {
		_, err := ValidateMapPinConfig(polygonConfig(regularPolygon(MapPinMaxVertices)...))
		assert.NoError(t, err)
	})

	tests := []struct {
		name    string
		polygon []MapPoint
		field   string
		message string
	}{
		{name: "missing", polygon: nil, field: "polygon", message: "required"},
		{name: "two vertices", polygon: []MapPoint{{100, 100}, {300, 100}}, field: "polygon", message: "at least 3"},
		{name: "two vertices plus a closing one", polygon: []MapPoint{{100, 100}, {300, 100}, {100, 100}}, field: "polygon", message: "at least 3"},
		{name: "too many vertices", polygon: regularPolygon(MapPinMaxVertices + 1), field: "polygon", message: "more than"},
		{name: "vertex outside the map", polygon: []MapPoint{{100, 100}, {1001, 100}, {200, 300}}, field: "polygon[1]", message: "outside"},
		{name: "negative coordinate", polygon: []MapPoint{{100, -1}, {300, 100}, {200, 300}}, field: "polygon[0]", message: "outside"},
		{name: "NaN coordinate", polygon: []MapPoint{{100, 100}, {math.NaN(), 100}, {200, 300}}, field: "polygon[1]", message: "finite"},
		{name: "+Inf coordinate", polygon: []MapPoint{{100, 100}, {300, 100}, {200, math.Inf(1)}}, field: "polygon[2]", message: "finite"},
		{name: "-Inf coordinate", polygon: []MapPoint{{math.Inf(-1), 100}, {300, 100}, {200, 300}}, field: "polygon[0]", message: "finite"},
		{name: "adjacent duplicate vertex", polygon: []MapPoint{{100, 100}, {300, 100}, {300, 100}, {200, 300}}, field: "polygon", message: "same point"},
		{name: "non-adjacent duplicate vertex", polygon: []MapPoint{{100, 100}, {300, 100}, {200, 200}, {300, 300}, {100, 300}, {200, 200}}, field: "polygon", message: "same point"},
		{name: "zero area", polygon: []MapPoint{{100, 100}, {200, 200}, {300, 300}}, field: "polygon", message: "zero area"},
		// Asymmetric, so its signed area is not 0 and the intersection check is what rejects it.
		{name: "self-intersecting bow-tie", polygon: []MapPoint{{100, 100}, {300, 300}, {300, 100}, {100, 400}}, field: "polygon", message: "intersect"},
		{name: "vertex touching a non-adjacent edge", polygon: []MapPoint{{100, 100}, {300, 100}, {300, 300}, {200, 100}, {100, 300}}, field: "polygon", message: "intersect"},
		{name: "spike doubling back along an edge", polygon: []MapPoint{{100, 300}, {100, 100}, {400, 100}, {250, 100}}, field: "polygon", message: "intersect"},
		{name: "spike doubling back past its start", polygon: []MapPoint{{100, 300}, {200, 100}, {400, 100}, {150, 100}, {150, 200}}, field: "polygon", message: "intersect"},
	}

	for _, tt := range tests {
		t.Run("rejects "+tt.name, func(t *testing.T) {
			_, err := ValidateMapPinConfig(polygonConfig(tt.polygon...))
			assertMapPinError(t, err, tt.field, tt.message)
		})
	}
}

func TestValidateMapPinConfigPointRadius(t *testing.T) {
	t.Run("accepts a valid config", func(t *testing.T) {
		_, err := ValidateMapPinConfig(pointRadius())
		assert.NoError(t, err)
	})

	t.Run("accepts full radius 0 and a target on the map's edge", func(t *testing.T) {
		c := pointRadius()
		c.FullRadius = ptr(0.0)
		c.Target = &MapPoint{X: 1000, Y: 0}
		_, err := ValidateMapPinConfig(c)
		assert.NoError(t, err)
	})

	tests := []struct {
		name    string
		modify  func(*MapPinConfig)
		field   string
		message string
	}{
		{name: "missing target", modify: func(c *MapPinConfig) { c.Target = nil }, field: "target", message: "required"},
		{name: "missing full radius", modify: func(c *MapPinConfig) { c.FullRadius = nil }, field: "fullPointsRadius", message: "required"},
		{name: "missing zero radius", modify: func(c *MapPinConfig) { c.ZeroRadius = nil }, field: "zeroPointsRadius", message: "required"},
		{name: "target outside the map", modify: func(c *MapPinConfig) { c.Target = &MapPoint{X: 500, Y: 1000.5} }, field: "target", message: "outside"},
		{name: "NaN target", modify: func(c *MapPinConfig) { c.Target = &MapPoint{X: math.NaN(), Y: 300} }, field: "target", message: "finite"},
		{name: "negative full radius", modify: func(c *MapPinConfig) { c.FullRadius = ptr(-1.0) }, field: "fullPointsRadius", message: "at least 0"},
		{name: "NaN full radius", modify: func(c *MapPinConfig) { c.FullRadius = ptr(math.NaN()) }, field: "fullPointsRadius", message: "finite"},
		{name: "infinite zero radius", modify: func(c *MapPinConfig) { c.ZeroRadius = ptr(math.Inf(1)) }, field: "zeroPointsRadius", message: "finite"},
		{name: "zero radius equal to full radius", modify: func(c *MapPinConfig) { c.ZeroRadius = ptr(40.0) }, field: "zeroPointsRadius", message: "greater than"},
		{name: "zero radius below full radius", modify: func(c *MapPinConfig) { c.ZeroRadius = ptr(10.0) }, field: "zeroPointsRadius", message: "greater than"},
		{name: "polygon given", modify: func(c *MapPinConfig) { c.Polygon = rectangle }, field: "polygon", message: "only used in POLYGON"},
		{name: "empty polygon given", modify: func(c *MapPinConfig) { c.Polygon = []MapPoint{} }, field: "polygon", message: "only used in POLYGON"},
	}

	for _, tt := range tests {
		t.Run("rejects "+tt.name, func(t *testing.T) {
			c := pointRadius()
			tt.modify(&c)
			_, err := ValidateMapPinConfig(c)
			assertMapPinError(t, err, tt.field, tt.message)
		})
	}
}

func TestValidateMapPinConfigShared(t *testing.T) {
	tests := []struct {
		name    string
		config  MapPinConfig
		field   string
		message string
	}{
		{name: "zero width", config: MapPinConfig{Mode: MapPinModePolygon, Width: 0, Height: 1000, Polygon: triangle}, field: "mapWidth", message: "positive"},
		{name: "negative height", config: MapPinConfig{Mode: MapPinModePolygon, Width: 1000, Height: -1, Polygon: triangle}, field: "mapHeight", message: "positive"},
		{name: "unknown mode", config: MapPinConfig{Mode: "CIRCLE", Width: 1000, Height: 1000}, field: "scoringMode", message: "unknown"},
		{name: "empty mode", config: MapPinConfig{Width: 1000, Height: 1000}, field: "scoringMode", message: "unknown"},
		{name: "POLYGON with a target", config: withPolygon(func(c *MapPinConfig) { c.Target = &MapPoint{X: 1, Y: 1} }), field: "target", message: "only used in POINT_RADIUS"},
		{name: "POLYGON with a full radius", config: withPolygon(func(c *MapPinConfig) { c.FullRadius = ptr(0.0) }), field: "fullPointsRadius", message: "only used in POINT_RADIUS"},
		{name: "POLYGON with a zero radius", config: withPolygon(func(c *MapPinConfig) { c.ZeroRadius = ptr(10.0) }), field: "zeroPointsRadius", message: "only used in POINT_RADIUS"},
	}

	for _, tt := range tests {
		t.Run("rejects "+tt.name, func(t *testing.T) {
			_, err := ValidateMapPinConfig(tt.config)
			assertMapPinError(t, err, tt.field, tt.message)
		})
	}
}

func TestValidateMapPinConfigDoesNotModifyInput(t *testing.T) {
	closed := []MapPoint{{100, 100}, {400, 100}, {400, 300}, {100, 100}}
	got, err := ValidateMapPinConfig(polygonConfig(closed...))
	require.NoError(t, err)

	got.Polygon[0] = MapPoint{X: 1, Y: 1}
	assert.Equal(t, MapPoint{X: 100, Y: 100}, closed[0], "the normalized polygon must not share the caller's array")
}

func TestValidateMapPosition(t *testing.T) {
	tests := []struct {
		name  string
		pin   MapPoint
		valid bool
	}{
		{name: "inside", pin: MapPoint{500, 500}, valid: true},
		{name: "top-left corner", pin: MapPoint{0, 0}, valid: true},
		{name: "bottom-right corner", pin: MapPoint{1000, 800}, valid: true},
		{name: "fractional", pin: MapPoint{123.456, 789.012}, valid: true},
		{name: "left of the map", pin: MapPoint{-0.001, 500}, valid: false},
		{name: "below the map", pin: MapPoint{500, 800.001}, valid: false},
		{name: "right of the map", pin: MapPoint{1000.5, 500}, valid: false},
		{name: "NaN", pin: MapPoint{math.NaN(), 500}, valid: false},
		{name: "infinite", pin: MapPoint{500, math.Inf(1)}, valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMapPosition(tt.pin, 1000, 800)
			if tt.valid {
				assert.NoError(t, err)
				return
			}
			assertMapPinError(t, err, "mapPosition", "")
		})
	}
}

func TestMapPinIsNotGraded(t *testing.T) {
	assert.False(t, IsGradable(QuestionTypeMapPin))

	resp := Response{SelectedAnswerIDs: []string{"A"}, SubmittedOrder: []string{"A"}}
	assert.Nil(t, Grade(QuestionTypeMapPin, resp, []Answer{{ID: "A", IsCorrect: true}}))
}

func withPolygon(modify func(*MapPinConfig)) MapPinConfig {
	c := polygonConfig(triangle...)
	modify(&c)
	return c
}

// regularPolygon returns n vertices evenly spaced on a circle in the middle of
// a 1000x1000 map: a valid convex polygon of any size.
func regularPolygon(n int) []MapPoint {
	vertices := make([]MapPoint, n)
	for i := range vertices {
		angle := 2 * math.Pi * float64(i) / float64(n)
		vertices[i] = MapPoint{X: 500 + 400*math.Cos(angle), Y: 500 + 400*math.Sin(angle)}
	}
	return vertices
}

func assertMapPinError(t *testing.T, err error, field, message string) {
	t.Helper()
	var mapPinErr *MapPinError
	if !assert.ErrorAs(t, err, &mapPinErr) {
		return
	}
	assert.Equal(t, field, mapPinErr.Field)
	assert.Contains(t, mapPinErr.Message, message)
}
