package auth0

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
)

var (
	tokenCache = cache.New[string, getTokenResponse]()
	cacheMutex sync.Mutex
)

type getTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   time.Time
	TokenType   string `json:"token_type"`
}

func sendTokenRequest[t any](ctx context.Context, body t, endpoint string) (getTokenResponse, error) {
	marshalled, err := json.Marshal(body)
	if err != nil {
		return getTokenResponse{}, err
	}

	res, err := http.Post(
		endpoint,
		"application/json",
		bytes.NewBuffer(marshalled),
	)
	if err != nil {
		return getTokenResponse{}, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	result, err := io.ReadAll(res.Body)
	if err != nil {
		return getTokenResponse{}, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return getTokenResponse{}, fmt.Errorf("failed to retrieve token (status %d): %s", res.StatusCode, string(result))
	}

	var response getTokenResponse
	err = json.Unmarshal(result, &response)
	if err != nil {
		return getTokenResponse{}, err
	}
	response.ExpiresAt = time.Now().Add(time.Second * time.Duration(response.ExpiresIn))
	return response, nil
}

type getTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	Audience     string `json:"audience"`
}

func (c *Client) getNewToken(ctx context.Context, audience string) (getTokenResponse, error) {
	path, _ := url.Parse(fmt.Sprintf("https://%s/oauth/token", c.config.Domain))
	return sendTokenRequest(ctx, getTokenRequest{
		GrantType:    "client_credentials",
		ClientID:     c.config.ClientID,
		ClientSecret: c.config.ClientSecret,
		Audience:     audience,
	}, path.String())
}

// GetToken returns a token to be used towards other apis with the client credentials
func (c *Client) GetToken(ctx context.Context, audience string) (string, error) {
	// Lock to prevent race conditions on cache access
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Check if we have a valid cached token
	t, ok := tokenCache.Get(audience)
	if ok && t.ExpiresAt.After(time.Now()) {
		return t.AccessToken, nil
	}

	// Token is missing or expired, fetch a new one
	t, err := c.getNewToken(ctx, audience)
	if err != nil {
		return "", err
	}

	// Cache the new token (24 hour expiration as fallback)
	tokenCache.Set(audience, t, cache.WithExpiration(time.Hour*24))

	return t.AccessToken, nil
}
