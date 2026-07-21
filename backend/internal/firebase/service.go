package firebase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/bcc-media/wayfarer/internal/config"
	"google.golang.org/api/option"
)

// Service handles Firebase Admin SDK operations
type Service struct {
	authClient      *auth.Client
	firestoreClient *firestore.Client
}

// New creates a new Firebase service from configuration.
// The service account can be provided as:
// - A file path to a JSON file
// - Base64-encoded JSON content
// - Raw JSON content
func New(ctx context.Context, cfg config.FirebaseConfig) (*Service, error) {
	if cfg.ServiceAccountJSON == "" {
		return nil, fmt.Errorf("firebase service account not configured")
	}

	var opt option.ClientOption
	var projectID string

	// Check if it's a file path
	if _, err := os.Stat(cfg.ServiceAccountJSON); err == nil {
		opt = option.WithCredentialsFile(cfg.ServiceAccountJSON)
		// Extract project ID from file
		projectID, err = extractProjectIDFromFile(cfg.ServiceAccountJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to extract project ID from credentials file: %w", err)
		}
	} else {
		// Try to decode as base64 JSON
		jsonBytes, err := base64.StdEncoding.DecodeString(cfg.ServiceAccountJSON)
		if err != nil {
			// Not base64, assume it's raw JSON
			jsonBytes = []byte(cfg.ServiceAccountJSON)
		}
		opt = option.WithCredentialsJSON(jsonBytes)
		// Extract project ID from JSON
		projectID, err = extractProjectIDFromJSON(jsonBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to extract project ID from credentials: %w", err)
		}
	}

	// Create Firebase config with project ID
	firebaseConfig := &firebase.Config{
		ProjectID: projectID,
	}

	slog.Debug("Initializing Firebase app", "projectID", projectID)

	app, err := firebase.NewApp(ctx, firebaseConfig, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Auth client: %w", err)
	}

	// Create Firestore client directly instead of through Firebase app
	// This ensures we have a properly initialized client with working Doc() method
	// Use the configured database name (defaults to "(default)")
	databaseName := cfg.DatabaseName
	if databaseName == "" {
		databaseName = "(default)"
	}

	slog.Debug("Creating Firestore client", "projectID", projectID, "database", databaseName)
	firestoreClient, err := firestore.NewClientWithDatabase(ctx, projectID, databaseName, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	if firestoreClient == nil {
		return nil, fmt.Errorf("firestore client is nil after initialization")
	}

	slog.Debug("Firestore client initialized", "projectID", projectID, "database", databaseName)

	return &Service{
		authClient:      authClient,
		firestoreClient: firestoreClient,
	}, nil
}

// Close closes the Firestore client connection.
func (s *Service) Close() error {
	if s.firestoreClient != nil {
		return s.firestoreClient.Close()
	}
	return nil
}

// CreateCustomToken generates a Firebase custom token for the given user.
// The custom claims (userId, churchId) are embedded directly in the token as
// developer claims; Firebase copies them into the ID token minted when the
// client calls signInWithCustomToken, so they are available in Firestore
// security rules via request.auth.token.
//
// Token generation is a local signing operation using the service account's
// private key — it makes no Firebase Admin API network calls and is not subject
// to Firebase rate limits. The Firebase Auth user record is auto-provisioned by
// signInWithCustomToken, so no GetUser/CreateUser round-trip is needed here.
//
// Custom tokens are valid for 1 hour and should be exchanged client-side for
// Firebase ID tokens which auto-refresh.
func (s *Service) CreateCustomToken(ctx context.Context, userID, churchID string) (string, error) {
	claims := map[string]any{
		"userId":   userID,
		"churchId": churchID,
	}

	token, err := s.authClient.CustomTokenWithClaims(ctx, userID, claims)
	if err != nil {
		return "", fmt.Errorf("failed to create custom token: %w", err)
	}

	return token, nil
}

// extractProjectIDFromFile reads a service account JSON file and extracts the project_id.
func extractProjectIDFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read credentials file: %w", err)
	}
	return extractProjectIDFromJSON(data)
}

// extractProjectIDFromJSON parses service account JSON and extracts the project_id.
func extractProjectIDFromJSON(jsonData []byte) (string, error) {
	var creds struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(jsonData, &creds); err != nil {
		return "", fmt.Errorf("failed to parse credentials JSON: %w", err)
	}
	if creds.ProjectID == "" {
		return "", fmt.Errorf("project_id not found in credentials")
	}
	return creds.ProjectID, nil
}
