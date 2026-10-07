package api

import (
	"fmt"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// churchLeaderboardLimit counts participants in the filtered local leaderboard,
// matching the original automatic policy, with larger tiers for 100+ and 200+ people.
func churchLeaderboardLimit(participants int) int {
	if participants >= 200 {
		return 100
	}
	if participants >= 100 {
		return 50
	}
	if participants >= 50 {
		return 20
	}
	if participants >= 20 {
		return 10
	}
	return 3
}

func leaderboardLimitModeToDB(mode *model.LeaderboardLimitMode, entity model.LeaderboardEntityType, filter *model.LeaderboardFilter) (*string, error) {
	value := model.LeaderboardLimitModeManual
	if mode != nil {
		value = *mode
	}
	if !value.IsValid() {
		return nil, fmt.Errorf("invalid leaderboard limit mode")
	}
	if value == model.LeaderboardLimitModeChurchSize {
		if entity != model.LeaderboardEntityTypePersons {
			return nil, fmt.Errorf("automatic church-size limits require a persons leaderboard")
		}
		if filter == nil || filter.ChurchID == nil || *filter.ChurchID == "" {
			return nil, fmt.Errorf("automatic church-size limits require a church filter")
		}
	}
	result := string(value)
	return &result, nil
}
