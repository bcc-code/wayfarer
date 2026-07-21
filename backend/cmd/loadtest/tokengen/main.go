package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

// LoadTestConfig is the output structure for k6 consumption
type LoadTestConfig struct {
	BaseURL   string      `json:"baseUrl"`
	Tokens    []UserToken `json:"tokens"`
	Generated time.Time   `json:"generated"`
}

// UserToken contains a user ID and their JWT token
type UserToken struct {
	UserID string `json:"userId"`
	Token  string `json:"token"`
}

// WayfarerClaims mirrors the middleware claims structure
type WayfarerClaims struct {
	UserID    string   `json:"user_id"`
	UserRoles []string `json:"user_roles"`
	jwt.RegisteredClaims
}

func main() {
	// Parse command line flags
	limit := flag.Int("limit", 1000, "Number of users to generate tokens for")
	output := flag.String("output", "../config.json", "Output file path")
	baseURL := flag.String("base-url", "http://localhost:8080", "GraphQL API base URL")
	validDays := flag.Int("valid-days", 7, "JWT token validity in days")
	secret := flag.String("secret", "", "JWT secret (defaults to JWT_SECRET env var)")
	flag.Parse()

	// Determine JWT secret
	jwtSecret := *secret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if jwtSecret == "" {
		fmt.Fprintln(os.Stderr, "Error: JWT secret is required (set --secret or JWT_SECRET)")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	lgr := logger.New(cfg.Server.Environment, logger.ParseLevel(cfg.Log.Level))
	slog.SetDefault(lgr)

	slog.Info("Starting load test token generation",
		"limit", *limit,
		"output", *output,
		"baseUrl", *baseURL,
		"validDays", *validDays,
	)

	// Connect to database
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Query user IDs from database
	slog.Info("Querying user IDs from database")
	rows, err := db.Pool.Query(ctx, "SELECT id FROM users ORDER BY id LIMIT @limit", pgx.NamedArgs{"limit": *limit})
	if err != nil {
		slog.Error("Failed to query users", "error", err)
		os.Exit(1)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			slog.Error("Failed to scan user ID", "error", err)
			os.Exit(1)
		}
		userIDs = append(userIDs, id)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Error iterating rows", "error", err)
		os.Exit(1)
	}

	slog.Info("Found users", "count", len(userIDs))

	if len(userIDs) == 0 {
		slog.Error("No users found in database. Run 'make seed-large' first.")
		os.Exit(1)
	}

	// Generate JWT tokens for each user
	slog.Info("Generating JWT tokens")
	tokens := make([]UserToken, 0, len(userIDs))
	now := time.Now()

	for _, userID := range userIDs {
		claims := WayfarerClaims{
			UserID:    userID,
			UserRoles: []string{"user"},
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "wayfarer",
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(*validDays) * 24 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			slog.Error("Failed to sign token", "userId", userID, "error", err)
			os.Exit(1)
		}

		tokens = append(tokens, UserToken{
			UserID: userID,
			Token:  tokenString,
		})
	}

	// Create output config
	outputConfig := LoadTestConfig{
		BaseURL:   *baseURL,
		Tokens:    tokens,
		Generated: now,
	}

	// Write JSON output file
	jsonData, err := json.MarshalIndent(outputConfig, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal JSON", "error", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, jsonData, 0644); err != nil {
		slog.Error("Failed to write output file", "error", err)
		os.Exit(1)
	}

	slog.Info("Token generation complete",
		"tokens", len(tokens),
		"output", *output,
	)
}
