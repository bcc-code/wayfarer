package seeders

import (
	"math/rand"
	"time"

	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SeedProjects creates 3 projects with events and streaks
func (s *Seeder) SeedProjects(stats *Stats) error {
	projects := []struct {
		name        string
		description string
		startOffset int // days from now
		endOffset   int // days from now
		archived    bool
	}{
		{
			name:        "Summer Bible Camp 2025",
			description: "Join us for an amazing summer adventure exploring God's word!",
			startOffset: -30,
			endOffset:   60,
			archived:    false,
		},
		{
			name:        "Youth Winter Retreat 2024",
			description: "A week of worship, fellowship, and spiritual growth in the mountains.",
			startOffset: -120,
			endOffset:   -30,
			archived:    false,
		},
		{
			name:        "Spring Revival 2024",
			description: "Experience renewal and revival this spring season.",
			startOffset: -200,
			endOffset:   -150,
			archived:    true,
		},
	}

	projectQuery := `
		INSERT INTO projects (id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, archived)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	eventQuery := `
		INSERT INTO events (id, project_id, name, description, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	streakQuery := `
		INSERT INTO streaks (id, project_id, name, description)
		VALUES ($1, $2, $3, $4)
	`

	for idx, proj := range projects {
		projectID := ulid.NewProjectID()
		startDate := time.Now().AddDate(0, 0, proj.startOffset)
		endDate := time.Now().AddDate(0, 0, proj.endOffset)

		// Generate branding data
		// Color schemes: background colors with white text for good contrast
		colorSchemes := []struct{ bg, text string }{
			{"E53E3E", "FFFFFF"}, // Red background, white text
			{"3182CE", "FFFFFF"}, // Blue background, white text
			{"38A169", "FFFFFF"}, // Green background, white text
			{"805AD5", "FFFFFF"}, // Purple background, white text
			{"DD6B20", "FFFFFF"}, // Orange background, white text
			{"319795", "FFFFFF"}, // Teal background, white text
			{"D53F8C", "FFFFFF"}, // Pink background, white text
		}
		scheme := colorSchemes[idx%len(colorSchemes)]
		firstLetter := string(proj.name[0])

		logoURL := "https://placehold.co/100x100/" + scheme.bg + "/" + scheme.text + "?text=" + firstLetter
		colorPrimary := s.Fake.Color().Hex()
		colorSecondary := s.Fake.Color().Hex()
		colorTertiary := s.Fake.Color().Hex()

		_, err := s.DB.Pool.Exec(s.Ctx, projectQuery,
			projectID,
			proj.name,
			proj.description,
			startDate,
			endDate,
			logoURL,
			colorPrimary,
			colorSecondary,
			colorTertiary,
			proj.archived,
		)
		if err != nil {
			return err
		}

		s.Data.ProjectIDs = append(s.Data.ProjectIDs, projectID)
		stats.Projects++

		// Create 3-5 events per project
		numEvents := 3 + rand.Intn(3)
		for i := 0; i < numEvents; i++ {
			eventID := ulid.NewEventID()
			eventStart := startDate.AddDate(0, 0, i*7)
			eventEnd := eventStart.AddDate(0, 0, 3)

			eventName := s.Fake.Lorem().Word() + " Event"
			eventDesc := s.Fake.Lorem().Sentence(10)

			_, err := s.DB.Pool.Exec(s.Ctx, eventQuery,
				eventID,
				projectID,
				eventName,
				eventDesc,
				eventStart,
				eventEnd,
			)
			if err != nil {
				return err
			}

			s.Data.EventIDs[projectID] = append(s.Data.EventIDs[projectID], eventID)
			stats.Events++
		}

		// Create 1-2 streaks per project
		numStreaks := 1 + rand.Intn(2)
		for i := 0; i < numStreaks; i++ {
			streakID := ulid.NewStreakID()
			streakName := s.Fake.Lorem().Word() + " Streak"
			streakDesc := "Maintain your streak by completing activities daily!"

			_, err := s.DB.Pool.Exec(s.Ctx, streakQuery,
				streakID,
				projectID,
				streakName,
				streakDesc,
			)
			if err != nil {
				return err
			}

			s.Data.StreakIDs[projectID] = append(s.Data.StreakIDs[projectID], streakID)
			stats.Streaks++
		}
	}

	return nil
}
