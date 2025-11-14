package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/logger"
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

	slog.Info("Starting database clear")

	// Connect to database
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Clear existing data
	slog.Info("Clearing all database tables")
	if err := clearDatabase(ctx, db); err != nil {
		slog.Error("Failed to clear database", "error", err)
		os.Exit(1)
	}

	slog.Info("Database cleared successfully")
}

// clearDatabase truncates all tables in reverse dependency order
func clearDatabase(ctx context.Context, db *database.DB) error {
	tables := []string{
		"user_streak_activity",
		"user_listening_progress",
		"user_reading_progress",
		"user_challenge_completions",
		"super_team_achievements",
		"team_achievements",
		"user_achievements",
		"team_members",
		"user_events",
		"user_projects",
		"score_adjustments",
		"streak_achievements",
		"listening_achievement_tracks",
		"listening_achievements",
		"reading_achievement_articles",
		"reading_achievements",
		"achievements",
		"challenges",
		"streak_relevant_days",
		"streaks",
		"teams",
		"super_teams",
		"events",
		"projects",
		"users",
		"churches",
	}

	for _, table := range tables {
		slog.Info("Truncating table", "table", table)
		_, err := db.Pool.Exec(ctx, "TRUNCATE TABLE "+table+" CASCADE")
		if err != nil {
			return err
		}
	}

	return nil
}
