package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWebhookSecret = "test-webhook-secret"

// signBody computes the X-Webhook-Signature value the middleware expects.
func signBody(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// newWebhookRouter builds a router with the HMAC middleware and a handler that
// echoes the (already restored) request body so we can assert it survived.
func newWebhookRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/hook", WebhookHMACAuth(secret), func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "read failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"body": string(body)})
	})
	return router
}

func TestWebhookHMACAuth_ValidSignature(t *testing.T) {
	router := newWebhookRouter(testWebhookSecret)
	body := []byte(`{"event":"test"}`)

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	req.Header.Set("X-Webhook-Signature", signBody(body, testWebhookSecret))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Body must be intact for the downstream handler.
	assert.Contains(t, w.Body.String(), `{\"event\":\"test\"}`)
}

func TestWebhookHMACAuth_MissingSignature(t *testing.T) {
	router := newWebhookRouter(testWebhookSecret)
	body := []byte(`{"event":"test"}`)

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	// No X-Webhook-Signature header.

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid signature")
}

func TestWebhookHMACAuth_InvalidSignature(t *testing.T) {
	router := newWebhookRouter(testWebhookSecret)
	body := []byte(`{"event":"test"}`)

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	req.Header.Set("X-Webhook-Signature", "sha256=deadbeef")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid signature")
}

func TestWebhookHMACAuth_SignatureForDifferentBody(t *testing.T) {
	router := newWebhookRouter(testWebhookSecret)
	body := []byte(`{"event":"test"}`)
	// Sign a different payload than the one we send (tampered body).
	sig := signBody([]byte(`{"event":"other"}`), testWebhookSecret)

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	req.Header.Set("X-Webhook-Signature", sig)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWebhookHMACAuth_EmptySecret(t *testing.T) {
	router := newWebhookRouter("")
	body := []byte(`{"event":"test"}`)

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	req.Header.Set("X-Webhook-Signature", signBody(body, ""))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Defense in depth: an empty secret must never authenticate anything.
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "webhook not configured")
}

func TestWebhookHMACAuth_BodyTooLarge(t *testing.T) {
	router := newWebhookRouter(testWebhookSecret)
	// Exceed maxWebhookBodyBytes (5 MiB).
	body := []byte(strings.Repeat("a", maxWebhookBodyBytes+1))

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	req.Header.Set("X-Webhook-Signature", signBody(body, testWebhookSecret))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestWebhookHMACAuth_BodyAtLimit(t *testing.T) {
	router := newWebhookRouter(testWebhookSecret)
	// Exactly at the limit must be accepted.
	body := []byte(strings.Repeat("a", maxWebhookBodyBytes))

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	req.Header.Set("X-Webhook-Signature", signBody(body, testWebhookSecret))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
