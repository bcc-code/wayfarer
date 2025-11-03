package seeders

import (
	"math/rand"
	"time"
)

// SeedProgress creates some user progress data (achievements and challenge completions)
func (s *Seeder) SeedProgress(stats *Stats) error {
	// Award some achievements to random users
	if err := s.seedUserAchievements(); err != nil {
		return err
	}

	// Complete some challenges for random users
	if err := s.seedChallengeCompletions(); err != nil {
		return err
	}

	return nil
}

func (s *Seeder) seedUserAchievements() error {
	query := `
		INSERT INTO user_achievements (user_id, achievement_id, achieved_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`

	// Award 20-30 random achievements
	numAwards := 20 + rand.Intn(11)

	for i := 0; i < numAwards; i++ {
		// Random user
		userID := s.Data.UserIDs[rand.Intn(len(s.Data.UserIDs))]

		// Random project
		projectID := s.Data.ProjectIDs[rand.Intn(len(s.Data.ProjectIDs))]

		// Random achievement from that project
		if len(s.Data.AchievementIDs[projectID]) == 0 {
			continue
		}
		achievementID := s.Data.AchievementIDs[projectID][rand.Intn(len(s.Data.AchievementIDs[projectID]))]

		achievedAt := time.Now().AddDate(0, 0, -rand.Intn(30))

		_, err := s.DB.Pool.Exec(s.Ctx, query, userID, achievementID, achievedAt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Seeder) seedChallengeCompletions() error {
	query := `
		INSERT INTO user_challenge_completions (user_id, challenge_id, completed_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`

	// Complete 30-50 random challenges
	numCompletions := 30 + rand.Intn(21)

	for i := 0; i < numCompletions; i++ {
		// Random user
		userID := s.Data.UserIDs[rand.Intn(len(s.Data.UserIDs))]

		// Random project
		projectID := s.Data.ProjectIDs[rand.Intn(len(s.Data.ProjectIDs))]

		// Random challenge from that project
		if len(s.Data.ChallengeIDs[projectID]) == 0 {
			continue
		}
		challengeID := s.Data.ChallengeIDs[projectID][rand.Intn(len(s.Data.ChallengeIDs[projectID]))]

		completedAt := time.Now().AddDate(0, 0, -rand.Intn(30))

		_, err := s.DB.Pool.Exec(s.Ctx, query, userID, challengeID, completedAt)
		if err != nil {
			return err
		}
	}

	return nil
}
