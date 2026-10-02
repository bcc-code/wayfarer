package services

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

func enabled(value *bool) bool { return value != nil && *value }

func hasRelativeFilter(f *model.LeaderboardFilter) bool {
	return f != nil && (enabled(f.MyChurch) || enabled(f.MyTeam) || enabled(f.MySuperTeam) || f.RelativeAgeRange != nil)
}

// relativeFilterRule describes one viewer-relative filter: the fixed filter it
// cannot be combined with and the leaderboard entity types it supports.
type relativeFilterRule struct {
	field            string // GraphQL name of the relative filter
	isSet            bool
	conflictingField string // GraphQL name of the fixed filter it replaces
	hasConflict      bool
	allowedEntities  []model.LeaderboardEntityType
	allowedLabel     string // human-readable allowedEntities, for errors
}

// ValidateLeaderboardRelativeFilter rejects combinations the leaderboard queries
// cannot apply, rather than silently showing a broader leaderboard.
func ValidateLeaderboardRelativeFilter(filter *model.LeaderboardFilter, entity model.LeaderboardEntityType) error {
	if filter == nil {
		return nil
	}
	personsOnly := []model.LeaderboardEntityType{model.LeaderboardEntityTypePersons}
	rules := []relativeFilterRule{
		{
			field:            "relativeAgeRange",
			isSet:            filter.RelativeAgeRange != nil,
			conflictingField: "ageRange",
			hasConflict:      filter.AgeRange != nil,
			allowedEntities:  personsOnly,
			allowedLabel:     "persons",
		},
		{
			field:            "myChurch",
			isSet:            enabled(filter.MyChurch),
			conflictingField: "churchId",
			hasConflict:      filter.ChurchID != nil,
			allowedEntities:  []model.LeaderboardEntityType{model.LeaderboardEntityTypePersons, model.LeaderboardEntityTypeTeams},
			allowedLabel:     "persons or teams",
		},
		{
			field:            "myTeam",
			isSet:            enabled(filter.MyTeam),
			conflictingField: "teamId",
			hasConflict:      filter.TeamID != nil,
			allowedEntities:  personsOnly,
			allowedLabel:     "persons",
		},
		{
			field:            "mySuperTeam",
			isSet:            enabled(filter.MySuperTeam),
			conflictingField: "superTeamId",
			hasConflict:      filter.SuperTeamID != nil,
			allowedEntities:  personsOnly,
			allowedLabel:     "persons",
		},
	}
	for _, rule := range rules {
		if !rule.isSet {
			continue
		}
		if rule.hasConflict {
			return fmt.Errorf("%s cannot be combined with %s", rule.field, rule.conflictingField)
		}
		if !slices.Contains(rule.allowedEntities, entity) {
			return fmt.Errorf("%s requires a %s leaderboard", rule.field, rule.allowedLabel)
		}
	}
	if age := filter.RelativeAgeRange; age != nil && (age.YearsYounger < 0 || age.YearsOlder < 0 || age.YearsYounger > 150 || age.YearsOlder > 150) {
		return fmt.Errorf("relative age offsets must be between 0 and 150")
	}
	return nil
}

// resolveRelativeFilter makes a request-local copy before resolving memberships.
// Cache keys therefore use concrete IDs, never shared viewer-relative flags.
// The boolean reports a missing membership: callers must return an empty board.
func (s *LeaderboardService) resolveRelativeFilter(ctx context.Context, params LeaderboardParams, isEvent bool) (LeaderboardParams, bool, error) {
	filter := params.Filter
	if err := ValidateLeaderboardRelativeFilter(filter, params.EntityType); err != nil {
		return params, false, err
	}
	if !hasRelativeFilter(filter) {
		return params, false, nil
	}
	if params.UserID == "" {
		return params, false, fmt.Errorf("relative leaderboard filters require an authenticated user")
	}
	resolved := *filter
	resolved.MyChurch, resolved.MyTeam, resolved.MySuperTeam, resolved.RelativeAgeRange = nil, nil, nil, nil
	params.Filter = &resolved

	if enabled(filter.MyChurch) || filter.RelativeAgeRange != nil {
		user, err := s.loaders.UserByIDLoader.Load(ctx, params.UserID)()
		if err != nil {
			return params, false, err
		}
		if user == nil || (enabled(filter.MyChurch) && user.ChurchID == "") {
			return params, true, nil
		}
		if enabled(filter.MyChurch) {
			resolved.ChurchID = &user.ChurchID
		}
		if offsets := filter.RelativeAgeRange; offsets != nil {
			age, err := viewerAge(user.Birthdate)
			if err != nil {
				return params, false, err
			}
			resolved.AgeRange = &model.AgeRangeInput{Min: max(0, age-offsets.YearsYounger), Max: age + offsets.YearsOlder}
		}
	}

	if !enabled(filter.MyTeam) && !enabled(filter.MySuperTeam) {
		return params, false, nil
	}
	projectID := params.ContextID
	if isEvent {
		projectID = s.eventProjectID(ctx, params.ContextID)
	}
	if enabled(filter.MyTeam) {
		teams, err := s.loaders.TeamsByUserLoader.Load(ctx, params.UserID)()
		if err != nil {
			return params, false, err
		}
		if resolved.TeamID = idInProject(teams, projectID, func(t *model.Team) (string, string) { return t.ID, t.ProjectID }); resolved.TeamID == nil {
			return params, true, nil
		}
	}
	if enabled(filter.MySuperTeam) {
		superTeams, err := s.loaders.SuperTeamsByUserLoader.Load(ctx, params.UserID)()
		if err != nil {
			return params, false, err
		}
		if resolved.SuperTeamID = idInProject(superTeams, projectID, func(t *model.SuperTeam) (string, string) { return t.ID, t.ProjectID }); resolved.SuperTeamID == nil {
			return params, true, nil
		}
	}
	return params, false, nil
}

// viewerAge returns the viewer's age as a calendar-year difference.
func viewerAge(birthdate string) (int, error) {
	born, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve viewer age: %w", err)
	}
	return time.Now().Year() - born.Year(), nil
}

// idInProject returns the ID of the first item belonging to projectID, or nil.
func idInProject[T any](items []T, projectID string, fields func(T) (id, project string)) *string {
	for _, item := range items {
		if id, project := fields(item); project == projectID {
			return &id
		}
	}
	return nil
}
