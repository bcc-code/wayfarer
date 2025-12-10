package seeders

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SeedAchievements creates all 4 types of achievements based on configuration
func (s *Seeder) SeedAchievements(stats *Stats) error {
	projectCount := 0
	for _, projectID := range s.Data.ProjectIDs {
		projectCount++

		// Split achievements across types
		// 60% simple, 20% reading, 15% listening, 5% streak
		numSimple := int(float64(s.Config.NumAchievements) * 0.6)
		numReading := int(float64(s.Config.NumAchievements) * 0.2)
		numListening := int(float64(s.Config.NumAchievements) * 0.15)
		// Remaining will be streak achievements

		if err := s.seedSimpleAchievements(projectID, numSimple, stats); err != nil {
			return err
		}

		if err := s.seedReadingAchievements(projectID, numReading, stats); err != nil {
			return err
		}

		if err := s.seedListeningAchievements(projectID, numListening, stats); err != nil {
			return err
		}

		// Create streak achievements (one per streak, up to remaining count)
		if err := s.seedStreakAchievements(projectID, stats); err != nil {
			return err
		}

		if projectCount%10 == 0 {
			slog.Info("Achievements progress", "projects_completed", projectCount, "total_projects", len(s.Data.ProjectIDs), "achievements_created", stats.Achievements)
		}
	}

	return nil
}

func (s *Seeder) seedSimpleAchievements(projectID string, count int, stats *Stats) error {
	query := `
		INSERT INTO achievements (id, achievement_type, project_id, event_id, challenge_id, name, description, image_url, points)
		VALUES ($1, 'SIMPLE', $2, $3, $4, $5, $6, $7, $8)
	`

	for i := 0; i < count; i++ {
		achievementID := ulid.NewAchievementID()
		name := s.Fake.Lorem().Word() + " Achievement"
		description := s.Fake.Lorem().Sentence(10)
		imageURL := fmt.Sprintf("https://placecats.com/%d/%d", 300+rand.Intn(100), 300+rand.Intn(100))
		points := (1 + rand.Intn(20)) * 5

		// 40% chance to be linked to an event
		var eventID *string
		if rand.Float32() < 0.4 && len(s.Data.EventIDs[projectID]) > 0 {
			eID := s.Data.EventIDs[projectID][rand.Intn(len(s.Data.EventIDs[projectID]))]
			eventID = &eID
		}

		// 50% chance to be linked to a challenge
		var challengeID *string
		if rand.Float32() < 0.5 && len(s.Data.ChallengeIDs[projectID]) > 0 {
			cID := s.Data.ChallengeIDs[projectID][rand.Intn(len(s.Data.ChallengeIDs[projectID]))]
			challengeID = &cID
		}

		_, err := s.DB.Pool.Exec(s.Ctx, query,
			achievementID,
			projectID,
			eventID,
			challengeID,
			name,
			description,
			imageURL,
			points,
		)
		if err != nil {
			return err
		}

		s.Data.AchievementIDs[projectID] = append(s.Data.AchievementIDs[projectID], achievementID)
		stats.Achievements++
	}

	return nil
}

func (s *Seeder) seedReadingAchievements(projectID string, count int, stats *Stats) error {
	achievementQuery := `
		INSERT INTO achievements (id, achievement_type, project_id, event_id, challenge_id, name, description, image_url, points)
		VALUES ($1, 'READING', $2, $3, $4, $5, $6, $7, $8)
	`

	readingQuery := `
		INSERT INTO reading_achievements (achievement_id)
		VALUES ($1)
	`

	// Create external content first, then link to article
	externalContentQuery := `
		INSERT INTO external_content (id, plan_id, task_id, content_id, content_type, published_at, synced_at, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	externalContentTranslationQuery := `
		INSERT INTO external_content_translations (external_content_id, language_code, title)
		VALUES ($1, $2, $3)
	`

	articleQuery := `
		INSERT INTO reading_achievement_articles (id, achievement_id, external_content_id)
		VALUES ($1, $2, $3)
	`

	for i := 0; i < count; i++ {
		achievementID := ulid.NewAchievementID()
		name := "Read: " + s.Fake.Lorem().Word()
		description := "Complete all articles to earn this achievement."
		imageURL := fmt.Sprintf("https://placecats.com/g/%d/%d", 300+rand.Intn(100), 300+rand.Intn(100))
		points := 50 + rand.Intn(100)

		var eventID *string
		if rand.Float32() < 0.3 && len(s.Data.EventIDs[projectID]) > 0 {
			eID := s.Data.EventIDs[projectID][rand.Intn(len(s.Data.EventIDs[projectID]))]
			eventID = &eID
		}

		_, err := s.DB.Pool.Exec(s.Ctx, achievementQuery,
			achievementID,
			projectID,
			eventID,
			nil,
			name,
			description,
			imageURL,
			points,
		)
		if err != nil {
			return err
		}

		_, err = s.DB.Pool.Exec(s.Ctx, readingQuery, achievementID)
		if err != nil {
			return err
		}

		// Add 3-6 articles via external content
		numArticles := 3 + rand.Intn(4)
		planID := fmt.Sprintf("seed-reading-plan-%s-%d", achievementID, i)

		for j := 0; j < numArticles; j++ {
			// Create external content
			externalContentID := ulid.NewExternalContentID()
			taskID := fmt.Sprintf("task-%d", j+1)
			contentID := fmt.Sprintf("article-%d", rand.Intn(10000))
			title := s.Fake.Lorem().Sentence(5)
			now := time.Now()

			_, err = s.DB.Pool.Exec(s.Ctx, externalContentQuery,
				externalContentID,
				planID,
				taskID,
				contentID,
				"periodical_article",
				now,
				now,
				"seed",
			)
			if err != nil {
				return fmt.Errorf("failed to create external content: %w", err)
			}

			// Create translation for the external content
			_, err = s.DB.Pool.Exec(s.Ctx, externalContentTranslationQuery,
				externalContentID,
				"en",
				title,
			)
			if err != nil {
				return fmt.Errorf("failed to create external content translation: %w", err)
			}

			// Create article linked to external content
			articleID := ulid.NewReadingAchievementID()
			_, err = s.DB.Pool.Exec(s.Ctx, articleQuery,
				articleID,
				achievementID,
				externalContentID,
			)
			if err != nil {
				return fmt.Errorf("failed to create reading achievement article: %w", err)
			}
		}

		s.Data.AchievementIDs[projectID] = append(s.Data.AchievementIDs[projectID], achievementID)
		stats.Achievements++
	}

	return nil
}

func (s *Seeder) seedListeningAchievements(projectID string, count int, stats *Stats) error {
	achievementQuery := `
		INSERT INTO achievements (id, achievement_type, project_id, event_id, challenge_id, name, description, image_url, points)
		VALUES ($1, 'LISTENING', $2, $3, $4, $5, $6, $7, $8)
	`

	listeningQuery := `
		INSERT INTO listening_achievements (achievement_id)
		VALUES ($1)
	`

	// Create external content first, then link to track
	externalContentQuery := `
		INSERT INTO external_content (id, plan_id, task_id, content_id, content_type, published_at, synced_at, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	externalContentTranslationQuery := `
		INSERT INTO external_content_translations (external_content_id, language_code, title)
		VALUES ($1, $2, $3)
	`

	trackQuery := `
		INSERT INTO listening_achievement_tracks (id, achievement_id, external_content_id)
		VALUES ($1, $2, $3)
	`

	for i := 0; i < count; i++ {
		achievementID := ulid.NewAchievementID()
		name := "Listen: " + s.Fake.Lorem().Word()
		description := "Listen to all tracks to earn this achievement."
		imageURL := fmt.Sprintf("https://placecats.com/neo/%d/%d", 300+rand.Intn(100), 300+rand.Intn(100))
		points := 50 + rand.Intn(100)

		var eventID *string
		if rand.Float32() < 0.3 && len(s.Data.EventIDs[projectID]) > 0 {
			eID := s.Data.EventIDs[projectID][rand.Intn(len(s.Data.EventIDs[projectID]))]
			eventID = &eID
		}

		_, err := s.DB.Pool.Exec(s.Ctx, achievementQuery,
			achievementID,
			projectID,
			eventID,
			nil,
			name,
			description,
			imageURL,
			points,
		)
		if err != nil {
			return err
		}

		_, err = s.DB.Pool.Exec(s.Ctx, listeningQuery, achievementID)
		if err != nil {
			return err
		}

		// Add 4-8 tracks via external content
		numTracks := 4 + rand.Intn(5)
		planID := fmt.Sprintf("seed-listening-plan-%s-%d", achievementID, i)

		for j := 0; j < numTracks; j++ {
			// Create external content
			externalContentID := ulid.NewExternalContentID()
			taskID := fmt.Sprintf("task-%d", j+1)
			contentID := fmt.Sprintf("track-%d", rand.Intn(10000))
			title := s.Fake.Lorem().Sentence(3)
			now := time.Now()

			_, err = s.DB.Pool.Exec(s.Ctx, externalContentQuery,
				externalContentID,
				planID,
				taskID,
				contentID,
				"media_episode",
				now,
				now,
				"seed",
			)
			if err != nil {
				return fmt.Errorf("failed to create external content: %w", err)
			}

			// Create translation for the external content
			_, err = s.DB.Pool.Exec(s.Ctx, externalContentTranslationQuery,
				externalContentID,
				"en",
				title,
			)
			if err != nil {
				return fmt.Errorf("failed to create external content translation: %w", err)
			}

			// Create track linked to external content
			trackID := ulid.NewListeningAchievementID()
			_, err = s.DB.Pool.Exec(s.Ctx, trackQuery,
				trackID,
				achievementID,
				externalContentID,
			)
			if err != nil {
				return fmt.Errorf("failed to create listening achievement track: %w", err)
			}
		}

		s.Data.AchievementIDs[projectID] = append(s.Data.AchievementIDs[projectID], achievementID)
		stats.Achievements++
	}

	return nil
}

func (s *Seeder) seedStreakAchievements(projectID string, stats *Stats) error {
	achievementQuery := `
		INSERT INTO achievements (id, achievement_type, project_id, event_id, challenge_id, name, description, image_url, points)
		VALUES ($1, 'STREAK', $2, $3, $4, $5, $6, $7, $8)
	`

	streakQuery := `
		INSERT INTO streak_achievements (achievement_id, streak_id, needed_streak)
		VALUES ($1, $2, $3)
	`

	// Create one streak achievement per streak
	for _, streakID := range s.Data.StreakIDs[projectID] {
		achievementID := ulid.NewAchievementID()
		name := "Streak Master"
		description := "Maintain your streak to earn this achievement."
		imageURL := fmt.Sprintf("https://placecats.com/bella/%d/%d", 300+rand.Intn(100), 300+rand.Intn(100))
		points := 100 + rand.Intn(200)

		_, err := s.DB.Pool.Exec(s.Ctx, achievementQuery,
			achievementID,
			projectID,
			nil,
			nil,
			name,
			description,
			imageURL,
			points,
		)
		if err != nil {
			return err
		}

		requiredDays := 7 + rand.Intn(22) // 7-28 days
		_, err = s.DB.Pool.Exec(s.Ctx, streakQuery,
			achievementID,
			streakID,
			requiredDays,
		)
		if err != nil {
			return err
		}

		s.Data.AchievementIDs[projectID] = append(s.Data.AchievementIDs[projectID], achievementID)
		stats.Achievements++
	}

	return nil
}
