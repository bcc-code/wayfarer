package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	// Parse command line flags
	direction := flag.String("direction", "up", "Migration direction: up, down, status")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		panic(err)
	}

	// Initialize structured logger
	lgr := logger.New(cfg.Server.Environment, logger.ParseLevel(cfg.Log.Level))
	slog.SetDefault(lgr)

	slog.Info("starting migration",
		"direction", *direction,
		"database", maskConnectionString(cfg.Database.URL),
	)

	// Connect to database using stdlib driver for goose compatibility
	db, err := sql.Open("pgx", cfg.Database.URL)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		panic(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()

	// Verify connection
	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping database", "error", err)
		panic(err)
	}

	// Run migrations
	if err := runMigrations(db, *direction); err != nil {
		slog.Error("Migration failed", "error", err)
		panic(err)
	}

	slog.Info("migration completed successfully")
}

func runMigrations(db *sql.DB, direction string) error {
	// Set up goose with embedded migrations from database package
	goose.SetBaseFS(database.Migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Run the migration
	switch direction {
	case "up":
		if err := goose.Up(db, "migrations"); err != nil {
			return fmt.Errorf("migration up failed: %w", err)
		}
	case "down":
		if err := goose.Down(db, "migrations"); err != nil {
			return fmt.Errorf("migration down failed: %w", err)
		}
	case "status":
		if err := goose.Status(db, "migrations"); err != nil {
			return fmt.Errorf("migration status failed: %w", err)
		}
	default:
		return fmt.Errorf("unknown direction: %s (use: up, down, status)", direction)
	}

	return nil
}

// maskConnectionString masks sensitive parts of the connection string for logging
func maskConnectionString(connStr string) string {
	// Simple masking - in production you might want more sophisticated masking
	if len(connStr) > 20 {
		return connStr[:10] + "..." + connStr[len(connStr)-10:]
	}
	return "***"
}
