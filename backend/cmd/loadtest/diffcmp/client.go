package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type gqlResponse struct {
	Status int
	Body   []byte
	Data   any
	Errors []gqlError
}

type gqlError struct {
	Message string `json:"message"`
	Path    []any  `json:"path"`
}

var httpClient = &http.Client{Timeout: 60 * time.Second}

// gqlPost fires one GraphQL request. Accept-Language is pinned so the branch's
// response-cache key language component is constant across the run.
func gqlPost(baseURL, token, query string, vars map[string]any) (*gqlResponse, error) {
	payload, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/graphql", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	out := &gqlResponse{Status: resp.StatusCode, Body: body}
	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []gqlError      `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("non-JSON response (status %d): %.200s", resp.StatusCode, body)
	}
	out.Errors = envelope.Errors
	if len(envelope.Data) > 0 {
		dec := json.NewDecoder(bytes.NewReader(envelope.Data))
		dec.UseNumber()
		if err := dec.Decode(&out.Data); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// restRequest fires a plain HTTP request (plugin webhooks, /metrics/http).
// When webhookSecret is non-empty the body is HMAC-signed the way the Ladder
// to Heaven plugin expects.
func restRequest(baseURL, method, path string, body []byte, bearer, webhookSecret string) (*gqlResponse, error) {
	req, err := http.NewRequest(method, baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if webhookSecret != "" {
		mac := hmac.New(sha256.New, []byte(webhookSecret))
		mac.Write(body)
		req.Header.Set("X-Webhook-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	out := &gqlResponse{Status: resp.StatusCode, Body: raw}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	// Non-JSON bodies are fine for REST endpoints; keep Data nil then.
	_ = dec.Decode(&out.Data)
	return out, nil
}

// wayfarerClaims mirrors the middleware claims structure (see cmd/gentoken).
type wayfarerClaims struct {
	UserID    string   `json:"user_id"`
	UserRoles []string `json:"user_roles"`
	jwt.RegisteredClaims
}

// mintToken signs an HS256 JWT compatible with the server's JWT middleware.
func mintToken(secret, userID string, roles []string) (string, error) {
	now := time.Now()
	claims := wayfarerClaims{
		UserID:    userID,
		UserRoles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wayfarer",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// seededID reproduces the bench dataset's mkid(prefix, n):
// prefix + upper(hex(n)) left-padded to 26 chars.
func seededID(prefix string, n int) string {
	return fmt.Sprintf("%s%026X", prefix, n)
}

func seededUserID(n int) string {
	return seededID("US", n)
}
