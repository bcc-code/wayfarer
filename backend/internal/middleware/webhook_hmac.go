package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/bcc-media/wayfarer/internal/webhook"
	"github.com/gin-gonic/gin"
)

// maxWebhookBodyBytes bounds inbound webhook payloads. Generous because the
// quiz-finalized payload can carry many per-user results.
const maxWebhookBodyBytes = 5 << 20 // 5 MiB

// WebhookHMACAuth returns middleware that enforces an HMAC-SHA256 signature
// (X-Webhook-Signature: sha256=<hex>) over a size-bounded request body.
// It reads and restores the body so downstream handlers can ShouldBindJSON.
func WebhookHMACAuth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Defense in depth: routes are not registered without a secret, but never
		// let an empty-secret path authenticate anything.
		if secretKey == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "webhook not configured"})
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxWebhookBodyBytes)
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Warn("webhook: failed to read request body", "path", c.FullPath(), "error", err)
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large or unreadable"})
			return
		}

		if !webhook.VerifySignature(body, c.GetHeader("X-Webhook-Signature"), secretKey) {
			slog.Warn("webhook: invalid signature", "path", c.FullPath())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}

		// Restore the body so downstream handlers can bind it.
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Next()
	}
}
