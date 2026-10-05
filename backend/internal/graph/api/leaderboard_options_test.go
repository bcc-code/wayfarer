package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestChurchLeaderboardLimit(t *testing.T) {
	for _, tt := range []struct{ participants, limit int }{{0, 3}, {19, 3}, {20, 10}, {49, 10}, {50, 20}, {99, 20}, {100, 50}, {199, 50}, {200, 100}, {500, 100}} {
		assert.Equal(t, tt.limit, churchLeaderboardLimit(tt.participants), "participants=%d", tt.participants)
	}
}

func TestLeaderboardLimitModeToDB(t *testing.T) {
	automatic := model.LeaderboardLimitModeChurchSize
	invalid := model.LeaderboardLimitMode("INVALID")
	church := &model.LeaderboardFilter{ChurchID: stringPtr("CH1")}
	for _, tt := range []struct {
		name   string
		mode   *model.LeaderboardLimitMode
		entity model.LeaderboardEntityType
		filter *model.LeaderboardFilter
		want   string
	}{
		{name: "legacy default", entity: model.LeaderboardEntityTypePersons, want: "MANUAL"},
		{name: "automatic", mode: &automatic, entity: model.LeaderboardEntityTypePersons, filter: church, want: "CHURCH_SIZE"},
		{name: "missing church", mode: &automatic, entity: model.LeaderboardEntityTypePersons},
		{name: "empty church", mode: &automatic, entity: model.LeaderboardEntityTypePersons, filter: &model.LeaderboardFilter{ChurchID: stringPtr("")}},
		{name: "wrong entity", mode: &automatic, entity: model.LeaderboardEntityTypeTeams, filter: church},
		{name: "invalid mode", mode: &invalid},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := leaderboardLimitModeToDB(tt.mode, tt.entity, tt.filter)
			if tt.want == "" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestConfiguredLeaderboardAutomaticLimits(t *testing.T) {
	for _, isEvent := range []bool{false, true} {
		for _, size := range []int{0, 19, 20, 49, 50, 99, 100, 199, 200, 500} {
			t.Run(fmt.Sprintf("event=%v/participants=%d", isEvent, size), func(t *testing.T) {
				queries := mocks.NewMockLeaderboardQuerier(t)
				c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
				require.NoError(t, err)
				t.Cleanup(c.Close)
				rows := make([]*sqlc.GetFullProjectPersonLeaderboardRow, size)
				eventRows := make([]*sqlc.GetFullEventPersonLeaderboardRow, size)
				for i := range rows {
					rows[i] = &sqlc.GetFullProjectPersonLeaderboardRow{EntityID: fmt.Sprintf("US%026d", i+1), Rank: int64(i + 1), Score: int64(size - i), ChurchID: "CH1"}
					eventRows[i] = (*sqlc.GetFullEventPersonLeaderboardRow)(rows[i])
				}
				if isEvent {
					queries.On("GetFullEventPersonLeaderboard", mock.Anything, mock.MatchedBy(func(p sqlc.GetFullEventPersonLeaderboardParams) bool { return p.Churchid == "CH1" })).Return(eventRows, nil)
				} else {
					queries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.MatchedBy(func(p sqlc.GetFullProjectPersonLeaderboardParams) bool { return p.Churchid == "CH1" })).Return(rows, nil)
				}
				config := &model.LeaderboardConfig{ProjectID: "PR1", EntityType: model.LeaderboardEntityTypePersons, LimitMode: model.LeaderboardLimitModeChurchSize, Filter: &model.LeaderboardFilterView{ChurchID: stringPtr("CH1")}}
				if isEvent {
					config.EventID = stringPtr("EV1")
				}
				userID := fmt.Sprintf("US%026d", size)
				ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)
				r := &Resolver{LeaderboardService: services.NewLeaderboardService(queries, c, nil)}
				for _, cap := range []*int{nil, intPtr(2), intPtr(15), intPtr(1000)} {
					config.MaxEntries = cap
					count := min(size, churchLeaderboardLimit(size))
					if cap != nil {
						count = min(count, *cap)
					}
					for _, page := range []struct {
						name           string
						first, last    *int
						after, before  *string
						want           int
						previous, next bool
					}{
						{name: "default", want: count},
						{name: "oversized", first: intPtr(1000), want: count},
						{name: "past cap", first: intPtr(1000), after: stringPtr(fmt.Sprint(max(1, count))), previous: count > 0},
						{name: "backward", last: intPtr(1000), want: count},
						{name: "before past cap", last: intPtr(1000), before: stringPtr("9999"), want: count},
						{name: "first page", first: intPtr(2), want: min(count, 2), next: count > 2},
					} {
						t.Run(page.name, func(t *testing.T) {
							board, err := r.getLeaderboardForConfig(ctx, config, page.first, page.after, page.last, page.before)
							require.NoError(t, err)
							require.Len(t, board.Edges, page.want)
							assert.Equal(t, size, board.TotalCount)
							assert.Equal(t, page.previous, board.PageInfo.HasPreviousPage)
							assert.Equal(t, page.next, board.PageInfo.HasNextPage)
							if size > 0 {
								require.NotNil(t, board.Me)
								assert.Equal(t, userID, board.Me.ID)
							}
						})
					}
				}
				if size > 0 {
					board, err := r.getLeaderboardForConfig(ctx, config, nil, nil, nil, nil)
					require.NoError(t, err)
					rivals, err := r.LeaderboardConnection().NearestChurchRivals(ctx, board, intPtr(3))
					require.NoError(t, err)
					require.Len(t, rivals, 3)
					assert.Equal(t, size-1, *rivals[0].Rank)
				}
			})
		}
	}
}
