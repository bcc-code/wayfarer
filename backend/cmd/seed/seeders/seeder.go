package seeders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/jaswdr/faker"
)

// Seeder holds the database connection and faker instance
type Seeder struct {
	DB   *database.DB
	Fake faker.Faker
	Ctx  context.Context
	Data *SeededData
}

// Stats tracks how many records were seeded
type Stats struct {
	Churches     int
	Users        int
	Projects     int
	Events       int
	SuperTeams   int
	Teams        int
	Streaks      int
	Challenges   int
	Achievements int
}

// SeededData holds IDs of seeded entities for cross-referencing
type SeededData struct {
	ChurchIDs      []string
	UserIDs        []string
	ProjectIDs     []string
	EventIDs       map[string][]string // projectID -> []eventID
	SuperTeamIDs   map[string][]string // projectID -> []superTeamID
	TeamIDs        map[string][]string // projectID -> []teamID
	StreakIDs      map[string][]string // projectID -> []streakID
	ChallengeIDs   map[string][]string // projectID -> []challengeID
	AchievementIDs map[string][]string // projectID -> []achievementID
}

// NewSeededData creates a new SeededData instance
func NewSeededData() *SeededData {
	return &SeededData{
		EventIDs:       make(map[string][]string),
		SuperTeamIDs:   make(map[string][]string),
		TeamIDs:        make(map[string][]string),
		StreakIDs:      make(map[string][]string),
		ChallengeIDs:   make(map[string][]string),
		AchievementIDs: make(map[string][]string),
	}
}
