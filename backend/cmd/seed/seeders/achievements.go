package seeders

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SeedAchievements creates all types of achievements based on configuration
func (s *Seeder) SeedAchievements(stats *Stats) error {
	projectCount := 0
	for _, projectID := range s.Data.ProjectIDs {
		projectCount++

		// Split achievements across types
		// 60% simple, 35% content, 5% streak
		numSimple := int(float64(s.Config.NumAchievements) * 0.6)
		numContent := int(float64(s.Config.NumAchievements) * 0.35)
		// Remaining will be streak achievements

		if err := s.seedSimpleAchievements(projectID, numSimple, stats); err != nil {
			return err
		}

		if err := s.seedContentAchievements(projectID, numContent, stats); err != nil {
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
		INSERT INTO achievements (id, achievement_type, project_id, event_id, challenge_id, name, description_pending, description_completed, image_pending, image_completed, points)
		VALUES ($1, 'SIMPLE', $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for i := 0; i < count; i++ {
		achievementID := ulid.NewAchievementID()
		name := s.Fake.Lorem().Word() + " Achievement"
		descriptionPending := s.Fake.Lorem().Sentence(10)
		descriptionCompleted := "You earned the " + name + "!"
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
			descriptionPending,
			descriptionCompleted,
			imageURL,
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

func (s *Seeder) seedContentAchievements(projectID string, count int, stats *Stats) error {
	achievementQuery := `
		INSERT INTO achievements (id, achievement_type, project_id, event_id, challenge_id, name, description_pending, description_completed, image_pending, image_completed, points)
		VALUES ($1, 'CONTENT', $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	contentQuery := `
		INSERT INTO content_achievements (achievement_id)
		VALUES ($1)
	`

	// Create external content first, then link as content item
	externalContentQuery := `
		INSERT INTO external_content (id, plan_id, task_id, content_id, content_type, published_at, synced_at, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	externalContentTranslationQuery := `
		INSERT INTO external_content_translations (external_content_id, language_code, title)
		VALUES ($1, $2, $3)
	`

	contentItemQuery := `
		INSERT INTO content_achievement_items (id, achievement_id, external_content_id, sort_order)
		VALUES ($1, $2, $3, $4)
	`

	contentTypes := []string{"periodical_article", "media_episode"}

	for i := 0; i < count; i++ {
		achievementID := ulid.NewAchievementID()
		name := "Complete: " + s.Fake.Lorem().Word()
		descriptionPending := "Complete all content items to earn this achievement."
		descriptionCompleted := "You completed all content items!"
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
			descriptionPending,
			descriptionCompleted,
			imageURL,
			imageURL,
			points,
		)
		if err != nil {
			return err
		}

		_, err = s.DB.Pool.Exec(s.Ctx, contentQuery, achievementID)
		if err != nil {
			return err
		}

		// Add 3-8 content items (mix of articles and tracks)
		numItems := 3 + rand.Intn(6)
		planID := fmt.Sprintf("seed-content-plan-%s-%d", achievementID, i)

		for j := 0; j < numItems; j++ {
			// Create external content
			externalContentID := ulid.NewExternalContentID()
			taskID := fmt.Sprintf("task-%d", j+1)
			contentType := contentTypes[rand.Intn(len(contentTypes))]
			contentID := fmt.Sprintf("%s-%d", contentType, rand.Intn(10000))
			title := s.Fake.Lorem().Sentence(5)
			now := time.Now()

			_, err = s.DB.Pool.Exec(s.Ctx, externalContentQuery,
				externalContentID,
				planID,
				taskID,
				contentID,
				contentType,
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

			// Create content item linked to external content
			contentItemID := ulid.NewContentItemID()
			_, err = s.DB.Pool.Exec(s.Ctx, contentItemQuery,
				contentItemID,
				achievementID,
				externalContentID,
				j,
			)
			if err != nil {
				return fmt.Errorf("failed to create content achievement item: %w", err)
			}
		}

		s.Data.AchievementIDs[projectID] = append(s.Data.AchievementIDs[projectID], achievementID)
		stats.Achievements++
	}

	return nil
}

func (s *Seeder) seedStreakAchievements(projectID string, stats *Stats) error {
	achievementQuery := `
		INSERT INTO achievements (id, achievement_type, project_id, event_id, challenge_id, name, description_pending, description_completed, image_pending, image_completed, points)
		VALUES ($1, 'STREAK', $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	junctionQuery := `
		INSERT INTO streak_achievements (achievement_id)
		VALUES ($1)
	`

	// Create one streak achievement per project
	achievementID := ulid.NewAchievementID()
	name := "Streak Master"
	descriptionPending := "Complete all content before the deadlines."
	descriptionCompleted := "You completed all content on time!"
	imageURL := fmt.Sprintf("https://placecats.com/bella/%d/%d", 300+rand.Intn(100), 300+rand.Intn(100))
	points := 100 + rand.Intn(200)

	_, err := s.DB.Pool.Exec(s.Ctx, achievementQuery,
		achievementID,
		projectID,
		nil,
		nil,
		name,
		descriptionPending,
		descriptionCompleted,
		imageURL,
		imageURL,
		points,
	)
	if err != nil {
		return err
	}

	_, err = s.DB.Pool.Exec(s.Ctx, junctionQuery, achievementID)
	if err != nil {
		return err
	}

	s.Data.AchievementIDs[projectID] = append(s.Data.AchievementIDs[projectID], achievementID)
	stats.Achievements++

	return nil
}
