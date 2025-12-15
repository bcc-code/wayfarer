package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// CacheInvalidateChannel is the PostgreSQL NOTIFY channel for cache invalidation
	CacheInvalidateChannel = "cache_invalidate"
)

// InvalidationType represents the type of entity being invalidated
type InvalidationType string

const (
	InvalidationTypeUser        InvalidationType = "user"
	InvalidationTypeProject     InvalidationType = "project"
	InvalidationTypeEvent       InvalidationType = "event"
	InvalidationTypeTeam        InvalidationType = "team"
	InvalidationTypeSuperTeam   InvalidationType = "superteam"
	InvalidationTypeChallenge   InvalidationType = "challenge"
	InvalidationTypeAchievement InvalidationType = "achievement"
	InvalidationTypeQuiz        InvalidationType = "quiz"
	InvalidationTypeClear       InvalidationType = "clear"
)

// InvalidationMessage is the payload sent via NOTIFY
type InvalidationMessage struct {
	Type      InvalidationType `json:"t"`
	ID        string           `json:"id,omitempty"`
	ProjectID string           `json:"pid,omitempty"`
	EventID   string           `json:"eid,omitempty"`
}

// Executor is the interface for executing SQL statements (satisfied by pgxpool.Pool and pgx.Conn)
type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// CacheSync handles cross-instance cache invalidation via PostgreSQL NOTIFY/LISTEN
type CacheSync struct {
	cache      *CacheWithRegistry
	pooledURL  string
	instanceID string

	conn   *pgx.Conn
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu        sync.RWMutex
	connected bool
}

// NewCacheSync creates a new cache sync instance
// pooledURL is the standard (pooled) database URL - direct URL will be derived automatically
func NewCacheSync(cache *CacheWithRegistry, pooledURL string) *CacheSync {
	return &CacheSync{
		cache:      cache,
		pooledURL:  pooledURL,
		instanceID: ulid.NewInstanceID(),
	}
}

// InstanceID returns this instance's unique identifier
func (s *CacheSync) InstanceID() string {
	return s.instanceID
}

// Start begins listening for cache invalidation notifications
func (s *CacheSync) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	directURL := deriveDirectURL(s.pooledURL)

	conn, err := pgx.Connect(s.ctx, directURL)
	if err != nil {
		return err
	}
	s.conn = conn

	_, err = conn.Exec(s.ctx, "LISTEN "+CacheInvalidateChannel)
	if err != nil {
		conn.Close(s.ctx)
		return err
	}

	s.mu.Lock()
	s.connected = true
	s.mu.Unlock()

	slog.Info("cache sync started", "channel", CacheInvalidateChannel, "instance_id", s.instanceID)

	s.wg.Add(1)
	go s.listen()

	return nil
}

// Stop gracefully shuts down the cache sync listener
func (s *CacheSync) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()

	s.mu.Lock()
	s.connected = false
	s.mu.Unlock()

	if s.conn != nil {
		// Use background context since s.ctx is cancelled
		s.conn.Close(context.Background())
	}

	slog.Info("cache sync stopped", "instance_id", s.instanceID)
}

// BroadcastWithPool sends an invalidation message using a connection pool
func (s *CacheSync) BroadcastWithPool(ctx context.Context, pool *pgxpool.Pool, msg InvalidationMessage) {
	payload, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal invalidation message", "error", err)
		return
	}

	// Include instance ID to allow self-filtering
	fullPayload := s.instanceID + ":" + string(payload)

	_, err = pool.Exec(ctx, "SELECT pg_notify($1, $2)", CacheInvalidateChannel, fullPayload)
	if err != nil {
		slog.Error("failed to send cache invalidation notification", "error", err, "type", msg.Type)
	}
}

func (s *CacheSync) listen() {
	defer s.wg.Done()

	for {
		notification, err := s.conn.WaitForNotification(s.ctx)
		if err != nil {
			if s.ctx.Err() != nil {
				// Context cancelled, shutting down
				return
			}
			slog.Error("error waiting for notification", "error", err)

			// Try to reconnect
			s.reconnect()
			continue
		}

		s.handleNotification(notification.Payload)
	}
}

func (s *CacheSync) handleNotification(payload string) {
	// Parse instance ID prefix
	parts := strings.SplitN(payload, ":", 2)
	if len(parts) != 2 {
		slog.Warn("invalid notification payload format", "payload", payload)
		return
	}

	senderID := parts[0]
	jsonPayload := parts[1]

	// Skip messages from self
	if senderID == s.instanceID {
		return
	}

	var msg InvalidationMessage
	if err := json.Unmarshal([]byte(jsonPayload), &msg); err != nil {
		slog.Error("failed to unmarshal invalidation message", "error", err, "payload", jsonPayload)
		return
	}

	slog.Info("received cache invalidation", "type", msg.Type, "id", msg.ID, "from", senderID)

	s.applyInvalidation(msg)
}

func (s *CacheSync) applyInvalidation(msg InvalidationMessage) {
	switch msg.Type {
	case InvalidationTypeUser:
		s.cache.invalidateUserLocal(msg.ID)
	case InvalidationTypeProject:
		s.cache.invalidateProjectLocal(msg.ID)
	case InvalidationTypeEvent:
		s.cache.invalidateEventLocal(msg.ID)
	case InvalidationTypeTeam:
		s.cache.invalidateTeamLocal(msg.ID)
	case InvalidationTypeSuperTeam:
		s.cache.invalidateSuperTeamLocal(msg.ID)
	case InvalidationTypeChallenge:
		var eventID *string
		if msg.EventID != "" {
			eventID = &msg.EventID
		}
		s.cache.invalidateChallengeLocal(msg.ID, msg.ProjectID, eventID)
	case InvalidationTypeAchievement:
		s.cache.invalidateAchievementLocal(msg.ID)
	case InvalidationTypeQuiz:
		s.cache.invalidateQuizLocal(msg.ID)
	case InvalidationTypeClear:
		s.cache.Clear()
	default:
		slog.Warn("unknown invalidation type", "type", msg.Type)
	}
}

func (s *CacheSync) reconnect() {
	s.mu.Lock()
	s.connected = false
	s.mu.Unlock()

	if s.conn != nil {
		s.conn.Close(context.Background())
		s.conn = nil
	}

	// Exponential backoff reconnection
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(backoff):
		}

		directURL := deriveDirectURL(s.pooledURL)
		conn, err := pgx.Connect(s.ctx, directURL)
		if err != nil {
			slog.Warn("failed to reconnect for cache sync", "error", err, "retry_in", backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		_, err = conn.Exec(s.ctx, "LISTEN "+CacheInvalidateChannel)
		if err != nil {
			conn.Close(s.ctx)
			slog.Warn("failed to LISTEN after reconnect", "error", err)
			continue
		}

		s.conn = conn
		s.mu.Lock()
		s.connected = true
		s.mu.Unlock()

		slog.Info("cache sync reconnected", "instance_id", s.instanceID)

		// Clear cache on reconnect since we may have missed invalidations
		s.cache.Clear()
		return
	}
}

// deriveDirectURL converts a Neon pooled URL to a direct URL by removing "-pooler" from the host
func deriveDirectURL(pooledURL string) string {
	parsed, err := url.Parse(pooledURL)
	if err != nil {
		// If parsing fails, try simple string replacement
		return strings.Replace(pooledURL, "-pooler", "", 1)
	}

	// Remove "-pooler" from the host
	parsed.Host = strings.Replace(parsed.Host, "-pooler", "", 1)

	return parsed.String()
}
