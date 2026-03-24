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
		INSERT INTO projects (id, name, description, start_date, end_date, logo_url,
			color_light_accent, color_light_accent_contrast, color_light_on_accent,
			color_light_background_default, color_light_background_raised, color_light_background_indent,
			color_light_text_default, color_light_text_muted, color_light_text_hint,
			color_light_shadow_default, color_light_shadow_blank, color_light_border_default,
			color_dark_accent, color_dark_accent_contrast, color_dark_on_accent,
			color_dark_background_default, color_dark_background_raised, color_dark_background_indent,
			color_dark_text_default, color_dark_text_muted, color_dark_text_hint,
			color_dark_shadow_default, color_dark_shadow_blank, color_dark_border_default,
			archived)
		VALUES ($1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31)
	`

	eventQuery := `
		INSERT INTO events (id, project_id, name, description, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Color schemes: accent colors for light/dark modes
	colorSchemes := []struct{ accent string }{
		{"#E53E3E"}, // Red
		{"#3182CE"}, // Blue
		{"#38A169"}, // Green
		{"#805AD5"}, // Purple
		{"#DD6B20"}, // Orange
		{"#319795"}, // Teal
		{"#D53F8C"}, // Pink
	}

	// Default branding colors (from migration 24)
	defaultBranding := struct {
		lightAccentContrast    string
		lightOnAccent          string
		lightBackgroundDefault string
		lightBackgroundRaised  string
		lightBackgroundIndent  string
		lightTextDefault       string
		lightTextMuted         string
		lightTextHint          string
		lightShadowDefault     string
		lightShadowBlank       string
		lightBorderDefault     string
		darkAccentContrast     string
		darkOnAccent           string
		darkBackgroundDefault  string
		darkBackgroundRaised   string
		darkBackgroundIndent   string
		darkTextDefault        string
		darkTextMuted          string
		darkTextHint           string
		darkShadowDefault      string
		darkShadowBlank        string
		darkBorderDefault      string
	}{
		lightAccentContrast:    "#938636",
		lightOnAccent:          "#01121a",
		lightBackgroundDefault: "#f3ede5",
		lightBackgroundRaised:  "#ffffff",
		lightBackgroundIndent:  "rgb(99 56 1 / 0.05)",
		lightTextDefault:       "#282521",
		lightTextMuted:         "rgb(40 37 33 / 0.65)",
		lightTextHint:          "rgb(40 37 33 / 0.4)",
		lightShadowDefault:     "rgb(40 37 33 / 0.1)",
		lightShadowBlank:       "rgb(40 37 33 / 0)",
		lightBorderDefault:     "rgb(40 37 33 / 0.15)",
		darkAccentContrast:     "#e8dfa7",
		darkOnAccent:           "#1a1401",
		darkBackgroundDefault:  "#122026",
		darkBackgroundRaised:   "#0a3644",
		darkBackgroundIndent:   "rgb(0 9 13 / 0.25)",
		darkTextDefault:        "#f3ede5",
		darkTextMuted:          "rgb(243 237 229 / 0.7)",
		darkTextHint:           "rgb(243 237 229 / 0.4)",
		darkShadowDefault:      "rgb(18 32 38 / 0.3)",
		darkShadowBlank:        "rgb(18 32 38 / 0)",
		darkBorderDefault:      "rgb(156 214 243 / 0.09)",
	}

	for i := 0; i < s.Config.NumProjects; i++ {
		projectID := ulid.NewProjectID()

		// Random project timing
		startOffset := rand.Intn(365) - 180 // -180 to +185 days
		duration := rand.Intn(120) + 30     // 30-150 days duration
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
		logoURL := "https://placehold.co/100x100/" + scheme.accent[1:] + "/FFFFFF?text=" + firstLetter

		_, err := s.DB.Pool.Exec(s.Ctx, projectQuery,
			projectID,
			name,
			description,
			startDate,
			endDate,
			logoURL,
			// Light mode colors
			scheme.accent,
			defaultBranding.lightAccentContrast,
			defaultBranding.lightOnAccent,
			defaultBranding.lightBackgroundDefault,
			defaultBranding.lightBackgroundRaised,
			defaultBranding.lightBackgroundIndent,
			defaultBranding.lightTextDefault,
			defaultBranding.lightTextMuted,
			defaultBranding.lightTextHint,
			defaultBranding.lightShadowDefault,
			defaultBranding.lightShadowBlank,
			defaultBranding.lightBorderDefault,
			// Dark mode colors
			scheme.accent,
			defaultBranding.darkAccentContrast,
			defaultBranding.darkOnAccent,
			defaultBranding.darkBackgroundDefault,
			defaultBranding.darkBackgroundRaised,
			defaultBranding.darkBackgroundIndent,
			defaultBranding.darkTextDefault,
			defaultBranding.darkTextMuted,
			defaultBranding.darkTextHint,
			defaultBranding.darkShadowDefault,
			defaultBranding.darkShadowBlank,
			defaultBranding.darkBorderDefault,
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

		if (i+1)%10 == 0 {
			slog.Info("Projects progress", "created", i+1, "total", s.Config.NumProjects)
		}
	}

	return nil
}
