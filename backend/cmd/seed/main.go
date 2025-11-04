package main

import (
	"context"
	"flag"
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
	// Parse command line flags
	seedValue := flag.Int64("seed", 0, "Seed value for reproducible data (0 = random)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize structured logger
	lgr := logger.New(cfg.Server.Environment, logger.ParseLevel(cfg.Log.Level))
	slog.SetDefault(lgr)

	// Set up random seed
	var seed int64
	if *seedValue == 0 {
		seed = time.Now().UnixNano()
	} else {
		seed = *seedValue
	}
	fake := faker.NewWithSeed(rand.NewSource(seed))

	slog.Info("Starting database seed", "seed", seed)

	// Connect to database
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Clear existing data
	slog.Info("Clearing existing data")
	if err := clearDatabase(ctx, db); err != nil {
		slog.Error("Failed to clear database", "error", err)
		os.Exit(1)
	}

	// Initialize seeder context
	seeder := &seeders.Seeder{
		DB:   db,
		Fake: fake,
		Ctx:  ctx,
		Data: seeders.NewSeededData(),
	}

	// Run seeders in order
	stats := &seeders.Stats{}

	slog.Info("Seeding churches")
	if err := seeder.SeedChurches(stats); err != nil {
		slog.Error("Failed to seed churches", "error", err)
		os.Exit(1)
	}

	slog.Info("Seeding users")
	if err := seeder.SeedUsers(stats); err != nil {
		slog.Error("Failed to seed users", "error", err)
		os.Exit(1)
	}

	slog.Info("Seeding projects and events")
	if err := seeder.SeedProjects(stats); err != nil {
		slog.Error("Failed to seed projects", "error", err)
		os.Exit(1)
	}

	slog.Info("Seeding teams")
	if err := seeder.SeedTeams(stats); err != nil {
		slog.Error("Failed to seed teams", "error", err)
		os.Exit(1)
	}

	slog.Info("Seeding challenges")
	if err := seeder.SeedChallenges(stats); err != nil {
		slog.Error("Failed to seed challenges", "error", err)
		os.Exit(1)
	}

	slog.Info("Seeding achievements")
	if err := seeder.SeedAchievements(stats); err != nil {
		slog.Error("Failed to seed achievements", "error", err)
		os.Exit(1)
	}

	slog.Info("Seeding progress data")
	if err := seeder.SeedProgress(stats); err != nil {
		slog.Error("Failed to seed progress", "error", err)
		os.Exit(1)
	}

	// Print summary
	slog.Info("Seed completed successfully",
		"churches", stats.Churches,
		"users", stats.Users,
		"projects", stats.Projects,
		"events", stats.Events,
		"superteams", stats.SuperTeams,
		"teams", stats.Teams,
		"streaks", stats.Streaks,
		"challenges", stats.Challenges,
		"achievements", stats.Achievements,
	)
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
		_, err := db.Pool.Exec(ctx, "TRUNCATE TABLE "+table+" CASCADE")
		if err != nil {
			return err
		}
	}

	return nil
}
