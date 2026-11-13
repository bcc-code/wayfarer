package quizgrading

import (
	"encoding/json"
	"fmt"
	"slices"
)

// MapPinColumns is a MAP_PIN config as stored in the map_* columns of
// quiz_questions. Every field is nil for other question types.
type MapPinColumns struct {
	Mode       *string
	Width      *int32
	Height     *int32
	TargetX    *float64
	TargetY    *float64
	FullRadius *float64
	ZeroRadius *float64
	// Polygon is a JSON array of {"x", "y"} objects.
	Polygon []byte
}

// Columns returns the config as column values. c should be validated
// (ValidateMapPinConfig); the fields of the other mode are nil.
func (c MapPinConfig) Columns() (MapPinColumns, error) {
	mode := c.Mode
	width, height := int32(c.Width), int32(c.Height)
	cols := MapPinColumns{
		Mode:       &mode,
		Width:      &width,
		Height:     &height,
		FullRadius: copyFloat(c.FullRadius),
		ZeroRadius: copyFloat(c.ZeroRadius),
	}
	if c.Target != nil {
		x, y := c.Target.X, c.Target.Y
		cols.TargetX, cols.TargetY = &x, &y
	}
	if c.Polygon != nil {
		polygon, err := json.Marshal(c.Polygon)
		if err != nil {
			return MapPinColumns{}, fmt.Errorf("failed to encode polygon: %w", err)
		}
		cols.Polygon = polygon
	}
	return cols, nil
}

// Config reads a stored question's config back. The DB CHECK guarantees the
// columns its mode needs are set, but not the shape of each polygon vertex: an
// error means map_polygon isn't a list of points. The config is then returned
// without its polygon, which MapPinScore rejects, so a broken row can't score.
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
	if c.TargetX != nil && c.TargetY != nil {
		cfg.Target = &MapPoint{X: *c.TargetX, Y: *c.TargetY}
	}
	cfg.FullRadius = copyFloat(c.FullRadius)
	cfg.ZeroRadius = copyFloat(c.ZeroRadius)
	if c.Polygon != nil {
		var polygon []MapPoint
		if err := json.Unmarshal(c.Polygon, &polygon); err != nil {
			return cfg, fmt.Errorf("map_polygon is not a list of points: %w", err)
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

func equalPtr[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func copyFloat(f *float64) *float64 {
	if f == nil {
		return nil
	}
	v := *f
	return &v
}
