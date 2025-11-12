package seeders

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SeedProjects creates projects with events and streaks
func (s *Seeder) SeedProjects(stats *Stats) error {
	projectNames := []string{
		"Summer Bible Camp", "Youth Winter Retreat", "Spring Revival",
		"Fall Conference", "Easter Celebration", "Christmas Outreach",
		"Missions Week", "Worship Night", "Prayer Summit",
		"Leadership Training", "Discipleship Course", "Baptism Service",
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

	for i := 0; i < s.Config.NumProjects; i++ {
		projectID := ulid.NewProjectID()

		// Random project timing
		startOffset := rand.Intn(365) - 180  // -180 to +185 days
		duration := rand.Intn(120) + 30      // 30-150 days duration
		startDate := time.Now().AddDate(0, 0, startOffset)
		endDate := startDate.AddDate(0, 0, duration)

		// Use project names cyclically, add year
		baseName := projectNames[i%len(projectNames)]
		year := time.Now().Year() + (startOffset / 365)
		name := fmt.Sprintf("%s %d", baseName, year)
		description := s.Fake.Lorem().Sentence(10)

		// 10% chance to be archived (if end date is in the past)
		archived := endDate.Before(time.Now()) && rand.Float64() < 0.1

		// Generate branding data
		scheme := colorSchemes[i%len(colorSchemes)]
		firstLetter := string(baseName[0])
		logoURL := "https://placehold.co/100x100/" + scheme.bg + "/" + scheme.text + "?text=" + firstLetter
		colorPrimary := s.Fake.Color().Hex()
		colorSecondary := s.Fake.Color().Hex()
		colorTertiary := s.Fake.Color().Hex()

		_, err := s.DB.Pool.Exec(s.Ctx, projectQuery,
			projectID,
			name,
			description,
			startDate,
			endDate,
			logoURL,
			colorPrimary,
			colorSecondary,
			colorTertiary,
			archived,
		)
		if err != nil {
			return err
		}

		s.Data.ProjectIDs = append(s.Data.ProjectIDs, projectID)
		stats.Projects++

		// Create 3-5 events per project
		numEvents := 3 + rand.Intn(3)
		for j := 0; j < numEvents; j++ {
			eventID := ulid.NewEventID()
			eventStart := startDate.AddDate(0, 0, j*7)
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
		for j := 0; j < numStreaks; j++ {
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

		if (i+1)%10 == 0 {
			slog.Info("Projects progress", "created", i+1, "total", s.Config.NumProjects)
		}
	}

	return nil
}
