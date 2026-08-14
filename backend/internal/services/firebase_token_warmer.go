package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// Firebase custom tokens are valid for exactly one hour, so a warm entry must
// be replaced well before that. The cache TTL is deliberately longer than the
// refresh interval: a slow pass then leaves a still-valid token in place rather
// than a hole that the request path has to fill with an RSA signature.
const (
	tokenWarmInterval = 40 * time.Minute
	tokenWarmTTL      = 55 * time.Minute

	// First pass runs at this concurrency to get the cache warm quickly after a
	// restart; steady-state passes are paced instead (see pacedDelay).
	tokenWarmBootWorkers = 4

	// Steady-state passes spread minting across this fraction of the interval,
	// so signing never contends with a traffic burst.
	tokenWarmSpreadFraction = 0.5
)

// FirebaseTokenMinter is the subset of the Firebase service the warmer needs.
type FirebaseTokenMinter interface {
	CreateCustomToken(ctx context.Context, userID, churchID string) (string, error)
}

// TokenWarmerQuerier is the database surface the warmer needs.
type TokenWarmerQuerier interface {
	GetUserIDsInProject(ctx context.Context, projectid string) ([]string, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*sqlc.GetUsersByIDsRow, error)
}

// ProjectIDProvider resolves the project whose users should be kept warm.
type ProjectIDProvider interface {
	GetCurrentProjectID(ctx context.Context) (string, error)
}

// FirebaseTokenWarmer keeps a Firebase custom token cached for every user in
// the current project, refreshed on a staggered schedule.
//
// Why this exists: minting a custom token is a local RSA-2048 signature, which
// costs ~1ms of CPU. That is cheap once and ruinous in aggregate — a spike of
// 10,000 cold users hitting GetFirebaseToken inside a few seconds turns into
// thousands of signatures per second, and a CPU profile of exactly that spike
// attributed ~25% of all server CPU to bigmod/montgomeryMul. Rotating tokens in
// the background instead costs ~5 signatures per second at steady state and
// makes the request path a pure cache hit.
type FirebaseTokenWarmer struct {
	minter   FirebaseTokenMinter
	queries  TokenWarmerQuerier
	cache    *cache.CacheWithRegistry
	projects ProjectIDProvider

	interval time.Duration
	ttl      time.Duration

	cancel context.CancelFunc
	wg     sync.WaitGroup

	// minted counts successful signatures, for /metrics/cache style reporting.
	mu     sync.Mutex
	minted int
	warmed int
}

// NewFirebaseTokenWarmer builds a warmer. Returns nil when Firebase is not
// configured, so callers can wire it unconditionally.
func NewFirebaseTokenWarmer(
	minter FirebaseTokenMinter,
	queries TokenWarmerQuerier,
	c *cache.CacheWithRegistry,
	projects ProjectIDProvider,
) *FirebaseTokenWarmer {
	if minter == nil || queries == nil || c == nil || projects == nil {
		return nil
	}
	return &FirebaseTokenWarmer{
		minter:   minter,
		queries:  queries,
		cache:    c,
		projects: projects,
		interval: tokenWarmInterval,
		ttl:      tokenWarmTTL,
	}
}

// FirebaseTokenCacheKey is the cache key the firebaseToken resolver reads.
// Kept here so the warmer and the resolver cannot drift apart.
func FirebaseTokenCacheKey(userID string) string {
	return fmt.Sprintf("firebase_token:%s", userID)
}

// Start warms the cache once, then keeps it warm on a ticker.
func (w *FirebaseTokenWarmer) Start(ctx context.Context) {
	if w == nil {
		return
	}
	ctx, w.cancel = context.WithCancel(ctx)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		// Boot pass: get warm fast so an early spike is already covered.
		if n, err := w.warmAll(ctx, true); err != nil {
			slog.Warn("firebase token warmer: initial pass failed", "error", err)
		} else {
			slog.Info("firebase token warmer: initial pass complete", "tokens", n)
		}

		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if n, err := w.warmAll(ctx, false); err != nil {
					slog.Warn("firebase token warmer: refresh failed", "error", err)
				} else {
					slog.Debug("firebase token warmer: refreshed", "tokens", n)
				}
			}
		}
	}()
}

// Stop halts the warmer and waits for the goroutine to exit.
func (w *FirebaseTokenWarmer) Stop() {
	if w == nil {
		return
	}
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
}

// Stats reports cumulative signatures and the size of the last pass.
func (w *FirebaseTokenWarmer) Stats() (minted, warmed int) {
	if w == nil {
		return 0, 0
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.minted, w.warmed
}

// warmAll mints and caches a token for every user in the current project.
// When boot is true the pass runs concurrently; otherwise it is paced so
// signing stays in the background.
func (w *FirebaseTokenWarmer) warmAll(ctx context.Context, boot bool) (int, error) {
	projectID, err := w.projects.GetCurrentProjectID(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve current project: %w", err)
	}
	if projectID == "" {
		return 0, nil
	}

	userIDs, err := w.queries.GetUserIDsInProject(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to list project users: %w", err)
	}
	if len(userIDs) == 0 {
		return 0, nil
	}

	// church_id is a token claim, so fetch it in bulk rather than per user.
	churchByUser := make(map[string]string, len(userIDs))
	const batch = 1000
	for start := 0; start < len(userIDs); start += batch {
		end := min(start+batch, len(userIDs))
		rows, err := w.queries.GetUsersByIDs(ctx, userIDs[start:end])
		if err != nil {
			return 0, fmt.Errorf("failed to load users: %w", err)
		}
		for _, row := range rows {
			churchByUser[row.ID] = row.ChurchID
		}
	}

	w.mu.Lock()
	w.warmed = len(churchByUser)
	w.mu.Unlock()

	if boot {
		return w.mintConcurrent(ctx, churchByUser), nil
	}
	return w.mintPaced(ctx, churchByUser), nil
}

// mintConcurrent signs with bounded parallelism, for the boot pass.
func (w *FirebaseTokenWarmer) mintConcurrent(ctx context.Context, churchByUser map[string]string) int {
	type job struct{ userID, churchID string }
	jobs := make(chan job)

	var wg sync.WaitGroup
	var mu sync.Mutex
	count := 0

	for i := 0; i < tokenWarmBootWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if w.mintOne(ctx, j.userID, j.churchID) {
					mu.Lock()
					count++
					mu.Unlock()
				}
			}
		}()
	}

	for userID, churchID := range churchByUser {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return count
		case jobs <- job{userID, churchID}:
		}
	}
	close(jobs)
	wg.Wait()
	return count
}

// mintPaced spreads signing over a fraction of the refresh interval so it never
// competes with request traffic.
func (w *FirebaseTokenWarmer) mintPaced(ctx context.Context, churchByUser map[string]string) int {
	spread := time.Duration(float64(w.interval) * tokenWarmSpreadFraction)
	delay := spread / time.Duration(max(len(churchByUser), 1))

	count := 0
	for userID, churchID := range churchByUser {
		if w.mintOne(ctx, userID, churchID) {
			count++
		}
		if delay > 0 {
			select {
			case <-ctx.Done():
				return count
			case <-time.After(delay):
			}
		}
	}
	return count
}

// mintOne signs one token and stores it under the resolver's cache key.
func (w *FirebaseTokenWarmer) mintOne(ctx context.Context, userID, churchID string) bool {
	token, err := w.minter.CreateCustomToken(ctx, userID, churchID)
	if err != nil {
		slog.Debug("firebase token warmer: mint failed", "user_id", userID, "error", err)
		return false
	}
	w.cache.SetWithTTL(FirebaseTokenCacheKey(userID), &model.FirebaseTokenResponse{
		Token:     token,
		ExpiresIn: 3600,
	}, w.ttl)

	w.mu.Lock()
	w.minted++
	w.mu.Unlock()
	return true
}
