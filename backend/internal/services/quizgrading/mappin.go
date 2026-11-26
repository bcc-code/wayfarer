package quizgrading

import (
	"errors"
	"fmt"
	"math"
)

// QuestionTypeMapPin is a question answered by dropping one pin on a map
// image. It is scored by MapPinScore, not graded: its is_correct stays nil.
const QuestionTypeMapPin = "MAP_PIN"

// Scoring modes of a MAP_PIN question.
const (
	// MapPinModePointRadius scores by distance to one point: full points
	// within FullRadius, falling linearly to 0 at ZeroRadius.
	MapPinModePointRadius = "POINT_RADIUS"
	// MapPinModePolygon gives full points inside the polygon or on its
	// boundary, and 0 outside.
	MapPinModePolygon = "POLYGON"
)

const (
	// MapPinMaxSize caps a map's width and height in pixels. It keeps every
	// cross product of two coordinate differences well inside int64, so the
	// geometry below is exact integer arithmetic.
	MapPinMaxSize = 100_000
	// MapPinMaxVertices caps a polygon's size, which also bounds the O(n²)
	// self-intersection check.
	MapPinMaxVertices = 100
)

// MapPoint is a position in the natural pixel space of the map image, in whole
// pixels: origin top-left, x to the right, y down.
type MapPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// MapPinConfig is the correct answer of a MAP_PIN question. Only the fields of
// its Mode are set; ValidateMapPinConfig rejects the other mode's fields.
type MapPinConfig struct {
	Mode   string
	Width  int
	Height int

	// POINT_RADIUS
	Target     *MapPoint
	FullRadius *int
	ZeroRadius *int

	// POLYGON, vertices in drawing order without a closing vertex.
	Polygon []MapPoint
}

// MapPinError is a map pin config or position that fails validation. Field
// names the offending input field.
type MapPinError struct {
	Field   string
	Message string
}

func (e *MapPinError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func mapPinError(field, format string, args ...any) error {
	return &MapPinError{Field: field, Message: fmt.Sprintf(format, args...)}
}

// ValidateMapPinConfig checks that c has exactly the configuration its mode
// needs and that its geometry is usable. It returns the config normalized: a
// polygon's closing vertex, if given, is stripped.
func ValidateMapPinConfig(c MapPinConfig) (MapPinConfig, error) {
	if c.Width <= 0 || c.Width > MapPinMaxSize {
		return c, mapPinError("mapWidth", "must be between 1 and %d", MapPinMaxSize)
	}
	if c.Height <= 0 || c.Height > MapPinMaxSize {
		return c, mapPinError("mapHeight", "must be between 1 and %d", MapPinMaxSize)
	}

	switch c.Mode {
	case MapPinModePointRadius:
		if c.Polygon != nil {
			return c, mapPinError("polygon", "is only used in POLYGON mode")
		}
		if err := validatePointRadius(c); err != nil {
			return c, err
		}
		return c, nil

	case MapPinModePolygon:
		switch {
		case c.Target != nil:
			return c, mapPinError("target", "is only used in POINT_RADIUS mode")
		case c.FullRadius != nil:
			return c, mapPinError("fullPointsRadius", "is only used in POINT_RADIUS mode")
		case c.ZeroRadius != nil:
			return c, mapPinError("zeroPointsRadius", "is only used in POINT_RADIUS mode")
		}
		polygon, err := normalizePolygon(c.Polygon, c.Width, c.Height)
		if err != nil {
			return c, err
		}
		c.Polygon = polygon
		return c, nil

	default:
		return c, mapPinError("scoringMode", "unknown scoring mode %q", c.Mode)
	}
}

func validatePointRadius(c MapPinConfig) error {
	switch {
	case c.Target == nil:
		return mapPinError("target", "is required in POINT_RADIUS mode")
	case c.FullRadius == nil:
		return mapPinError("fullPointsRadius", "is required in POINT_RADIUS mode")
	case c.ZeroRadius == nil:
		return mapPinError("zeroPointsRadius", "is required in POINT_RADIUS mode")
	}

	if err := checkInBounds("target", *c.Target, c.Width, c.Height); err != nil {
		return err
	}

	full, zero := *c.FullRadius, *c.ZeroRadius
	if full < 0 {
		return mapPinError("fullPointsRadius", "must be at least 0")
	}
	if zero <= full {
		return mapPinError("zeroPointsRadius", "must be greater than fullPointsRadius")
	}
	return nil
}

// normalizePolygon strips a closing vertex and checks that the rest is a
// simple polygon inside the map: 3 to MapPinMaxVertices distinct vertices,
// non-zero area, and no two edges touching except neighbours at their shared
// vertex. Convex or concave, either winding.
func normalizePolygon(polygon []MapPoint, width, height int) ([]MapPoint, error) {
	if polygon == nil {
		return nil, mapPinError("polygon", "is required in POLYGON mode")
	}

	vertices := polygon
	if n := len(vertices); n > 1 && vertices[0] == vertices[n-1] {
		vertices = vertices[:n-1]
	}
	n := len(vertices)

	if n < 3 {
		return nil, mapPinError("polygon", "needs at least 3 distinct vertices")
	}
	if n > MapPinMaxVertices {
		return nil, mapPinError("polygon", "has more than %d vertices", MapPinMaxVertices)
	}

	for i, v := range vertices {
		if err := checkInBounds(fmt.Sprintf("polygon[%d]", i), v, width, height); err != nil {
			return nil, err
		}
	}

	for i := range vertices {
		for j := i + 1; j < n; j++ {
			if vertices[i] == vertices[j] {
				return nil, mapPinError("polygon", "vertices %d and %d are the same point", i, j)
			}
		}
	}

	if doubledArea(vertices) == 0 {
		return nil, mapPinError("polygon", "has zero area")
	}

	// Only non-adjacent edges need checking. Adjacent edges overlapping (a
	// spike that doubles back) leaves the far end of one edge on the other,
	// and the edge starting there is then non-adjacent to it, so it is caught
	// below; with 3 vertices it is a zero-area triangle instead.
	for i := range n {
		a, b := vertices[i], vertices[(i+1)%n]
		for j := i + 2; j < n; j++ {
			if i == 0 && j == n-1 {
				continue // the first and last edges share vertex 0
			}
			c, d := vertices[j], vertices[(j+1)%n]
			if segmentsTouch(a, b, c, d) {
				return nil, mapPinError("polygon", "edges %d and %d intersect", i, j)
			}
		}
	}

	return append([]MapPoint(nil), vertices...), nil
}

// ValidateMapPosition checks that a submitted pin is on the map, edges
// included.
func ValidateMapPosition(p MapPoint, width, height int) error {
	return checkInBounds("mapPosition", p, width, height)
}

func checkInBounds(field string, p MapPoint, width, height int) error {
	if p.X < 0 || p.Y < 0 || p.X > width || p.Y > height {
		return mapPinError(field, "(%d, %d) is outside the %dx%d map", p.X, p.Y, width, height)
	}
	return nil
}

// MapPinScore returns the points a pin at p earns, in [0, maxPoints], and for
// POINT_RADIUS its distance to the target (nil for POLYGON). c must have
// passed ValidateMapPinConfig. A pin off the map is an error, not 0 points:
// the API never stores one, so it means broken data. A negative maxPoints is
// an error too, since it would score negative.
func MapPinScore(c MapPinConfig, maxPoints int, p MapPoint) (int, *float64, error) {
	if maxPoints < 0 {
		return 0, nil, fmt.Errorf("maxPoints must not be negative, got %d", maxPoints)
	}
	if err := ValidateMapPosition(p, c.Width, c.Height); err != nil {
		return 0, nil, err
	}

	switch c.Mode {
	case MapPinModePointRadius:
		if c.Target == nil || c.FullRadius == nil || c.ZeroRadius == nil {
			return 0, nil, errors.New("map pin config is missing its target or radii")
		}
		d := math.Hypot(float64(p.X-c.Target.X), float64(p.Y-c.Target.Y))
		return radiusPoints(d, *c.FullRadius, *c.ZeroRadius, maxPoints), &d, nil

	case MapPinModePolygon:
		if len(c.Polygon) < 3 {
			return 0, nil, errors.New("map pin config is missing its polygon")
		}
		if PointInPolygon(p, c.Polygon) {
			return maxPoints, nil, nil
		}
		return 0, nil, nil

	default:
		return 0, nil, fmt.Errorf("unknown map pin scoring mode %q", c.Mode)
	}
}

// radiusPoints is maxPoints within full, 0 from zero on, and linear in
// between, rounded to the nearest integer.
func radiusPoints(d float64, full, zero, maxPoints int) int {
	f, z := float64(full), float64(zero)
	if d <= f {
		return maxPoints
	}
	if d >= z {
		return 0
	}
	return int(math.Round(float64(maxPoints) * (z - d) / (z - f)))
}

// PointInPolygon reports whether p is inside the polygon or on its boundary.
// Works for concave polygons and either winding. Exact: integer arithmetic
// only.
func PointInPolygon(p MapPoint, polygon []MapPoint) bool {
	n := len(polygon)
	for i := range n {
		if onSegment(p, polygon[i], polygon[(i+1)%n]) {
			return true
		}
	}

	// Crossing number: count the edges a ray from p towards +x crosses. Each
	// edge spans the half-open range of y above its lower end, so a ray
	// through a vertex counts the two edges meeting there once in total when
	// they continue on opposite sides, and zero or two times when the vertex
	// is a peak or valley the ray only grazes.
	inside := false
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		a, b := polygon[i], polygon[j]
		if (a.Y > p.Y) != (b.Y > p.Y) {
			// p.X < the edge's x at p.Y, multiplied out by dy to stay exact;
			// the comparison flips when dy is negative.
			dy := int64(b.Y - a.Y)
			lhs := int64(p.X-a.X) * dy
			rhs := int64(p.Y-a.Y) * int64(b.X-a.X)
			if (dy > 0 && lhs < rhs) || (dy < 0 && lhs > rhs) {
				inside = !inside
			}
		}
	}
	return inside
}

// onSegment reports whether p lies on the segment ab, ends included.
func onSegment(p, a, b MapPoint) bool {
	return cross(a, b, p) == 0 &&
		min(a.X, b.X) <= p.X && p.X <= max(a.X, b.X) &&
		min(a.Y, b.Y) <= p.Y && p.Y <= max(a.Y, b.Y)
}

// segmentsTouch reports whether segments ab and cd share any point: a proper
// crossing, an end on the other segment, or a collinear overlap.
func segmentsTouch(a, b, c, d MapPoint) bool {
	d1, d2 := cross(c, d, a), cross(c, d, b)
	d3, d4 := cross(a, b, c), cross(a, b, d)
	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}
	return (d1 == 0 && onSegment(a, c, d)) ||
		(d2 == 0 && onSegment(b, c, d)) ||
		(d3 == 0 && onSegment(c, a, b)) ||
		(d4 == 0 && onSegment(d, a, b))
}

// cross is the z component of (a−o) × (b−o): positive when o→a→b turns
// counter-clockwise in a y-up frame. int64 can't overflow for coordinates up
// to MapPinMaxSize.
func cross(o, a, b MapPoint) int64 {
	return int64(a.X-o.X)*int64(b.Y-o.Y) - int64(a.Y-o.Y)*int64(b.X-o.X)
}

// doubledArea is twice the polygon's area (shoelace), so it stays an integer.
func doubledArea(vertices []MapPoint) int64 {
	var sum int64
	for i, a := range vertices {
		b := vertices[(i+1)%len(vertices)]
		sum += int64(a.X)*int64(b.Y) - int64(b.X)*int64(a.Y)
	}
	if sum < 0 {
		return -sum
	}
	return sum
}
