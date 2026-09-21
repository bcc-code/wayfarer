package api

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/jackc/pgx/v5/pgtype"
)

// buildProjectFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildProjectFilterParamsCursor(filter *model.ProjectFilter, first *int, after *string, last *int, before *string) (sqlc.GetProjectsFilteredCursorParams, error) {
	params := sqlc.GetProjectsFilteredCursorParams{}

	// Apply filters if provided
	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		params.Archived = filter.Archived

		if filter.StartDateAfter != nil {
			params.Startdateafter = pgtype.Timestamptz{
				Time:  filter.StartDateAfter.Time,
				Valid: true,
			}
		}

		if filter.StartDateBefore != nil {
			params.Startdatebefore = pgtype.Timestamptz{
				Time:  filter.StartDateBefore.Time,
				Valid: true,
			}
		}

		if filter.EndDateAfter != nil {
			params.Enddateafter = pgtype.Timestamptz{
				Time:  filter.EndDateAfter.Time,
				Valid: true,
			}
		}

		if filter.EndDateBefore != nil {
			params.Enddatebefore = pgtype.Timestamptz{
				Time:  filter.EndDateBefore.Time,
				Valid: true,
			}
		}
	}

	// Handle cursor pagination
	isBackward := false
	var limit int

	if first != nil && last != nil {
		return params, fmt.Errorf("cannot specify both first and last")
	}

	if first != nil {
		limit = *first + 1 // Fetch one extra to determine hasNextPage
		isBackward = false
	} else if last != nil {
		limit = *last + 1 // Fetch one extra to determine hasPreviousPage
		isBackward = true
	} else {
		// Default page size
		limit = 11 // 10 items + 1 to check for next page
		isBackward = false
	}

	params.Querylimit = int32(limit)
	params.Isbackward = isBackward

	// Set cursors
	if after != nil && *after != "" {
		params.Aftercursor = *after
	}

	if before != nil && *before != "" {
		params.Beforecursor = *before
	}

	return params, nil
}

// buildCountProjectsFilterParams converts GraphQL filter to database count query parameters
func buildCountProjectsFilterParams(filter *model.ProjectFilter) sqlc.CountProjectsFilteredParams {
	params := sqlc.CountProjectsFilteredParams{}

	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		params.Archived = filter.Archived

		if filter.StartDateAfter != nil {
			params.Startdateafter = pgtype.Timestamptz{
				Time:  filter.StartDateAfter.Time,
				Valid: true,
			}
		}

		if filter.StartDateBefore != nil {
			params.Startdatebefore = pgtype.Timestamptz{
				Time:  filter.StartDateBefore.Time,
				Valid: true,
			}
		}

		if filter.EndDateAfter != nil {
			params.Enddateafter = pgtype.Timestamptz{
				Time:  filter.EndDateAfter.Time,
				Valid: true,
			}
		}

		if filter.EndDateBefore != nil {
			params.Enddatebefore = pgtype.Timestamptz{
				Time:  filter.EndDateBefore.Time,
				Valid: true,
			}
		}
	}

	return params
}

// buildProjectCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildProjectCacheKeyParams(filter *model.ProjectFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if len(filter.Ids) > 0 {
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
		}
		if filter.Archived != nil {
			params["archived"] = fmt.Sprintf("%t", *filter.Archived)
		}
		if filter.StartDateAfter != nil {
			params["startdateafter"] = filter.StartDateAfter.Format("2006-01-02T15:04:05Z07:00")
		}
		if filter.StartDateBefore != nil {
			params["startdatebefore"] = filter.StartDateBefore.Format("2006-01-02T15:04:05Z07:00")
		}
		if filter.EndDateAfter != nil {
			params["enddateafter"] = filter.EndDateAfter.Format("2006-01-02T15:04:05Z07:00")
		}
		if filter.EndDateBefore != nil {
			params["enddatebefore"] = filter.EndDateBefore.Format("2006-01-02T15:04:05Z07:00")
		}
	}

	// Add pagination parameters
	if first != nil {
		params["first"] = fmt.Sprintf("%d", *first)
	}
	if after != nil && *after != "" {
		params["after"] = *after
	}
	if last != nil {
		params["last"] = fmt.Sprintf("%d", *last)
	}
	if before != nil && *before != "" {
		params["before"] = *before
	}

	return params
}

// ==================== Project activity trend ====================

const (
	// defaultTrendDays is what `activityTrend` returns when the client asks for
	// no particular window.
	defaultTrendDays = 14
	// maxTrendDays bounds the window. The field is a sparkline feed, so a
	// client asking for years of daily rows is a mistake rather than a use
	// case, and an unbounded value is an easy way to make the server do work
	// proportional to whatever a caller types.
	maxTrendDays = 90
)

// resolveTrendDays applies the default and clamps the window.
//
// A nil value means the argument was omitted. Non-positive values are treated
// as the default rather than as an error: an empty series is a worse answer
// than a sensible one, and there is nothing a caller could do with the error.
func resolveTrendDays(days *int) int {
	if days == nil || *days <= 0 {
		return defaultTrendDays
	}
	if *days > maxTrendDays {
		return maxTrendDays
	}
	return *days
}

// trendWindowStart is the first day (UTC midnight) of a `days`-long window that
// ends on the day containing `now`. Used both to build the query's `since`
// bound and to fill the series, so the two cannot disagree about the window.
func trendWindowStart(now time.Time, days int) time.Time {
	today := now.UTC().Truncate(24 * time.Hour)
	return today.AddDate(0, 0, -(days - 1))
}

// buildActivityTrend expands sparse daily rows into one point per day across
// the whole window, oldest first.
//
// The query only returns days that have activity, which would make a sparkline
// lie: three points a week apart plotted as adjacent columns reads as steady
// activity rather than as three isolated spikes. Filling here rather than with
// a generate_series join keeps the SQL a plain grouped scan and makes the
// windowing unit-testable without a database.
//
// Rows outside the window are ignored rather than trusted, so a stale `since`
// or a clock skew cannot stretch the series beyond `days` points.
func buildActivityTrend(
	rows []*sqlc.GetProjectActivityTrendRow,
	days int,
	now time.Time,
) []model.ProjectActivityPoint {
	start := trendWindowStart(now, days)

	type dayTotals struct {
		points      int
		activeUsers int
	}
	byDay := make(map[string]dayTotals, len(rows))
	for _, row := range rows {
		if !row.Day.Valid {
			continue
		}
		day := row.Day.Time.UTC().Truncate(24 * time.Hour)
		if day.Before(start) {
			continue
		}
		byDay[day.Format(time.DateOnly)] = dayTotals{
			points:      int(row.Points),
			activeUsers: int(row.ActiveUsers),
		}
	}

	points := make([]model.ProjectActivityPoint, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i)
		totals := byDay[day.Format(time.DateOnly)]
		points = append(points, model.ProjectActivityPoint{
			Date:        scalars.Date{Time: day},
			Points:      totals.points,
			ActiveUsers: totals.activeUsers,
		})
	}
	return points
}

// activityTrend backs the `Project.activityTrend` resolver.
//
// The body lives here rather than in `projects.resolvers.go` because that file
// is generated: keeping it thin means `make generate` never has to preserve
// anything but a one-line delegation, and the generated file's import list
// stays as gqlgen wrote it.
func (r *projectResolver) activityTrend(
	ctx context.Context,
	projectID string,
	days *int,
) ([]model.ProjectActivityPoint, error) {
	window := resolveTrendDays(days)
	now := time.Now()

	cacheKey := fmt.Sprintf("project:%s:activity-trend:%d", projectID, window)
	if cached, ok := r.Cache.Get(cacheKey); ok {
		if trend, ok := cached.([]model.ProjectActivityPoint); ok {
			return trend, nil
		}
	}

	rows, err := r.DB.Queries.GetProjectActivityTrend(ctx, sqlc.GetProjectActivityTrendParams{
		ProjectID: projectID,
		Since: pgtype.Timestamptz{
			// The same window start the series is built from, so the query
			// cannot return a day the series has no slot for.
			Time:  trendWindowStart(now, window),
			Valid: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get activity trend for project %s: %w", projectID, err)
	}

	trend := buildActivityTrend(rows, window, now)
	r.Cache.Set(cacheKey, trend)

	return trend, nil
}
