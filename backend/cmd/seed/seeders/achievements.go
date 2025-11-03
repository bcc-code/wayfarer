package seeders

import (
	"fmt"
	"math/rand"

	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SeedAchievements creates all 4 types of achievements
func (s *Seeder) SeedAchievements(stats *Stats) error {
	for _, projectID := range s.Data.ProjectIDs {
		// Create 5-8 simple achievements
		if err := s.seedSimpleAchievements(projectID, 5+rand.Intn(4), stats); err != nil {
			return err
		}

		// Create 2-4 reading achievements
		if err := s.seedReadingAchievements(projectID, 2+rand.Intn(3), stats); err != nil {
			return err
		}

		// Create 2-4 listening achievements
		if err := s.seedListeningAchievements(projectID, 2+rand.Intn(3), stats); err != nil {
			return err
		}

		// Create 1-2 streak achievements (one per streak)
		if err := s.seedStreakAchievements(projectID, stats); err != nil {
			return err
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

	articleQuery := `
		INSERT INTO reading_achievement_articles (id, achievement_id, article_id, title, author, url)
		VALUES ($1, $2, $3, $4, $5, $6)
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

		// Add 3-6 articles
		numArticles := 3 + rand.Intn(4)
		for j := 0; j < numArticles; j++ {
			articleID := ulid.NewReadingAchievementID()
			externalArticleID := fmt.Sprintf("ART-%d", rand.Intn(10000))
			title := s.Fake.Lorem().Sentence(5)
			author := s.Fake.Person().Name()
			url := s.Fake.Internet().URL()

			_, err = s.DB.Pool.Exec(s.Ctx, articleQuery,
				articleID,
				achievementID,
				externalArticleID,
				title,
				author,
				url,
			)
			if err != nil {
				return err
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

	trackQuery := `
		INSERT INTO listening_achievement_tracks (id, achievement_id, track_id, name, description, image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
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

		// Add 4-8 tracks
		numTracks := 4 + rand.Intn(5)
		for j := 0; j < numTracks; j++ {
			trackID := ulid.NewListeningAchievementID()
			externalTrackID := fmt.Sprintf("TRK-%d", rand.Intn(10000))
			trackName := s.Fake.Lorem().Sentence(3)
			trackDesc := s.Fake.Lorem().Sentence(8)
			trackImageURL := fmt.Sprintf("https://placecats.com/millie/%d/%d", 300+rand.Intn(100), 300+rand.Intn(100))

			_, err = s.DB.Pool.Exec(s.Ctx, trackQuery,
				trackID,
				achievementID,
				externalTrackID,
				trackName,
				trackDesc,
				trackImageURL,
			)
			if err != nil {
				return err
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
