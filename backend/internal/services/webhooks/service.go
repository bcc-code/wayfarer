package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
)

// Service handles webhook dispatching
type Service struct {
	queries    *sqlc.Queries
	httpClient *http.Client
}

// NewService creates a new webhook dispatcher service
func NewService(queries *sqlc.Queries) *Service {
	return &Service{
		queries: queries,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// DispatchExternalContentEvent sends webhooks for external content events to a specific project
// This should be called asynchronously (in a goroutine)
func (s *Service) DispatchExternalContentEvent(ctx context.Context, projectID string, user *UserData, data ExternalContentEventData) {
	s.dispatch(ctx, projectID, EventTypeExternalContent, user, data)
}

// DispatchGlobalExternalContentEvent sends webhooks for external content events to ALL active projects
// This is used when an external content event is received and should notify all interested projects
// This should be called asynchronously (in a goroutine)
func (s *Service) DispatchGlobalExternalContentEvent(ctx context.Context, user *UserData, data ExternalContentEventData) {
	s.dispatchGlobal(ctx, EventTypeExternalContent, user, data)
}

// DispatchPointsAwarded sends webhooks for points awarded events
// This should be called asynchronously (in a goroutine)
func (s *Service) DispatchPointsAwarded(ctx context.Context, projectID string, user *UserData, data PointsAwardedData) {
	s.dispatch(ctx, projectID, EventTypePointsAwarded, user, data)
}

// dispatchGlobal sends the payload to all active webhooks for the given event type across all active projects
func (s *Service) dispatchGlobal(ctx context.Context, eventType EventType, user *UserData, data interface{}) {
	webhooks, err := s.queries.GetActiveWebhooksByEventType(ctx, string(eventType))
	if err != nil {
		slog.Error("webhooks: failed to get active webhooks for global dispatch",
			"event_type", eventType,
			"error", err,
		)
		return
	}

	if len(webhooks) == 0 {
		return
	}

	for _, webhook := range webhooks {
		payload := WebhookPayload{
			EventType: string(eventType),
			Timestamp: time.Now().UTC(),
			ProjectID: webhook.ProjectID,
			Data:      data,
		}

		// Include user data if configured
		if webhook.IncludeUserData && user != nil {
			payload.User = user
		}

		// Only include data if configured
		if !webhook.IncludeEventData {
			payload.Data = nil
		}

		s.dispatchToWebhook(ctx, webhook, payload)
	}
}

// dispatch sends the payload to all active webhooks for the given project and event type
func (s *Service) dispatch(ctx context.Context, projectID string, eventType EventType, user *UserData, data interface{}) {
	webhooks, err := s.queries.GetActiveWebhooksByProjectAndEvent(ctx, sqlc.GetActiveWebhooksByProjectAndEventParams{
		Projectid: projectID,
		Eventtype: string(eventType),
	})
	if err != nil {
		slog.Error("webhooks: failed to get active webhooks",
			"project_id", projectID,
			"event_type", eventType,
			"error", err,
		)
		return
	}

	if len(webhooks) == 0 {
		return
	}

	payload := WebhookPayload{
		EventType: string(eventType),
		Timestamp: time.Now().UTC(),
		ProjectID: projectID,
		Data:      data,
	}

	for _, webhook := range webhooks {
		// Include user data if configured
		if webhook.IncludeUserData && user != nil {
			payload.User = user
		} else {
			payload.User = nil
		}

		// Only include data if configured
		if !webhook.IncludeEventData {
			payload.Data = nil
		} else {
			payload.Data = data
		}

		s.dispatchToWebhook(ctx, webhook, payload)
	}
}

// dispatchToWebhook sends the payload to a single webhook endpoint and logs the result
func (s *Service) dispatchToWebhook(ctx context.Context, webhook *sqlc.Webhook, payload WebhookPayload) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("webhooks: failed to marshal payload",
			"webhook_id", webhook.ID,
			"error", err,
		)
		return
	}

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.Url, bytes.NewReader(payloadBytes))
	if err != nil {
		s.logWebhookResult(ctx, webhook, payloadBytes, nil, nil, time.Since(start), fmt.Errorf("failed to create request: %w", err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Wayfarer-Webhook/1.0")

	// Add HMAC signature if secret is configured
	if webhook.Secret != nil && *webhook.Secret != "" {
		signature := s.signPayload(payloadBytes, *webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := s.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		s.logWebhookResult(ctx, webhook, payloadBytes, nil, nil, duration, err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response body (limited to 10KB)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024))
	if err != nil {
		slog.Warn("webhooks: failed to read response body",
			"webhook_id", webhook.ID,
			"error", err,
		)
		bodyBytes = nil
	}

	var bodyStr *string
	if len(bodyBytes) > 0 {
		s := string(bodyBytes)
		bodyStr = &s
	}

	statusCode := int32(resp.StatusCode)
	s.logWebhookResult(ctx, webhook, payloadBytes, &statusCode, bodyStr, duration, nil)
}

// signPayload creates an HMAC-SHA256 signature of the payload
func (s *Service) signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// logWebhookResult creates a log entry for the webhook dispatch
func (s *Service) logWebhookResult(ctx context.Context, webhook *sqlc.Webhook, payload []byte, statusCode *int32, responseBody *string, duration time.Duration, err error) {
	logID := ulid.NewWebhookLogID()

	var errMsg *string
	if err != nil {
		msg := err.Error()
		errMsg = &msg
	}

	durationMs := int32(duration.Milliseconds())

	_, logErr := s.queries.CreateWebhookLog(ctx, sqlc.CreateWebhookLogParams{
		ID:                 logID,
		Webhookid:          webhook.ID,
		Eventtype:          webhook.EventType,
		Requestpayload:     payload,
		Responsestatuscode: statusCode,
		Responsebody:       responseBody,
		Durationms:         durationMs,
		Errormessage:       errMsg,
	})
	if logErr != nil {
		slog.Error("webhooks: failed to create log entry",
			"webhook_id", webhook.ID,
			"error", logErr,
		)
	}

	if err != nil {
		slog.Warn("webhooks: dispatch failed",
			"webhook_id", webhook.ID,
			"url", webhook.Url,
			"duration_ms", durationMs,
			"error", err,
		)
	} else {
		slog.Debug("webhooks: dispatch completed",
			"webhook_id", webhook.ID,
			"url", webhook.Url,
			"status_code", statusCode,
			"duration_ms", durationMs,
		)
	}
}

// SendTestWebhook sends a test payload to a webhook and returns the log entry
func (s *Service) SendTestWebhook(ctx context.Context, webhookID string) (*sqlc.WebhookLog, error) {
	webhook, err := s.queries.GetWebhookByID(ctx, webhookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	payload := WebhookPayload{
		EventType: webhook.EventType,
		Timestamp: time.Now().UTC(),
		ProjectID: webhook.ProjectID,
		User: &UserData{
			ID:        "test-user-id",
			MembersID: "test-members-id",
			Email:     "test@example.com",
			Name:      "Test User",
		},
		Data: TestEventData{
			Message: "This is a test webhook from Wayfarer",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.Url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Wayfarer-Webhook/1.0")
	req.Header.Set("X-Webhook-Test", "true")

	if webhook.Secret != nil && *webhook.Secret != "" {
		signature := s.signPayload(payloadBytes, *webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := s.httpClient.Do(req)
	duration := time.Since(start)

	logID := ulid.NewWebhookLogID()
	durationMs := int32(duration.Milliseconds())

	if err != nil {
		errMsg := err.Error()
		log, logErr := s.queries.CreateWebhookLog(ctx, sqlc.CreateWebhookLogParams{
			ID:                 logID,
			Webhookid:          webhook.ID,
			Eventtype:          webhook.EventType,
			Requestpayload:     payloadBytes,
			Responsestatuscode: nil,
			Responsebody:       nil,
			Durationms:         durationMs,
			Errormessage:       &errMsg,
		})
		if logErr != nil {
			return nil, fmt.Errorf("failed to create log: %w", logErr)
		}
		return log, nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 10*1024))
	var bodyStr *string
	if len(bodyBytes) > 0 {
		s := string(bodyBytes)
		bodyStr = &s
	}

	statusCode := int32(resp.StatusCode)
	log, logErr := s.queries.CreateWebhookLog(ctx, sqlc.CreateWebhookLogParams{
		ID:                 logID,
		Webhookid:          webhook.ID,
		Eventtype:          webhook.EventType,
		Requestpayload:     payloadBytes,
		Responsestatuscode: &statusCode,
		Responsebody:       bodyStr,
		Durationms:         durationMs,
		Errormessage:       nil,
	})
	if logErr != nil {
		return nil, fmt.Errorf("failed to create log: %w", logErr)
	}

	return log, nil
}
