package seeders

import (
	"log/slog"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
)

// SeedProgress creates user progress data based on configured completion rates
func (s *Seeder) SeedProgress(stats *Stats) error {
	slog.Info("Starting progress seeding with configured rates",
		"project_participation", s.Config.ProjectParticipationRate,
		"achievement_completion", s.Config.AchievementCompletionRate)

	// Award achievements to users based on completion rate
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
	// For each project, award achievements to users who are in that project
	// based on the achievement completion rate
	batchSize := 1000
	totalAwarded := 0

	for projectIdx, projectID := range s.Data.ProjectIDs {
		if len(s.Data.AchievementIDs[projectID]) == 0 {
			continue
		}

		batchRows := [][]interface{}{}

		// For each user, they have ProjectParticipationRate chance of being in this project
		// For each achievement in the project, they have AchievementCompletionRate chance of having it
		for _, userID := range s.Data.UserIDs {
			// Check if user would be in this project (based on participation rate)
			if rand.Float64() >= s.Config.ProjectParticipationRate {
				continue
			}

			// For each achievement in this project
			for _, achievementID := range s.Data.AchievementIDs[projectID] {
				if rand.Float64() < s.Config.AchievementCompletionRate {
					achievedAt := time.Now().AddDate(0, 0, -rand.Intn(30))
					batchRows = append(batchRows, []interface{}{userID, achievementID, achievedAt})

					// Batch insert when we hit batch size
					if len(batchRows) >= batchSize {
						_, err := s.DB.Pool.CopyFrom(
							s.Ctx,
							pgx.Identifier{"user_achievements"},
							[]string{"user_id", "achievement_id", "achieved_at"},
							pgx.CopyFromRows(batchRows),
						)
						if err != nil {
							return err
						}
						totalAwarded += len(batchRows)
						batchRows = [][]interface{}{}
					}
				}
			}
		}

		// Insert remaining rows for this project
		if len(batchRows) > 0 {
			_, err := s.DB.Pool.CopyFrom(
				s.Ctx,
				pgx.Identifier{"user_achievements"},
				[]string{"user_id", "achievement_id", "achieved_at"},
				pgx.CopyFromRows(batchRows),
			)
			if err != nil {
				return err
			}
			totalAwarded += len(batchRows)
		}

		if (projectIdx+1)%10 == 0 {
			slog.Info("User achievements progress", "projects_completed", projectIdx+1, "total_projects", len(s.Data.ProjectIDs), "achievements_awarded", totalAwarded)
		}
	}

	slog.Info("User achievements completed", "total_awarded", totalAwarded)
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
