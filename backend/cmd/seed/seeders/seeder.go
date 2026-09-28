package seeders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/jaswdr/faker"
)

// SeedConfig holds configuration for the seeding process
type SeedConfig struct {
	NumUsers                  int
	NumProjects               int
	NumChurches               int
	NumSuperTeams             int
	NumAchievements           int
	TeamSize                  int
	ProjectParticipationRate  float64
	AchievementCompletionRate float64
}

// Seeder holds the database connection and faker instance
type Seeder struct {
	DB     *database.DB
	Fake   faker.Faker
	Ctx    context.Context
	Data   *SeededData
	Config SeedConfig
}

// Stats tracks how many records were seeded
type Stats struct {
	Churches     int
	Users        int
	Projects     int
	Events       int
	SuperTeams   int
	Teams        int
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
	ChallengeIDs   map[string][]string // projectID -> []challengeID
	AchievementIDs map[string][]string // projectID -> []achievementID
}

// NewSeededData creates a new SeededData instance
func NewSeededData() *SeededData {
	return &SeededData{
		EventIDs:       make(map[string][]string),
		SuperTeamIDs:   make(map[string][]string),
		TeamIDs:        make(map[string][]string),
		ChallengeIDs:   make(map[string][]string),
		AchievementIDs: make(map[string][]string),
	}
}
