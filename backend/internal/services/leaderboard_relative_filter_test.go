package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func relativeTestLoader[T any](values map[string]T) *dataloader.Loader[string, T] {
	return dataloader.NewBatchedLoader(func(_ context.Context, keys []string) []*dataloader.Result[T] {
		results := make([]*dataloader.Result[T], len(keys))
		for i, key := range keys {
			results[i] = &dataloader.Result[T]{Data: values[key]}
		}
		return results
	})
}

func TestRelativeLeaderboardFilters(t *testing.T) {
	yes := true
	filter := &model.LeaderboardFilter{MyChurch: &yes, MyTeam: &yes, MySuperTeam: &yes, RelativeAgeRange: &model.RelativeAgeRangeInput{YearsYounger: 3, YearsOlder: 2}}
	ldrs := &loaders.Loaders{
		UserByIDLoader:         relativeTestLoader(map[string]*model.User{"u": {ChurchID: "church", Birthdate: fmt.Sprintf("%d-12-31", time.Now().Year()-20)}}),
		TeamsByUserLoader:      relativeTestLoader(map[string][]*model.Team{"u": {{ID: "wrong", ProjectID: "other"}, {ID: "team", ProjectID: "project"}}}),
		SuperTeamsByUserLoader: relativeTestLoader(map[string][]*model.SuperTeam{"u": {{ID: "super", ProjectID: "project"}}}),
		EventByIDLoader:        relativeTestLoader(map[string]*model.Event{"event": {ProjectID: "project"}}),
	}
	s := NewLeaderboardService(nil, newTestCache(), ldrs)
	for _, isEvent := range []bool{false, true} {
		id := "project"
		if isEvent {
			id = "event"
		}
		resolved, empty, err := s.resolveRelativeFilter(context.Background(), LeaderboardParams{ContextID: id, UserID: "u", EntityType: model.LeaderboardEntityTypePersons, Filter: filter}, isEvent)
		require.NoError(t, err)
		require.False(t, empty)
		require.Equal(t, "church", *resolved.Filter.ChurchID)
		require.Equal(t, "team", *resolved.Filter.TeamID)
		require.Equal(t, "super", *resolved.Filter.SuperTeamID)
		require.Equal(t, &model.AgeRangeInput{Min: 17, Max: 22}, resolved.Filter.AgeRange)
		require.Nil(t, resolved.Filter.MyChurch)
	}
	require.Nil(t, filter.ChurchID, "persisted filter must remain viewer-independent")
	require.Nil(t, filter.AgeRange)
}

func TestRelativeLeaderboardMissingMembership(t *testing.T) {
	yes := true
	s := NewLeaderboardService(nil, newTestCache(), &loaders.Loaders{TeamsByUserLoader: relativeTestLoader(map[string][]*model.Team{})})
	params := LeaderboardParams{ContextID: "project", UserID: "u", EntityType: model.LeaderboardEntityTypePersons, Filter: &model.LeaderboardFilter{MyTeam: &yes}}
	entries, me, total, err := s.GetProjectLeaderboard(context.Background(), params)
	require.NoError(t, err)
	require.Empty(t, entries)
	require.Nil(t, me)
	require.Zero(t, total)
	rivals, err := s.NearestChurchRivals(context.Background(), params, false, 5)
	require.NoError(t, err)
	require.Empty(t, rivals)
}

func TestRelativeLeaderboardCacheIsolation(t *testing.T) {
	yes := true
	queries := mocks.NewMockLeaderboardQuerier(t)
	s := NewLeaderboardService(queries, newTestCache(), &loaders.Loaders{UserByIDLoader: relativeTestLoader(map[string]*model.User{"a": {ChurchID: "ca"}, "b": {ChurchID: "cb"}})})
	for _, id := range []string{"a", "b"} {
		queries.On("GetFullProjectPersonLeaderboard", mock.Anything, mock.MatchedBy(func(p sqlc.GetFullProjectPersonLeaderboardParams) bool { return p.Churchid == "c"+id })).Return([]*sqlc.GetFullProjectPersonLeaderboardRow{personRow(id, 1, 10)}, nil).Once()
	}
	filter := &model.LeaderboardFilter{MyChurch: &yes}
	for _, id := range []string{"a", "b", "a"} {
		params := personLeaderboardParams(id)
		params.Filter = filter
		entries, _, _, err := s.GetProjectLeaderboard(context.Background(), params)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		require.Equal(t, id, entries[0].EntityID)
		s.cache.Wait()
	}
}

func TestValidateRelativeLeaderboardFilters(t *testing.T) {
	yes := true
	id := "fixed"
	for _, filter := range []*model.LeaderboardFilter{
		{MyChurch: &yes, ChurchID: &id}, {MyTeam: &yes, TeamID: &id}, {MySuperTeam: &yes, SuperTeamID: &id},
		{RelativeAgeRange: &model.RelativeAgeRangeInput{YearsYounger: -1}},
		{RelativeAgeRange: &model.RelativeAgeRangeInput{YearsOlder: 151}},
		{RelativeAgeRange: &model.RelativeAgeRangeInput{}, AgeRange: &model.AgeRangeInput{}},
	} {
		require.Error(t, ValidateLeaderboardRelativeFilter(filter, model.LeaderboardEntityTypePersons))
	}
	require.Error(t, ValidateLeaderboardRelativeFilter(&model.LeaderboardFilter{MyTeam: &yes}, model.LeaderboardEntityTypeTeams))
	require.NoError(t, ValidateLeaderboardRelativeFilter(&model.LeaderboardFilter{RelativeAgeRange: &model.RelativeAgeRangeInput{}}, model.LeaderboardEntityTypePersons))
}

func TestIDInProject(t *testing.T) {
	fields := func(t *model.Team) (string, string) { return t.ID, t.ProjectID }
	teams := []*model.Team{{ID: "a", ProjectID: "other"}, {ID: "b", ProjectID: "p"}, {ID: "c", ProjectID: "p"}}
	require.Equal(t, "b", *idInProject(teams, "p", fields))
	require.Nil(t, idInProject(teams, "missing", fields))
	require.Nil(t, idInProject(nil, "p", fields))
}
