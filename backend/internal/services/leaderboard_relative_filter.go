package services

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

func enabled(value *bool) bool { return value != nil && *value }

// ValidateLeaderboardRelativeFilter rejects combinations the leaderboard queries
// cannot apply, rather than silently showing a broader leaderboard.
func ValidateLeaderboardRelativeFilter(filter *model.LeaderboardFilter, entity model.LeaderboardEntityType) error {
	if filter == nil {
		return nil
	}
	if age := filter.RelativeAgeRange; age != nil {
		if filter.AgeRange != nil {
			return fmt.Errorf("relativeAgeRange cannot be combined with ageRange")
		}
		if entity != model.LeaderboardEntityTypePersons {
			return fmt.Errorf("relativeAgeRange requires a persons leaderboard")
		}
		if age.YearsYounger < 0 || age.YearsOlder < 0 || age.YearsYounger > 150 || age.YearsOlder > 150 {
			return fmt.Errorf("relative age offsets must be between 0 and 150")
		}
	}
	if enabled(filter.MyChurch) {
		if filter.ChurchID != nil {
			return fmt.Errorf("myChurch cannot be combined with churchId")
		}
		if entity != model.LeaderboardEntityTypePersons && entity != model.LeaderboardEntityTypeTeams {
			return fmt.Errorf("myChurch requires a persons or teams leaderboard")
		}
	}
	if enabled(filter.MyTeam) {
		if filter.TeamID != nil {
			return fmt.Errorf("myTeam cannot be combined with teamId")
		}
		if entity != model.LeaderboardEntityTypePersons {
			return fmt.Errorf("myTeam requires a persons leaderboard")
		}
	}
	if enabled(filter.MySuperTeam) {
		if filter.SuperTeamID != nil {
			return fmt.Errorf("mySuperTeam cannot be combined with superTeamId")
		}
		if entity != model.LeaderboardEntityTypePersons {
			return fmt.Errorf("mySuperTeam requires a persons leaderboard")
		}
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
	if filter == nil || !(enabled(filter.MyChurch) || enabled(filter.MyTeam) || enabled(filter.MySuperTeam) || filter.RelativeAgeRange != nil) {
		return params, false, nil
	}
	if params.UserID == "" {
		return params, false, fmt.Errorf("relative leaderboard filters require an authenticated user")
	}
	resolved := *filter
	params.Filter = &resolved
	if enabled(filter.MyChurch) || filter.RelativeAgeRange != nil {
		user, err := s.loaders.UserByIDLoader.Load(ctx, params.UserID)()
		if err != nil {
			return params, false, err
		}
		if user == nil {
			return params, true, nil
		}
		if enabled(filter.MyChurch) {
			if user.ChurchID == "" {
				return params, true, nil
			}
			resolved.ChurchID = &user.ChurchID
		}
		if filter.RelativeAgeRange != nil {
			birthdate, err := time.Parse("2006-01-02", user.Birthdate)
			if err != nil {
				return params, false, fmt.Errorf("cannot resolve viewer age: %w", err)
			}
			age := time.Now().Year() - birthdate.Year()
			resolved.AgeRange = &model.AgeRangeInput{Min: max(0, age-filter.RelativeAgeRange.YearsYounger), Max: age + filter.RelativeAgeRange.YearsOlder}
			resolved.RelativeAgeRange = nil
		}
	}
	if enabled(filter.MyTeam) || enabled(filter.MySuperTeam) {
		projectID := params.ContextID
		if isEvent {
			event, err := s.loaders.EventByIDLoader.Load(ctx, params.ContextID)()
			if err != nil {
				return params, false, err
			}
			if event == nil {
				return params, false, fmt.Errorf("event not found")
			}
			projectID = event.ProjectID
		}
		if enabled(filter.MyTeam) {
			teams, err := s.loaders.TeamsByUserLoader.Load(ctx, params.UserID)()
			if err != nil {
				return params, false, err
			}
			for _, team := range teams {
				if team.ProjectID == projectID {
					id := team.ID
					resolved.TeamID = &id
					break
				}
			}
			if resolved.TeamID == nil {
				return params, true, nil
			}
		}
		if enabled(filter.MySuperTeam) {
			teams, err := s.loaders.SuperTeamsByUserLoader.Load(ctx, params.UserID)()
			if err != nil {
				return params, false, err
			}
			for _, team := range teams {
				if team.ProjectID == projectID {
					id := team.ID
					resolved.SuperTeamID = &id
					break
				}
			}
			if resolved.SuperTeamID == nil {
				return params, true, nil
			}
		}
	}
	resolved.MyChurch, resolved.MyTeam, resolved.MySuperTeam = nil, nil, nil
	return params, false, nil
}
