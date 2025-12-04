package ssf

import (
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sony/gobreaker/v2"
)

// Config contains configuration options for the SSF client
type Config struct {
	BaseURL   string
	APIKey    string
	DebugMode bool
	Timeout   time.Duration
}

// Client is the base struct for communicating with the SSF API
type Client struct {
	config  Config
	resty   *resty.Client
	breaker *gobreaker.CircuitBreaker[[]byte]
	logger  *slog.Logger
}

// New returns a new SSF client
func New(config Config, logger *slog.Logger) *Client {
	restyClient := resty.New().
		SetBaseURL(config.BaseURL).
		SetTimeout(config.Timeout).
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Bearer "+config.APIKey)

	// Configure debug logging via resty if enabled
	if config.DebugMode {
		restyClient.SetDebug(true)
	}

	// Create circuit breaker
	breaker := gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
		Name:        "ssf-api",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("SSF circuit breaker state changed",
				"name", name,
				"from", from.String(),
				"to", to.String(),
			)
		},
	})

	return &Client{
		config:  config,
		resty:   restyClient,
		breaker: breaker,
		logger:  logger,
	}
}

// IsConfigured returns true if the client has an API key configured
func (c *Client) IsConfigured() bool {
	return c.config.APIKey != ""
}
