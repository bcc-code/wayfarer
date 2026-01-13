package firebase

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/stretchr/testify/assert"
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
