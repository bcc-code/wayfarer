package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
)

// NotificationType represents the type of push notification
type NotificationType string

const (
	NotificationTypeAchievementUnlocked NotificationType = "achievement_unlocked"
	NotificationTypeChallengeAvailable  NotificationType = "challenge_available"
	NotificationTypeGeneric             NotificationType = "generic"
)

// PushPayload matches the frontend's expected push notification format
type PushPayload struct {
	Title          string                 `json:"title"`
	Body           string                 `json:"body"`
	Icon           string                 `json:"icon,omitempty"`
	Badge          string                 `json:"badge,omitempty"`
	NotificationID string                 `json:"notificationId"`
	Type           NotificationType       `json:"type"`
	URL            string                 `json:"url,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Tag            string                 `json:"tag,omitempty"`
}

// TargetCriteria defines how to select notification recipients
type TargetCriteria struct {
	UserIDs    []string `json:"userIds,omitempty"`
	TeamIDs    []string `json:"teamIds,omitempty"`
	ProjectIDs []string `json:"projectIds,omitempty"`
	EventIDs   []string `json:"eventIds,omitempty"`
	AllUsers   bool     `json:"allUsers,omitempty"`
}

// SendResult contains the results of a notification send operation
type SendResult struct {
	TotalRecipients       int
	SuccessfulDeliveries  int
	FailedDeliveries      int
	InvalidSubscriptionIDs []string
}

// Querier defines the database operations needed for push notifications
type Querier interface {
	// Subscription management
	CreatePushSubscription(ctx context.Context, params sqlc.CreatePushSubscriptionParams) (*sqlc.PushSubscription, error)
	UpdatePushSubscription(ctx context.Context, params sqlc.UpdatePushSubscriptionParams) (*sqlc.PushSubscription, error)
	GetPushSubscriptionByEndpoint(ctx context.Context, endpoint string) (*sqlc.PushSubscription, error)
	DeletePushSubscriptionByEndpoint(ctx context.Context, endpoint string) error
	DeletePushSubscriptionByID(ctx context.Context, id string) error
	GetPushSubscriptionsByUserID(ctx context.Context, userid string) ([]*sqlc.PushSubscription, error)

	// User targeting
	GetUserIDsInTeams(ctx context.Context, teamids []string) ([]string, error)
	GetUserIDsInProjects(ctx context.Context, projectids []string) ([]string, error)
	GetUserIDsInEvents(ctx context.Context, eventids []string) ([]string, error)
	GetAllSubscribedUserIDs(ctx context.Context) ([]string, error)
	GetEnabledSubscriptionsForUsers(ctx context.Context, params sqlc.GetEnabledSubscriptionsForUsersParams) ([]*sqlc.PushSubscription, error)

	// Preferences
	GetUserNotificationPreferences(ctx context.Context, userid string) ([]*sqlc.PushNotificationPreference, error)
	UpsertNotificationPreference(ctx context.Context, params sqlc.UpsertNotificationPreferenceParams) (*sqlc.PushNotificationPreference, error)
	IsNotificationTypeEnabled(ctx context.Context, params sqlc.IsNotificationTypeEnabledParams) (bool, error)

	// Logging
	CreatePushNotificationLog(ctx context.Context, params sqlc.CreatePushNotificationLogParams) (*sqlc.PushNotificationLog, error)
	UpdatePushNotificationLogStats(ctx context.Context, params sqlc.UpdatePushNotificationLogStatsParams) error
}

// Service handles push notification operations
type Service struct {
	queries     Querier
	vapidConfig config.VAPIDConfig
	logger      *slog.Logger
}

// NewService creates a new push notification service
func NewService(queries Querier, vapidConfig config.VAPIDConfig, logger *slog.Logger) *Service {
	return &Service{
		queries:     queries,
		vapidConfig: vapidConfig,
		logger:      logger,
	}
}

// IsConfigured returns true if VAPID keys are configured
func (s *Service) IsConfigured() bool {
	return s.vapidConfig.PublicKey != "" && s.vapidConfig.PrivateKey != ""
}

// SendNotification sends a push notification to recipients matching the criteria
func (s *Service) SendNotification(ctx context.Context, payload PushPayload, criteria TargetCriteria, senderID *string) (*SendResult, error) {
	if !s.IsConfigured() {
		return nil, fmt.Errorf("VAPID keys not configured")
	}

	// Resolve target user IDs
	userIDs, err := s.resolveTargetUserIDs(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target users: %w", err)
	}

	if len(userIDs) == 0 {
		s.logger.Info("no target users found for notification", "criteria", criteria)
		return &SendResult{}, nil
	}

	// Get subscriptions for users who have this notification type enabled
	subscriptions, err := s.queries.GetEnabledSubscriptionsForUsers(ctx, sqlc.GetEnabledSubscriptionsForUsersParams{
		Userids:          userIDs,
		Notificationtype: string(payload.Type),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		s.logger.Info("no subscriptions found for target users", "userCount", len(userIDs))
		return &SendResult{TotalRecipients: len(userIDs)}, nil
	}

	// Generate notification ID if not provided
	if payload.NotificationID == "" {
		payload.NotificationID = ulid.NewPushNotificationID()
	}

	// Send to all subscriptions
	result := s.sendToSubscriptions(ctx, subscriptions, payload)

	// Log the notification
	criteriaJSON, _ := json.Marshal(criteria)
	dataJSON, _ := json.Marshal(payload.Data)

	sentBy := ""
	if senderID != nil {
		sentBy = *senderID
	}

	_, logErr := s.queries.CreatePushNotificationLog(ctx, sqlc.CreatePushNotificationLogParams{
		ID:                   ulid.NewPushNotificationID(),
		Notificationtype:     string(payload.Type),
		Title:                payload.Title,
		Body:                 payload.Body,
		Url:                  payload.URL,
		Data:                 dataJSON,
		Targetcriteria:       criteriaJSON,
		Sentby:               sentBy,
		Totalrecipients:      int32(result.TotalRecipients),
		Successfuldeliveries: int32(result.SuccessfulDeliveries),
		Faileddeliveries:     int32(result.FailedDeliveries),
	})
	if logErr != nil {
		s.logger.Error("failed to log notification", "error", logErr)
	}

	return result, nil
}

// SendToUser sends a push notification to a specific user
func (s *Service) SendToUser(ctx context.Context, userID string, payload PushPayload) (*SendResult, error) {
	return s.SendNotification(ctx, payload, TargetCriteria{UserIDs: []string{userID}}, nil)
}

// SendToUsers sends a push notification to specific users
func (s *Service) SendToUsers(ctx context.Context, userIDs []string, payload PushPayload) (*SendResult, error) {
	return s.SendNotification(ctx, payload, TargetCriteria{UserIDs: userIDs}, nil)
}

// resolveTargetUserIDs resolves the target criteria to a list of user IDs
func (s *Service) resolveTargetUserIDs(ctx context.Context, criteria TargetCriteria) ([]string, error) {
	userIDSet := make(map[string]struct{})

	// Add direct user IDs
	for _, id := range criteria.UserIDs {
		userIDSet[id] = struct{}{}
	}

	// Add users from teams
	if len(criteria.TeamIDs) > 0 {
		teamUserIDs, err := s.queries.GetUserIDsInTeams(ctx, criteria.TeamIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get users in teams: %w", err)
		}
		for _, id := range teamUserIDs {
			userIDSet[id] = struct{}{}
		}
	}

	// Add users from projects
	if len(criteria.ProjectIDs) > 0 {
		projectUserIDs, err := s.queries.GetUserIDsInProjects(ctx, criteria.ProjectIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get users in projects: %w", err)
		}
		for _, id := range projectUserIDs {
			userIDSet[id] = struct{}{}
		}
	}

	// Add users from events
	if len(criteria.EventIDs) > 0 {
		eventUserIDs, err := s.queries.GetUserIDsInEvents(ctx, criteria.EventIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get users in events: %w", err)
		}
		for _, id := range eventUserIDs {
			userIDSet[id] = struct{}{}
		}
	}

	// Add all subscribed users
	if criteria.AllUsers {
		allUserIDs, err := s.queries.GetAllSubscribedUserIDs(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get all subscribed users: %w", err)
		}
		for _, id := range allUserIDs {
			userIDSet[id] = struct{}{}
		}
	}

	// Convert set to slice
	result := make([]string, 0, len(userIDSet))
	for id := range userIDSet {
		result = append(result, id)
	}

	return result, nil
}

// sendToSubscriptions sends the payload to all subscriptions and handles errors
func (s *Service) sendToSubscriptions(ctx context.Context, subscriptions []*sqlc.PushSubscription, payload PushPayload) *SendResult {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("failed to marshal payload", "error", err)
		return &SendResult{
			TotalRecipients:  len(subscriptions),
			FailedDeliveries: len(subscriptions),
		}
	}

	result := &SendResult{
		TotalRecipients:        len(subscriptions),
		InvalidSubscriptionIDs: make([]string, 0),
	}

	for _, sub := range subscriptions {
		err := s.sendToSubscription(sub, payloadBytes)
		if err != nil {
			result.FailedDeliveries++
			s.logger.Warn("failed to send push notification",
				"subscriptionID", sub.ID,
				"userID", sub.UserID,
				"error", err,
			)

			// Check if subscription is invalid and should be removed
			if isInvalidSubscription(err) {
				result.InvalidSubscriptionIDs = append(result.InvalidSubscriptionIDs, sub.ID)
				// Remove invalid subscription
				if delErr := s.queries.DeletePushSubscriptionByID(ctx, sub.ID); delErr != nil {
					s.logger.Error("failed to delete invalid subscription", "id", sub.ID, "error", delErr)
				} else {
					s.logger.Info("deleted invalid subscription", "id", sub.ID)
				}
			}
		} else {
			result.SuccessfulDeliveries++
		}
	}

	return result
}

// sendToSubscription sends the payload to a single subscription
func (s *Service) sendToSubscription(sub *sqlc.PushSubscription, payloadBytes []byte) error {
	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dhKey,
			Auth:   sub.AuthKey,
		},
	}

	resp, err := webpush.SendNotification(payloadBytes, subscription, &webpush.Options{
		Subscriber:      s.vapidConfig.Subject,
		VAPIDPublicKey:  s.vapidConfig.PublicKey,
		VAPIDPrivateKey: s.vapidConfig.PrivateKey,
		TTL:             86400, // 24 hours
	})
	if err != nil {
		return fmt.Errorf("webpush send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push service returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// isInvalidSubscription checks if the error indicates the subscription is no longer valid
func isInvalidSubscription(err error) bool {
	// Check for common invalid subscription indicators
	// 404 Not Found or 410 Gone typically mean the subscription is expired
	errStr := err.Error()
	return contains(errStr, "404") || contains(errStr, "410") || contains(errStr, "status 404") || contains(errStr, "status 410")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SubscribeRequest represents a push subscription request from the frontend
type SubscribeRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// Subscribe creates or updates a push subscription for a user
func (s *Service) Subscribe(ctx context.Context, userID string, req SubscribeRequest, userAgent string) (*sqlc.PushSubscription, error) {
	// Check if subscription already exists
	existing, err := s.queries.GetPushSubscriptionByEndpoint(ctx, req.Endpoint)
	if err == nil && existing != nil {
		// Update existing subscription
		return s.queries.UpdatePushSubscription(ctx, sqlc.UpdatePushSubscriptionParams{
			Endpoint:  req.Endpoint,
			P256dhkey: req.Keys.P256dh,
			Authkey:   req.Keys.Auth,
			Useragent: userAgent,
		})
	}

	// Create new subscription
	return s.queries.CreatePushSubscription(ctx, sqlc.CreatePushSubscriptionParams{
		ID:        ulid.NewPushSubscriptionID(),
		Userid:    userID,
		Endpoint:  req.Endpoint,
		P256dhkey: req.Keys.P256dh,
		Authkey:   req.Keys.Auth,
		Useragent: userAgent,
	})
}

// Unsubscribe removes a push subscription
func (s *Service) Unsubscribe(ctx context.Context, endpoint string) error {
	return s.queries.DeletePushSubscriptionByEndpoint(ctx, endpoint)
}

// GetUserPreferences returns notification preferences for a user
func (s *Service) GetUserPreferences(ctx context.Context, userID string) ([]*sqlc.PushNotificationPreference, error) {
	return s.queries.GetUserNotificationPreferences(ctx, userID)
}

// SetPreference sets a notification preference for a user
func (s *Service) SetPreference(ctx context.Context, userID string, notificationType NotificationType, enabled bool) (*sqlc.PushNotificationPreference, error) {
	return s.queries.UpsertNotificationPreference(ctx, sqlc.UpsertNotificationPreferenceParams{
		Userid:           userID,
		Notificationtype: string(notificationType),
		Enabled:          enabled,
	})
}

// GetVAPIDPublicKey returns the VAPID public key for client-side subscription
func (s *Service) GetVAPIDPublicKey() string {
	return s.vapidConfig.PublicKey
}

// httpStatusFromError extracts HTTP status code from webpush error
func httpStatusFromError(err error) int {
	// webpush-go returns errors that contain the status code
	// This is a simple check - in production you might want more robust parsing
	if err == nil {
		return http.StatusOK
	}
	return 0
}
