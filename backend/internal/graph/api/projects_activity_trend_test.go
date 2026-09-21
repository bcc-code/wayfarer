package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// Noon UTC, so day-boundary arithmetic is not sitting on a cutover.
var trendNow = time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)

func trendRow(day string, points int64, users int32) *sqlc.GetProjectActivityTrendRow {
	parsed, err := time.Parse(time.DateOnly, day)
	if err != nil {
		panic(err)
	}
	return &sqlc.GetProjectActivityTrendRow{
		Day:         pgtype.Date{Time: parsed, Valid: true},
		Points:      points,
		ActiveUsers: users,
	}
}

func TestResolveTrendDays(t *testing.T) {
	days := func(v int) *int { return &v }

	tests := []struct {
		name string
		in   *int
		want int
	}{
		{name: "omitted falls back to the default", in: nil, want: defaultTrendDays},
		{name: "zero falls back rather than erroring", in: days(0), want: defaultTrendDays},
		{name: "negative falls back rather than erroring", in: days(-7), want: defaultTrendDays},
		{name: "a sensible value passes through", in: days(30), want: 30},
		{name: "one day is allowed", in: days(1), want: 1},
		{name: "the maximum passes through", in: days(maxTrendDays), want: maxTrendDays},
		{name: "beyond the maximum is clamped", in: days(100000), want: maxTrendDays},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveTrendDays(tt.in))
		})
	}
}

func TestTrendWindowStart(t *testing.T) {
	// A 14-day window ending today starts 13 days ago, not 14 — the window is
	// inclusive of both ends, and an off-by-one here would silently produce a
	// 15-point series.
	assert.Equal(t,
		time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
		trendWindowStart(trendNow, 14),
	)

	assert.Equal(t,
		time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		trendWindowStart(trendNow, 1),
		"a one-day window is today only",
	)
}

func TestBuildActivityTrendFillsEveryDay(t *testing.T) {
	// Two days of activity inside a seven-day window. Plotted without the gaps
	// filled, these would read as two adjacent columns — i.e. steady activity
	// — rather than two spikes four days apart.
	rows := []*sqlc.GetProjectActivityTrendRow{
		trendRow("2026-09-16", 120, 12),
		trendRow("2026-09-20", 340, 30),
	}

	trend := buildActivityTrend(rows, 7, trendNow)

	require.Len(t, trend, 7, "one point per day in the window")

	// Oldest first, contiguous, no gaps.
	for i, point := range trend {
		want := time.Date(2026, 9, 15+i, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, want.Format(time.DateOnly), point.Date.Format(time.DateOnly))
	}

	assert.Equal(t, 0, trend[0].Points, "15 Sep had no activity")
	assert.Equal(t, 120, trend[1].Points, "16 Sep")
	assert.Equal(t, 12, trend[1].ActiveUsers)
	assert.Equal(t, 0, trend[2].Points, "17 Sep had no activity")
	assert.Equal(t, 340, trend[5].Points, "20 Sep")
	assert.Equal(t, 30, trend[5].ActiveUsers)
	assert.Equal(t, 0, trend[6].Points, "21 Sep, today, had no activity yet")
}

func TestBuildActivityTrendWithNoRows(t *testing.T) {
	// A quiet project must still produce a full flat series: an empty array
	// would render as no chart at all, which looks like a broken panel rather
	// than a quiet month.
	trend := buildActivityTrend(nil, 14, trendNow)

	require.Len(t, trend, 14)
	for _, point := range trend {
		assert.Equal(t, 0, point.Points)
		assert.Equal(t, 0, point.ActiveUsers)
	}
}

func TestBuildActivityTrendIgnoresRowsOutsideWindow(t *testing.T) {
	// A stale `since` bound or a clock difference must not stretch the series.
	rows := []*sqlc.GetProjectActivityTrendRow{
		trendRow("2026-08-01", 999, 99),
		trendRow("2026-09-20", 10, 1),
	}

	trend := buildActivityTrend(rows, 3, trendNow)

	require.Len(t, trend, 3)
	assert.Equal(t, "2026-09-19", trend[0].Date.Format(time.DateOnly))
	assert.Equal(t, 0, trend[0].Points)
	assert.Equal(t, 10, trend[1].Points)

	var total int
	for _, point := range trend {
		total += point.Points
	}
	assert.Equal(t, 10, total, "the August row must not appear anywhere")
}

func TestBuildActivityTrendSkipsInvalidDates(t *testing.T) {
	// pgtype.Date zero value is Valid:false. Reading .Time off it would place a
	// row on year 1 and silently drop a real day from the series.
	rows := []*sqlc.GetProjectActivityTrendRow{
		{Day: pgtype.Date{}, Points: 500, ActiveUsers: 50},
		trendRow("2026-09-21", 7, 2),
	}

	trend := buildActivityTrend(rows, 2, trendNow)

	require.Len(t, trend, 2)
	assert.Equal(t, 0, trend[0].Points)
	assert.Equal(t, 7, trend[1].Points, "today's real row survives")
}

func TestBuildActivityTrendNormalisesRowTimestamps(t *testing.T) {
	// The column is a date, but pgtype.Date carries a time.Time — a driver or
	// timezone quirk that leaves a time-of-day on it must not miss its bucket.
	rows := []*sqlc.GetProjectActivityTrendRow{
		{
			Day:         pgtype.Date{Time: time.Date(2026, 9, 21, 14, 30, 0, 0, time.UTC), Valid: true},
			Points:      42,
			ActiveUsers: 4,
		},
	}

	trend := buildActivityTrend(rows, 2, trendNow)

	require.Len(t, trend, 2)
	assert.Equal(t, 42, trend[1].Points, "bucketed onto 21 Sep, not dropped")
}

func TestBuildActivityTrendIsStableAcrossWindowSizes(t *testing.T) {
	rows := []*sqlc.GetProjectActivityTrendRow{trendRow("2026-09-21", 5, 1)}

	for _, days := range []int{1, 7, 14, 30, maxTrendDays} {
		trend := buildActivityTrend(rows, days, trendNow)
		require.Len(t, trend, days)
		assert.Equal(t,
			"2026-09-21",
			trend[len(trend)-1].Date.Format(time.DateOnly),
			"the window always ends today",
		)
		assert.Equal(t, 5, trend[len(trend)-1].Points)
	}
}
