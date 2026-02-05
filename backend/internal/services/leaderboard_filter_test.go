package services

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
)

func intPtr(i int) *int {
	return &i
}

func TestBuildFullProjectPersonParams_WithTeamId(t *testing.T) {
	s := &LeaderboardService{}

	params := LeaderboardParams{
		ContextID: "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Filter: &model.LeaderboardFilter{
			TeamID: stringPtr("TM01ARZ3NDEKTSV4RRFFQ69G5FAV"),
		},
	}

	result := s.buildFullProjectPersonParams(params)

	assert.Equal(t, "PR01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Projectid)
	assert.Equal(t, "TM01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Teamid)
	assert.Equal(t, "", result.Superteamid)
}

func TestBuildFullProjectPersonParams_WithSuperTeamId(t *testing.T) {
	s := &LeaderboardService{}

	params := LeaderboardParams{
		ContextID: "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Filter: &model.LeaderboardFilter{
			SuperTeamID: stringPtr("ST01ARZ3NDEKTSV4RRFFQ69G5FAV"),
		},
	}

	result := s.buildFullProjectPersonParams(params)

	assert.Equal(t, "PR01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Projectid)
	assert.Equal(t, "", result.Teamid)
	assert.Equal(t, "ST01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Superteamid)
}

func TestBuildFullProjectPersonParams_WithBothTeamFilters(t *testing.T) {
	s := &LeaderboardService{}

	params := LeaderboardParams{
		ContextID: "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Filter: &model.LeaderboardFilter{
			TeamID:      stringPtr("TM01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			SuperTeamID: stringPtr("ST01ARZ3NDEKTSV4RRFFQ69G5FAV"),
		},
	}

	result := s.buildFullProjectPersonParams(params)

	assert.Equal(t, "TM01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Teamid)
	assert.Equal(t, "ST01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Superteamid)
}

func TestBuildFullProjectPersonParams_WithAllFilters(t *testing.T) {
	s := &LeaderboardService{}

	params := LeaderboardParams{
		ContextID: "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Filter: &model.LeaderboardFilter{
			ChurchID:    stringPtr("CH01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			MinScore:    intPtr(100),
			MaxScore:    intPtr(500),
			TeamID:      stringPtr("TM01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			SuperTeamID: stringPtr("ST01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			AgeRange:    &model.AgeRangeInput{Min: 18, Max: 30},
		},
	}

	result := s.buildFullProjectPersonParams(params)

	assert.Equal(t, "PR01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Projectid)
	assert.Equal(t, "CH01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Churchid)
	assert.Equal(t, int32(100), result.Minscore)
	assert.Equal(t, int32(500), result.Maxscore)
	assert.Equal(t, int32(18), result.Minage)
	assert.Equal(t, int32(30), result.Maxage)
	assert.Equal(t, "TM01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Teamid)
	assert.Equal(t, "ST01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Superteamid)
}

func TestBuildFullProjectPersonParams_NilFilter(t *testing.T) {
	s := &LeaderboardService{}

	params := LeaderboardParams{
		ContextID: "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Filter:    nil,
	}

	result := s.buildFullProjectPersonParams(params)

	assert.Equal(t, "PR01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Projectid)
	assert.Equal(t, "", result.Teamid)
	assert.Equal(t, "", result.Superteamid)
	assert.Equal(t, "", result.Churchid)
}

func TestBuildFullEventPersonParams_WithTeamId(t *testing.T) {
	s := &LeaderboardService{}

	params := LeaderboardParams{
		ContextID: "EV01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Filter: &model.LeaderboardFilter{
			TeamID: stringPtr("TM01ARZ3NDEKTSV4RRFFQ69G5FAV"),
		},
	}

	result := s.buildFullEventPersonParams(params)

	assert.Equal(t, "EV01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Eventid)
	assert.Equal(t, "TM01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Teamid)
	assert.Equal(t, "", result.Superteamid)
}

func TestBuildFullEventPersonParams_WithSuperTeamId(t *testing.T) {
	s := &LeaderboardService{}

	params := LeaderboardParams{
		ContextID: "EV01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Filter: &model.LeaderboardFilter{
			SuperTeamID: stringPtr("ST01ARZ3NDEKTSV4RRFFQ69G5FAV"),
		},
	}

	result := s.buildFullEventPersonParams(params)

	assert.Equal(t, "EV01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Eventid)
	assert.Equal(t, "", result.Teamid)
	assert.Equal(t, "ST01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Superteamid)
}

func TestBuildFullEventPersonParams_WithBothTeamFilters(t *testing.T) {
	s := &LeaderboardService{}

	params := LeaderboardParams{
		ContextID: "EV01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Filter: &model.LeaderboardFilter{
			TeamID:      stringPtr("TM01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			SuperTeamID: stringPtr("ST01ARZ3NDEKTSV4RRFFQ69G5FAV"),
		},
	}

	result := s.buildFullEventPersonParams(params)

	assert.Equal(t, "TM01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Teamid)
	assert.Equal(t, "ST01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Superteamid)
}

func TestBuildFullEventPersonParams_WithAllFilters(t *testing.T) {
	s := &LeaderboardService{}

	params := LeaderboardParams{
		ContextID: "EV01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Filter: &model.LeaderboardFilter{
			ChurchID:    stringPtr("CH01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			MinScore:    intPtr(50),
			MaxScore:    intPtr(1000),
			TeamID:      stringPtr("TM01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			SuperTeamID: stringPtr("ST01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			AgeRange:    &model.AgeRangeInput{Min: 20, Max: 40},
		},
	}

	result := s.buildFullEventPersonParams(params)

	assert.Equal(t, "EV01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Eventid)
	assert.Equal(t, "CH01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Churchid)
	assert.Equal(t, int32(50), result.Minscore)
	assert.Equal(t, int32(1000), result.Maxscore)
	assert.Equal(t, int32(20), result.Minage)
	assert.Equal(t, int32(40), result.Maxage)
	assert.Equal(t, "TM01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Teamid)
	assert.Equal(t, "ST01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Superteamid)
}

func TestBuildFullEventPersonParams_NilFilter(t *testing.T) {
	s := &LeaderboardService{}

	params := LeaderboardParams{
		ContextID: "EV01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Filter:    nil,
	}

	result := s.buildFullEventPersonParams(params)

	assert.Equal(t, "EV01ARZ3NDEKTSV4RRFFQ69G5FAV", result.Eventid)
	assert.Equal(t, "", result.Teamid)
	assert.Equal(t, "", result.Superteamid)
	assert.Equal(t, "", result.Churchid)
}

func TestGetFilterString_TeamId(t *testing.T) {
	tests := []struct {
		name     string
		filter   *model.LeaderboardFilter
		field    string
		expected string
	}{
		{
			name:     "nil filter returns empty",
			filter:   nil,
			field:    "teamId",
			expected: "",
		},
		{
			name:     "nil teamId returns empty",
			filter:   &model.LeaderboardFilter{},
			field:    "teamId",
			expected: "",
		},
		{
			name: "valid teamId",
			filter: &model.LeaderboardFilter{
				TeamID: stringPtr("TM01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			},
			field:    "teamId",
			expected: "TM01ARZ3NDEKTSV4RRFFQ69G5FAV",
		},
		{
			name:     "nil superTeamId returns empty",
			filter:   &model.LeaderboardFilter{},
			field:    "superTeamId",
			expected: "",
		},
		{
			name: "valid superTeamId",
			filter: &model.LeaderboardFilter{
				SuperTeamID: stringPtr("ST01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			},
			field:    "superTeamId",
			expected: "ST01ARZ3NDEKTSV4RRFFQ69G5FAV",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFilterString(tt.filter, tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}
