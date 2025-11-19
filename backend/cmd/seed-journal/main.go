package main

import (
	"context"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/bcc-media/wayfarer/cmd/seed/seeders"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/jaswdr/faker"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize structured logger
	lgr := logger.New(cfg.Server.Environment, logger.ParseLevel(cfg.Log.Level))
	slog.SetDefault(lgr)

	// Set up random seed
	seed := time.Now().UnixNano()
	fake := faker.NewWithSeed(rand.NewSource(seed))

	slog.Info("Starting score_journal seeding only", "seed", seed)

	// Connect to database
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Query existing data from database
	seededData := seeders.NewSeededData()

	// Get all user IDs
	userRows, err := db.Pool.Query(ctx, "SELECT id FROM users ORDER BY id")
	if err != nil {
		slog.Error("Failed to query users", "error", err)
		os.Exit(1)
	}
	for userRows.Next() {
		var userID string
		if err := userRows.Scan(&userID); err != nil {
			slog.Error("Failed to scan user", "error", err)
			os.Exit(1)
		}
		seededData.UserIDs = append(seededData.UserIDs, userID)
	}
	userRows.Close()

	if len(seededData.UserIDs) == 0 {
		slog.Error("No users found in database. Please run the main seed script first.")
		os.Exit(1)
	}

	slog.Info("Loaded users", "count", len(seededData.UserIDs))

	// Get all project IDs
	projectRows, err := db.Pool.Query(ctx, "SELECT id FROM projects ORDER BY id")
	if err != nil {
		slog.Error("Failed to query projects", "error", err)
		os.Exit(1)
	}
	for projectRows.Next() {
		var projectID string
		if err := projectRows.Scan(&projectID); err != nil {
			slog.Error("Failed to scan project", "error", err)
			os.Exit(1)
		}
		seededData.ProjectIDs = append(seededData.ProjectIDs, projectID)
	}
	projectRows.Close()

	if len(seededData.ProjectIDs) == 0 {
		slog.Error("No projects found in database. Please run the main seed script first.")
		os.Exit(1)
	}

	slog.Info("Loaded projects", "count", len(seededData.ProjectIDs))

	// Get all event IDs grouped by project
	eventRows, err := db.Pool.Query(ctx, "SELECT id, project_id FROM events ORDER BY project_id, id")
	if err != nil {
		slog.Error("Failed to query events", "error", err)
		os.Exit(1)
	}
	for eventRows.Next() {
		var eventID, projectID string
		if err := eventRows.Scan(&eventID, &projectID); err != nil {
			slog.Error("Failed to scan event", "error", err)
			os.Exit(1)
		}
		seededData.EventIDs[projectID] = append(seededData.EventIDs[projectID], eventID)
	}
	eventRows.Close()

	slog.Info("Loaded events", "count", countEvents(seededData.EventIDs))

	// Initialize seeder context
	seeder := &seeders.Seeder{
		DB:   db,
		Fake: fake,
		Ctx:  ctx,
		Data: seededData,
		Config: seeders.SeedConfig{
			ProjectParticipationRate:  0.9, // Same as default
			AchievementCompletionRate: 0.7, // Same as default
		},
	}

	// Run score_journal seeder
	stats := &seeders.Stats{}
	slog.Info("Seeding score journal entries")
	if err := seeder.SeedScoreJournal(stats); err != nil {
		slog.Error("Failed to seed score journal", "error", err)
		os.Exit(1)
	}

	slog.Info("Score journal seeding completed successfully")
}

func countEvents(eventIDs map[string][]string) int {
	count := 0
	for _, events := range eventIDs {
		count += len(events)
	}
	return count
}
