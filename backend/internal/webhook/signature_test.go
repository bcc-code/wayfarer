package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifySignature(t *testing.T) {
	secretKey := "test-secret-key"
	payload := []byte(`{"event_type":"test","timestamp":"2024-01-01T00:00:00Z"}`)

	// Pre-computed invalid signature for testing
	invalidSignature := "sha256=f7d4b8e3c9a1b2d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9"

	tests := []struct {
		name      string
		payload   []byte
		signature string
		secretKey string
		expected  bool
	}{
		{
			name:      "empty signature",
			payload:   payload,
			signature: "",
			secretKey: secretKey,
			expected:  false,
		},
		{
			name:      "missing sha256 prefix",
			payload:   payload,
			signature: "invalid-format",
			secretKey: secretKey,
			expected:  false,
		},
		{
			name:      "invalid hash",
			payload:   payload,
			signature: "sha256=invalidhash",
			secretKey: secretKey,
			expected:  false,
		},
		{
			name:      "wrong payload",
			payload:   []byte(`{"different":"payload"}`),
			signature: invalidSignature,
			secretKey: secretKey,
			expected:  false,
		},
		{
			name:      "wrong secret key",
			payload:   payload,
			signature: invalidSignature,
			secretKey: "wrong-key",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifySignature(tt.payload, tt.signature, tt.secretKey)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test with correctly computed signature
	t.Run("valid signature", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte(secretKey))
		mac.Write(payload)
		correctSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		result := VerifySignature(payload, correctSignature, secretKey)
		assert.True(t, result, "Valid signature should be accepted")
	})
}
