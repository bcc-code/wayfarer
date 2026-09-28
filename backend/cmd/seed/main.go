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
	numUsers := flag.Int("users", 75, "Number of users to generate")
	numProjects := flag.Int("projects", 3, "Number of projects to generate")
	numChurches := flag.Int("churches", 8, "Number of churches to generate")
	numSuperTeams := flag.Int("superteams", 5, "Number of superteams per project")
	numAchievements := flag.Int("achievements", 50, "Number of achievements per project")
	teamSize := flag.Int("team-size", 8, "Average team size")
	projectParticipationRate := flag.Float64("project-participation", 0.9, "Probability a user joins a project (0.0-1.0)")
	achievementCompletionRate := flag.Float64("achievement-completion", 0.7, "Probability a user completes an achievement (0.0-1.0)")
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

	// Initialize seeder context with configuration
	seeder := &seeders.Seeder{
		DB:   db,
		Fake: fake,
		Ctx:  ctx,
		Data: seeders.NewSeededData(),
		Config: seeders.SeedConfig{
			NumUsers:                  *numUsers,
			NumProjects:               *numProjects,
			NumChurches:               *numChurches,
			NumSuperTeams:             *numSuperTeams,
			NumAchievements:           *numAchievements,
			TeamSize:                  *teamSize,
			ProjectParticipationRate:  *projectParticipationRate,
			AchievementCompletionRate: *achievementCompletionRate,
		},
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

	slog.Info("Seeding score journal")
	if err := seeder.SeedScoreJournal(stats); err != nil {
		slog.Error("Failed to seed score journal", "error", err)
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
		"challenges", stats.Challenges,
		"achievements", stats.Achievements,
	)
}
