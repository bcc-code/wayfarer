package utils

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Int32PtrToIntPtr converts a *int32 to *int
func Int32PtrToIntPtr(i *int32) *int {
	if i == nil {
		return nil
	}
	v := int(*i)
	return &v
}

// TimestamptzToPtr converts a pgtype.Timestamptz to *time.Time
func TimestamptzToPtr(ts pgtype.Timestamptz) *time.Time {
	if ts.Valid {
		return &ts.Time
	}
	return nil
}
