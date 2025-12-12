package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/bcc-media/wayfarer/internal/auth0"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/sony/gobreaker/v2"
)

func main() {
	ctx := context.Background()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	lgr := logger.New(cfg.Server.Environment, logger.ParseLevel(cfg.Log.Level))
	slog.SetDefault(lgr)

	slog.Info("Starting church sync")

	// Connect to database
	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Validate required configuration
	if cfg.Auth0.Domain == "" || cfg.Auth0.ClientID == "" || cfg.Members.Domain == "" {
		slog.Error("Missing required configuration",
			"auth0_domain", cfg.Auth0.Domain,
			"members_domain", cfg.Members.Domain,
		)
		os.Exit(1)
	}

	// Initialize Auth0 client
	auth0Client := auth0.New(auth0.Config{
		Domain:       cfg.Auth0.Domain,
		ClientID:     cfg.Auth0.ClientID,
		ClientSecret: cfg.Auth0.ClientSecret,
	})

	// Create circuit breaker for Members API
	membersBreaker := gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
		Name:    "members-api-sync",
		Timeout: 5 * time.Second,
	})

	// Initialize Members API client
	membersClient := members.New(
		members.Config{Domain: cfg.Members.Domain},
		auth0Client,
		membersBreaker,
	)

	// Execute sync
	if err := syncChurches(ctx, membersClient, db.Queries); err != nil {
		slog.Error("Church sync failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Church sync completed successfully")
}

func syncChurches(ctx context.Context, membersClient *members.Client, queries *sqlc.Queries) error {
	// Fetch all organizations from Members API
	slog.Info("Fetching organizations from Members API")
	orgs, err := membersClient.GetAllOrganizations(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch organizations: %w", err)
	}
	slog.Info("Fetched organizations", "count", len(orgs))

	// Warn if exactly 999 orgs (possible pagination issue)
	if len(orgs) == 999 {
		slog.Warn("Fetched exactly 999 organizations - may be hitting pagination limit")
	}

	// Process each organization
	var successCount, errorCount int
	for i, org := range orgs {
		if err := upsertChurch(ctx, queries, org); err != nil {
			slog.Error("Failed to upsert church",
				"external_id", org.OrgID,
				"name", org.Name,
				"error", err,
			)
			errorCount++
			continue
		}
		successCount++

		// Log progress every 100 churches
		if (i+1)%100 == 0 {
			slog.Info("Progress", "processed", i+1, "total", len(orgs))
		}
	}

	// Log final results
	slog.Info("Sync completed",
		"total", len(orgs),
		"success", successCount,
		"errors", errorCount,
	)

	if errorCount > 0 {
		return fmt.Errorf("sync completed with %d errors", errorCount)
	}

	return nil
}

func upsertChurch(ctx context.Context, queries *sqlc.Queries, org members.Organization) error {
	// Handle empty country code
	country := org.VisitingAddress.CountryCode
	if country == "" {
		country = "UNKNOWN"
	}

	// Generate new church ID (will only be used if insert happens)
	churchID := ulid.NewChurchID()

	// Convert OrgID to int32 pointer
	externalID := int32(org.OrgID)

	// Execute upsert
	_, err := queries.UpsertChurch(ctx, sqlc.UpsertChurchParams{
		ID:         churchID,
		ExternalID: &externalID,
		Name:       org.Name,
		Country:    country,
		Category:   "S",
	})

	return err
}
