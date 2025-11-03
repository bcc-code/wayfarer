package seeders

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SeedChallenges creates 15-25 challenges per project
func (s *Seeder) SeedChallenges(stats *Stats) error {
	query := `
		INSERT INTO challenges (id, project_id, event_id, name, description, image_url, url, button_text, published_at, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for _, projectID := range s.Data.ProjectIDs {
		numChallenges := 15 + rand.Intn(11)

		for i := 0; i < numChallenges; i++ {
			challengeID := ulid.NewChallengeID()
			name := s.Fake.Lorem().Word() + " Challenge"
			description := s.Fake.Lorem().Sentence(12)
			imageURL := fmt.Sprintf("https://placecats.com/%d/%d", 400+rand.Intn(100), 300+rand.Intn(100))
			url := s.Fake.Internet().URL()
			buttonText := "Complete Challenge"

			// 30% chance to be linked to an event
			var eventID *string
			if rand.Float32() < 0.3 && len(s.Data.EventIDs[projectID]) > 0 {
				eID := s.Data.EventIDs[projectID][rand.Intn(len(s.Data.EventIDs[projectID]))]
				eventID = &eID
			}

			// 40% chance to have an end time
			var endTime *time.Time
			if rand.Float32() < 0.4 {
				et := time.Now().AddDate(0, 0, 7+rand.Intn(60))
				endTime = &et
			}

			// 80% chance to be published
			var publishedAt *time.Time
			if rand.Float32() < 0.8 {
				pt := time.Now().AddDate(0, 0, -rand.Intn(30))
				publishedAt = &pt
			}

			_, err := s.DB.Pool.Exec(s.Ctx, query,
				challengeID,
				projectID,
				eventID,
				name,
				description,
				imageURL,
				url,
				buttonText,
				publishedAt,
				endTime,
			)
			if err != nil {
				return err
			}

			s.Data.ChallengeIDs[projectID] = append(s.Data.ChallengeIDs[projectID], challengeID)
			stats.Challenges++
		}
	}

	return nil
}
