package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/bcc-media/wayfarer/internal/auth0"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/services"

	"github.com/sony/gobreaker/v2"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	lgr := logger.New(cfg.Server.Environment, logger.ParseLevel(cfg.Log.Level))
	slog.SetDefault(lgr)

	slog.Info("Starting user import from Members API")

	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if cfg.Auth0.Domain == "" || cfg.Auth0.ClientID == "" || cfg.Members.Domain == "" {
		slog.Error("Missing required configuration",
			"auth0_domain", cfg.Auth0.Domain,
			"members_domain", cfg.Members.Domain,
		)
		os.Exit(1)
	}

	auth0Client := auth0.New(auth0.Config{
		Domain:       cfg.Auth0.Domain,
		ClientID:     cfg.Auth0.ClientID,
		ClientSecret: cfg.Auth0.ClientSecret,
	})

	membersBreaker := gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
		Name:         "members-api-import",
		Timeout:      30 * time.Second,
		IsSuccessful: members.IsBreakerSuccess,
	})

	membersClient := members.New(
		members.Config{Domain: cfg.Members.Domain},
		auth0Client,
		membersBreaker,
	)

	importService := &services.MemberImportService{
		DB:            db,
		MembersClient: membersClient,
	}

	result, err := importService.ImportNewMembers(ctx)
	if err != nil {
		slog.Error("User import failed", "error", err)
		os.Exit(1)
	}

	slog.Info("User import completed successfully",
		"fetched", result.Fetched,
		"imported", result.Imported,
		"skipped", result.Skipped,
		"errors", len(result.Errors),
	)
	for _, e := range result.Errors {
		slog.Warn("member_import: error during run", "detail", e)
	}
}
