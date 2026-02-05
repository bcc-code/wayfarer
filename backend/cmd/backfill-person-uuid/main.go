package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/bcc-media/wayfarer/internal/auth0"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sony/gobreaker/v2"
)

const batchSize = 100

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

	slog.Info("Starting person_uuid backfill")

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

	// Initialize Auth0 client for Members API authentication
	auth0Client := auth0.New(auth0.Config{
		Domain:       cfg.Auth0.Domain,
		ClientID:     cfg.Auth0.ClientID,
		ClientSecret: cfg.Auth0.ClientSecret,
	})

	// Create circuit breaker for Members API
	membersBreaker := gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
		Name:    "members-api-backfill",
		Timeout: 10 * time.Second,
	})

	// Initialize Members API client
	membersClient := members.New(
		members.Config{Domain: cfg.Members.Domain},
		auth0Client,
		membersBreaker,
	)

	// Execute backfill
	if err := backfillPersonUUID(ctx, membersClient, db.Queries); err != nil {
		slog.Error("Backfill failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Backfill completed successfully")
}

func backfillPersonUUID(ctx context.Context, membersClient *members.Client, queries *sqlc.Queries) error {
	var totalProcessed, successCount, errorCount, skippedCount int

	for {
		// Fetch batch of users without person_uuid (query only returns users with numeric members_id)
		users, err := queries.GetUsersWithoutPersonUUID(ctx, batchSize)
		if err != nil {
			return err
		}

		if len(users) == 0 {
			slog.Info("No more users to process")
			break
		}

		slog.Info("Processing batch", "count", len(users))

		for _, user := range users {
			totalProcessed++

			// Parse members_id as int (should always succeed since query filters for numeric IDs)
			personID, err := strconv.Atoi(user.MembersID)
			if err != nil {
				slog.Warn("Invalid members_id format, skipping",
					"user_id", user.ID,
					"members_id", user.MembersID,
					"error", err,
				)
				skippedCount++
				continue
			}

			// Fetch member from Members API
			member, err := membersClient.Lookup(ctx, personID)
			if err != nil {
				slog.Warn("Failed to fetch member from API, skipping",
					"user_id", user.ID,
					"person_id", personID,
					"error", err,
				)
				errorCount++
				continue
			}

			// Check if member has a valid UUID
			if member.Uid == uuid.Nil {
				slog.Warn("Member has no UUID, skipping",
					"user_id", user.ID,
					"person_id", personID,
				)
				skippedCount++
				continue
			}

			// Update user with person_uuid
			personUUID := pgtype.UUID{Bytes: member.Uid, Valid: true}
			err = queries.UpdateUserPersonUUID(ctx, sqlc.UpdateUserPersonUUIDParams{
				ID:         user.ID,
				PersonUuid: personUUID,
			})
			if err != nil {
				slog.Error("Failed to update user person_uuid",
					"user_id", user.ID,
					"person_uuid", member.Uid.String(),
					"error", err,
				)
				errorCount++
				continue
			}

			successCount++

			// Log progress every 50 users
			if totalProcessed%50 == 0 {
				slog.Info("Progress",
					"processed", totalProcessed,
					"success", successCount,
					"errors", errorCount,
					"skipped", skippedCount,
				)
			}
		}

		// Small delay between batches to avoid overwhelming the API
		time.Sleep(100 * time.Millisecond)
	}

	// Final summary
	slog.Info("Backfill completed",
		"total_processed", totalProcessed,
		"success", successCount,
		"errors", errorCount,
		"skipped", skippedCount,
	)

	return nil
}
