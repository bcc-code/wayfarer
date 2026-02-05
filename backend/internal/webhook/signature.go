package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// VerifySignature verifies the HMAC-SHA256 signature of a webhook payload.
// The signature is expected in the format "sha256=<hex>".
func VerifySignature(payload []byte, signature string, secretKey string) bool {
	if signature == "" {
		return false
	}

	// Expected format: "sha256=<hex>"
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	providedHash := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(payload)
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedHash), []byte(providedHash))
}
