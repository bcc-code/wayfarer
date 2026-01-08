package firebase

import (
	"context"
	"encoding/base64"
	"fmt"
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

	// Check if it's a file path
	if _, err := os.Stat(cfg.ServiceAccountJSON); err == nil {
		opt = option.WithCredentialsFile(cfg.ServiceAccountJSON)
	} else {
		// Try to decode as base64 JSON
		jsonBytes, err := base64.StdEncoding.DecodeString(cfg.ServiceAccountJSON)
		if err != nil {
			// Not base64, assume it's raw JSON
			jsonBytes = []byte(cfg.ServiceAccountJSON)
		}
		opt = option.WithCredentialsJSON(jsonBytes)
	}

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Auth client: %w", err)
	}

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firestore client: %w", err)
	}

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
// Custom claims (userId, churchId) are set on the user record itself.
// Custom tokens are valid for 1 hour and should be exchanged client-side
// for Firebase ID tokens which auto-refresh.
// If the user doesn't exist in Firebase Auth, they will be created first.
func (s *Service) CreateCustomToken(ctx context.Context, userID, churchID string) (string, error) {
	// Ensure user exists in Firebase Auth
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return "", fmt.Errorf("failed to ensure user exists: %w", err)
	}

	// Set custom claims on the user record
	claims := map[string]interface{}{
		"userId":   userID,
		"churchId": churchID,
	}
	if err := s.authClient.SetCustomUserClaims(ctx, userID, claims); err != nil {
		return "", fmt.Errorf("failed to set custom claims: %w", err)
	}

	token, err := s.authClient.CustomToken(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to create custom token: %w", err)
	}

	return token, nil
}

// ensureUserExists checks if a user exists in Firebase Auth and creates them if not.
func (s *Service) ensureUserExists(ctx context.Context, userID string) error {
	// Try to get the user
	_, err := s.authClient.GetUser(ctx, userID)
	if err == nil {
		// User exists
		return nil
	}

	// Check if it's a "user not found" error
	if !auth.IsUserNotFound(err) {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// User doesn't exist, create them
	email := fmt.Sprintf("%s@fake-firestore-email.bcc.media", userID)
	params := (&auth.UserToCreate{}).
		UID(userID).
		Email(email)

	_, err = s.authClient.CreateUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
