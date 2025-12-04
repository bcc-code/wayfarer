package ssf

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// executeRequest performs a GET request through the circuit breaker
func (c *Client) executeRequest(ctx context.Context, endpoint string, acceptLanguage string) ([]byte, error) {
	startTime := time.Now()

	c.logger.Debug("SSF API request starting",
		"endpoint", endpoint,
		"language", acceptLanguage,
	)

	body, err := c.breaker.Execute(func() ([]byte, error) {
		resp, err := c.resty.R().
			SetContext(ctx).
			SetHeader("Accept-Language", acceptLanguage).
			Get(endpoint)

		if err != nil {
			c.logger.Error("SSF API request failed",
				"endpoint", endpoint,
				"error", err,
				"duration_ms", time.Since(startTime).Milliseconds(),
			)
			return nil, fmt.Errorf("SSF API request failed: %w", err)
		}

		if resp.IsError() {
			c.logger.Error("SSF API returned error status",
				"endpoint", endpoint,
				"status", resp.StatusCode(),
				"body", string(resp.Body()),
				"duration_ms", time.Since(startTime).Milliseconds(),
			)
			return nil, fmt.Errorf("SSF API returned status %d: %s", resp.StatusCode(), string(resp.Body()))
		}

		c.logger.Debug("SSF API request completed",
			"endpoint", endpoint,
			"status", resp.StatusCode(),
			"size_bytes", len(resp.Body()),
			"duration_ms", time.Since(startTime).Milliseconds(),
		)

		return resp.Body(), nil
	})

	return body, err
}

// get performs a GET request and unmarshals the response
func get[T any](ctx context.Context, c *Client, endpoint string, acceptLanguage string) (*T, error) {
	body, err := c.executeRequest(ctx, endpoint, acceptLanguage)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("Failed to unmarshal SSF API response",
			"endpoint", endpoint,
			"error", err,
		)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}
