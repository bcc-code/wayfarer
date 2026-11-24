package quizgrading

import (
	"fmt"
	"math"
	"slices"

	"github.com/jackc/pgx/v5/pgtype"
)

// MapPinColumns is a MAP_PIN config as stored in the map_* columns of
// quiz_questions. Positions are Postgres points (float8 underneath); Config
// turns them into whole-pixel MapPoints straight away. Every field is NULL for
// other question types.
type MapPinColumns struct {
	Mode       *string
	Width      *int32
	Height     *int32
	Target     pgtype.Point
	FullRadius *int32
	ZeroRadius *int32
	Polygon    []pgtype.Point
}

// NewMapPinColumns collects the map_* columns of a quiz_questions row, in
// table order. The sqlc row types differ per query, so each passes its fields.
func NewMapPinColumns(mode *string, width, height *int32, target pgtype.Point, fullRadius, zeroRadius *int32, polygon []pgtype.Point) MapPinColumns {
	return MapPinColumns{
		Mode:       mode,
		Width:      width,
		Height:     height,
		Target:     target,
		FullRadius: fullRadius,
		ZeroRadius: zeroRadius,
		Polygon:    polygon,
	}
}

// Columns returns the config as column values. c should be validated
// (ValidateMapPinConfig), which also keeps every value within int32; the
// fields of the other mode are NULL.
func (c MapPinConfig) Columns() MapPinColumns {
	mode := c.Mode
	cols := MapPinColumns{
		Mode:       &mode,
		Width:      Int32Ptr(&c.Width),
		Height:     Int32Ptr(&c.Height),
		Target:     PointToPG(c.Target),
		FullRadius: Int32Ptr(c.FullRadius),
		ZeroRadius: Int32Ptr(c.ZeroRadius),
	}
	if c.Polygon != nil {
		cols.Polygon = make([]pgtype.Point, len(c.Polygon))
		for i := range c.Polygon {
			cols.Polygon[i] = PointToPG(&c.Polygon[i])
		}
	}
	return cols
}

// Config reads a stored question's config back, converting points to whole
// pixels. The DB CHECK guarantees the columns its mode needs are set, but
// can't check each polygon vertex: an error means a vertex is NULL or not a
// finite number. The config is then returned without its polygon, which
// MapPinScore rejects, so a broken row can't score.
func (c MapPinColumns) Config() (MapPinConfig, error) {
	var cfg MapPinConfig
	if c.Mode != nil {
		cfg.Mode = *c.Mode
	}
	if c.Width != nil {
		cfg.Width = int(*c.Width)
	}
	if c.Height != nil {
		cfg.Height = int(*c.Height)
	}
	cfg.FullRadius = IntPtr(c.FullRadius)
	cfg.ZeroRadius = IntPtr(c.ZeroRadius)

	target, err := PointFromPG(c.Target)
	if err != nil {
		return cfg, fmt.Errorf("map_target: %w", err)
	}
	cfg.Target = target

	if c.Polygon != nil {
		polygon := make([]MapPoint, len(c.Polygon))
		for i, v := range c.Polygon {
			p, err := PointFromPG(v)
			if err != nil {
				return cfg, fmt.Errorf("map_polygon[%d]: %w", i, err)
			}
			if p == nil {
				return cfg, fmt.Errorf("map_polygon[%d] is NULL", i)
			}
			polygon[i] = *p
		}
		cfg.Polygon = polygon
	}
	return cfg, nil
}

// Equal reports whether two normalized configs are identical, polygon vertex
// order included.
func (c MapPinConfig) Equal(o MapPinConfig) bool {
	return c.Mode == o.Mode &&
		c.Width == o.Width &&
		c.Height == o.Height &&
		equalPtr(c.Target, o.Target) &&
		equalPtr(c.FullRadius, o.FullRadius) &&
		equalPtr(c.ZeroRadius, o.ZeroRadius) &&
		slices.Equal(c.Polygon, o.Polygon)
}

// PointFromPG converts a Postgres point to a whole-pixel MapPoint, rounding to
// the nearest pixel; the API only writes whole pixels, so that is exact. A
// NULL point is nil. NaN or infinity is an error rather than a pin at some
// arbitrary int.
func PointFromPG(p pgtype.Point) (*MapPoint, error) {
	if !p.Valid {
		return nil, nil
	}
	x, y := p.P.X, p.P.Y
	if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
		return nil, fmt.Errorf("point (%g, %g) is not finite", x, y)
	}
	return &MapPoint{X: int(math.Round(x)), Y: int(math.Round(y))}, nil
}

// PointToPG converts a MapPoint for a point column; nil is NULL.
func PointToPG(p *MapPoint) pgtype.Point {
	if p == nil {
		return pgtype.Point{}
	}
	return pgtype.Point{P: pgtype.Vec2{X: float64(p.X), Y: float64(p.Y)}, Valid: true}
}

// Int32Ptr converts a validated value for an INT column; nil stays nil.
func Int32Ptr(v *int) *int32 {
	if v == nil {
		return nil
	}
	n := int32(*v)
	return &n
}

// IntPtr converts an INT column's value; nil stays nil.
func IntPtr(v *int32) *int {
	if v == nil {
		return nil
	}
	n := int(*v)
	return &n
}

func equalPtr[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
