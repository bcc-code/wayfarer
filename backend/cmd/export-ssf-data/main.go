package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/bcc-media/wayfarer/internal/ssf"
	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: export-ssf-data <person-uuid>")
		fmt.Fprintln(os.Stderr, "Example: export-ssf-data 550e8400-e29b-41d4-a716-446655440000")
		os.Exit(1)
	}

	personUUID, err := uuid.Parse(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid UUID: %v\n", err)
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger (logs go to stderr so JSON output can go to stdout)
	lgr := logger.New(cfg.Server.Environment, logger.ParseLevel(cfg.Log.Level))
	slog.SetDefault(lgr)

	// Validate SSF configuration
	if cfg.SSF.APIKey == "" {
		slog.Error("SSF API key not configured")
		os.Exit(1)
	}

	// Initialize SSF client
	ssfClient := ssf.New(ssf.Config{
		BaseURL:   cfg.SSF.BaseURL,
		APIKey:    cfg.SSF.APIKey,
		DebugMode: cfg.SSF.DebugMode,
		Timeout:   cfg.SSF.Timeout,
	}, lgr)

	// Fetch all content events with pagination
	var allEvents []ssf.ContentEvent
	page := 1

	slog.Info("Fetching SSF data", "person_uuid", personUUID.String())

	for {
		result, err := ssfClient.GetMemberContentEvents(ctx, personUUID.String(), page)
		if err != nil {
			slog.Error("Failed to fetch content events", "page", page, "error", err)
			os.Exit(1)
		}

		allEvents = append(allEvents, result.Items...)

		slog.Info("Fetched page",
			"page", page,
			"items", len(result.Items),
			"total_so_far", len(allEvents),
			"has_more", result.HasMore,
		)

		if !result.HasMore {
			break
		}

		page++
		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	// Output JSON to stdout
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(allEvents); err != nil {
		slog.Error("Failed to encode JSON", "error", err)
		os.Exit(1)
	}

	slog.Info("Export completed", "total_events", len(allEvents))
}
