package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// GraphQLClient is a test client for executing GraphQL queries
type GraphQLClient struct {
	server    *httptest.Server
	authToken string
	headers   map[string]string
}

// GraphQLRequest represents a GraphQL request body
type GraphQLRequest struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName string         `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string         `json:"message"`
	Path       []any          `json:"path,omitempty"`
	Extensions map[string]any `json:"extensions,omitempty"`
}

// NewGraphQLClient creates a client with the test server
func NewGraphQLClient(router *gin.Engine) *GraphQLClient {
	server := httptest.NewServer(router)
	return &GraphQLClient{
		server: server,
	}
}

// WithAuth returns a new client with the authentication token set
func (c *GraphQLClient) WithAuth(token string) *GraphQLClient {
	// Copy existing headers
	newHeaders := make(map[string]string)
	for k, v := range c.headers {
		newHeaders[k] = v
	}
	return &GraphQLClient{
		server:    c.server,
		authToken: token,
		headers:   newHeaders,
	}
}

// WithHeader returns a new client with an additional header set
func (c *GraphQLClient) WithHeader(key, value string) *GraphQLClient {
	newHeaders := make(map[string]string)
	for k, v := range c.headers {
		newHeaders[k] = v
	}
	newHeaders[key] = value
	return &GraphQLClient{
		server:    c.server,
		authToken: c.authToken,
		headers:   newHeaders,
	}
}

// WithLanguage returns a new client with the Accept-Language header set
func (c *GraphQLClient) WithLanguage(lang string) *GraphQLClient {
	return c.WithHeader("Accept-Language", lang)
}

// Execute runs a GraphQL query and returns the response
func (c *GraphQLClient) Execute(ctx context.Context, query string, variables map[string]any) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.server.URL+"/graphql", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	// Add custom headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle non-200 responses (like 401 Unauthorized)
	if resp.StatusCode != http.StatusOK {
		return &GraphQLResponse{
			Errors: []GraphQLError{
				{
					Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
				},
			},
		}, nil
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w (body: %s)", err, string(body))
	}

	return &gqlResp, nil
}

// MustExecute is like Execute but fails the test on error
func (c *GraphQLClient) MustExecute(t *testing.T, query string, variables map[string]any) *GraphQLResponse {
	t.Helper()
	resp, err := c.Execute(context.Background(), query, variables)
	require.NoError(t, err)
	return resp
}

// Close shuts down the test server
func (c *GraphQLClient) Close() {
	if c.server != nil {
		c.server.Close()
	}
}

// HasErrors returns true if the response contains errors
func (r *GraphQLResponse) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorMessage returns the first error message, or empty string if no errors
func (r *GraphQLResponse) ErrorMessage() string {
	if len(r.Errors) == 0 {
		return ""
	}
	return r.Errors[0].Message
}

// UnmarshalData unmarshals the data field into the provided struct
func (r *GraphQLResponse) UnmarshalData(v any) error {
	return json.Unmarshal(r.Data, v)
}
