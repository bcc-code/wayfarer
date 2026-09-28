package firebase

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"

	firebase "firebase.google.com/go/v4"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

func TestNew_MissingConfig(t *testing.T) {
	ctx := context.Background()
	cfg := config.FirebaseConfig{
		ServiceAccountJSON: "",
	}

	_, err := New(ctx, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestNew_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	cfg := config.FirebaseConfig{
		ServiceAccountJSON: "not-valid-json",
	}

	_, err := New(ctx, cfg)
	assert.Error(t, err)
	// Firebase SDK returns an error when the JSON is invalid
}

func TestNew_InvalidBase64(t *testing.T) {
	ctx := context.Background()
	// This is valid base64 but not valid JSON for a service account
	cfg := config.FirebaseConfig{
		ServiceAccountJSON: "aW52YWxpZA==", // "invalid" in base64
	}

	_, err := New(ctx, cfg)
	assert.Error(t, err)
	// Firebase SDK returns an error when credentials are invalid
}

// newTestService builds a Service whose auth client can sign custom tokens
// locally using a freshly generated RSA key. It performs no network calls, so
// it is safe for hermetic unit tests. Only the auth client is populated;
// firestoreClient is left nil (unused by CreateCustomToken).
func newTestService(t *testing.T) *Service {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

	sa := map[string]string{
		"type":         "service_account",
		"project_id":   "test-project",
		"private_key":  string(privatePEM),
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id":    "test-client-id",
		"token_uri":    "https://oauth2.googleapis.com/token",
	}
	saJSON, err := json.Marshal(sa)
	require.NoError(t, err)

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: "test-project"}, option.WithCredentialsJSON(saJSON))
	require.NoError(t, err)

	authClient, err := app.Auth(ctx)
	require.NoError(t, err)

	return &Service{authClient: authClient}
}

// decodeCustomTokenClaims parses the (locally signed, unverified) custom-token
// JWT and returns its payload as a map. Signature verification is not needed —
// the test only asserts on the claims the backend embedded.
func decodeCustomTokenClaims(t *testing.T, token string) map[string]any {
	t.Helper()

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3, "custom token should be a JWT with three segments")

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(payload, &decoded))
	return decoded
}

func TestCreateCustomToken_EmbedsClaims(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	const (
		userID   = "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
		churchID = "CH01ARZ3NDEKTSV4RRFFQ69G5FAV"
	)

	token, err := svc.CreateCustomToken(ctx, userID, churchID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload := decodeCustomTokenClaims(t, token)

	// The token's subject (uid) auto-provisions the Firebase Auth user on sign-in.
	assert.Equal(t, userID, payload["uid"])

	// Developer claims are embedded under the "claims" object and are copied into
	// the ID token, where Firestore rules can read them via request.auth.token.
	claims, ok := payload["claims"].(map[string]any)
	require.True(t, ok, "expected developer claims object in token payload")
	assert.Equal(t, userID, claims["userId"])
	assert.Equal(t, churchID, claims["churchId"])
}

func TestCreateCustomToken_EmptyUserID(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.CreateCustomToken(context.Background(), "", "CH01ARZ3NDEKTSV4RRFFQ69G5FAV")
	assert.Error(t, err)
}
