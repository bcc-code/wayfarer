package scalars

import (
	"fmt"
	"io"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jackc/pgx/v5/pgtype"
)

// ToDateTimePointer converts a pgtype.Timestamptz to *DateTime
func ToDateTimePointer(ts pgtype.Timestamptz) *DateTime {
	if !ts.Valid {
		return nil
	}
	return &DateTime{Time: ts.Time}
}

// DateTime scalar representing RFC3339 datetime
type DateTime struct {
	time.Time
}

// MarshalGQL implements the graphql.Marshaler interface
func (d DateTime) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + d.Format(time.RFC3339) + `"`))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (d *DateTime) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("DateTime must be a string")
	}

	// Try parsing as RFC3339 first (e.g., "2025-12-12T00:00:00Z")
	t, err := time.Parse(time.RFC3339, str)
	if err == nil {
		d.Time = t
		return nil
	}

	// Try parsing as date only (e.g., "2025-12-12"), assume UTC midnight
	t, err = time.Parse("2006-01-02", str)
	if err == nil {
		d.Time = t
		return nil
	}

	// Try parsing as datetime-local format (e.g., "2025-12-12T00:00"), assume UTC
	t, err = time.Parse("2006-01-02T15:04", str)
	if err == nil {
		d.Time = t
		return nil
	}

	// Try parsing as datetime-local with seconds (e.g., "2025-12-12T00:00:00"), assume UTC
	t, err = time.Parse("2006-01-02T15:04:05", str)
	if err == nil {
		d.Time = t
		return nil
	}

	return fmt.Errorf("invalid DateTime format: %w", err)
}

// Date scalar representing YYYY-MM-DD date
type Date struct {
	time.Time
}

// MarshalGQL implements the graphql.Marshaler interface
func (d Date) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + d.Format("2006-01-02") + `"`))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (d *Date) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("Date must be a string")
	}

	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return fmt.Errorf("invalid Date format: %w", err)
	}

	d.Time = t
	return nil
}

// HTML scalar representing sanitized HTML content
type HTML string

// MarshalGQL implements the graphql.Marshaler interface
func (h HTML) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(h)).MarshalGQL(w)
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (h *HTML) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("HTML must be a string")
	}
	*h = HTML(str)
	return nil
}

// Markdown scalar representing markdown content
type Markdown string

// MarshalGQL implements the graphql.Marshaler interface
func (m Markdown) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(m)).MarshalGQL(w)
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (m *Markdown) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("Markdown must be a string")
	}
	*m = Markdown(str)
	return nil
}
