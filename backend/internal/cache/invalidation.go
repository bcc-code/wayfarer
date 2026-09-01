package cache

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// KeyRegistry tracks all keys in the cache for tag-based bulk invalidation
// This is necessary because ristretto doesn't natively support prefix-based deletion
type KeyRegistry struct {
	mu   sync.RWMutex
	keys map[string]map[string]struct{} // prefix -> set of keys with that prefix
}

// NewKeyRegistry creates a new key registry
func NewKeyRegistry() *KeyRegistry {
	return &KeyRegistry{
		keys: make(map[string]map[string]struct{}),
	}
}

// Register adds a key to the registry under its prefixes. Registering the
// same key multiple times (e.g. concurrent cache misses) is idempotent.
func (kr *KeyRegistry) Register(key string) {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	// Extract all relevant prefixes/tags from the key
	prefixes := extractPrefixes(key)
	for _, prefix := range prefixes {
		set := kr.keys[prefix]
		if set == nil {
			set = make(map[string]struct{})
			kr.keys[prefix] = set
		}
		set[key] = struct{}{}
	}
}

// Unregister removes a key from the registry
func (kr *KeyRegistry) Unregister(key string) {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	prefixes := extractPrefixes(key)
	for _, prefix := range prefixes {
		set := kr.keys[prefix]
		delete(set, key)
		if len(set) == 0 {
			delete(kr.keys, prefix)
		}
	}
}

// GetKeys returns all keys matching the given prefix
func (kr *KeyRegistry) GetKeys(prefix string) []string {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	set := kr.keys[prefix]
	result := make([]string, 0, len(set))
	for key := range set {
		result = append(result, key)
	}
	return result
}

// Clear removes all keys from the registry
func (kr *KeyRegistry) Clear() {
	kr.mu.Lock()
	defer kr.mu.Unlock()
	kr.keys = make(map[string]map[string]struct{})
}

// extractPrefixes extracts all relevant prefixes from a cache key
// For example, "challenge:project:PROJ123:CH456" would extract:
// - "challenge:" (entity type)
// - "challenge:project:PROJ123" (all challenges in project)
func extractPrefixes(key string) []string {
	prefixes := []string{}

	// Whole-response GraphQL cache keys register under the umbrella prefix
	// (cleared on project invalidation) and, for per-user entries, under the
	// exact per-user prefix (cleared on that user's mutations/invalidation).
	if strings.HasPrefix(key, PrefixGQLResponse) {
		prefixes = append(prefixes, PrefixGQLResponse)
		if rest, ok := strings.CutPrefix(key, prefixGQLResponseUser); ok {
			if i := strings.IndexByte(rest, ':'); i > 0 {
				prefixes = append(prefixes, prefixGQLResponseUser+rest[:i+1])
			}
		}
		return prefixes
	}

	// Handle leaderboard keys specially - they need to be registered under
	// prefixes that match the invalidation patterns in InvalidateProject/InvalidateEvent
	// Key formats:
	// - leaderboard:full:project:{projectID}:{entityType}:{paramsHash}
	// - leaderboard:full:event:{eventID}:{entityType}:{paramsHash}
	// - leaderboard:project:{projectID}:{entityType}:{paramsHash}:{page}
	// - leaderboard:position:project:{projectID}:{entityType}:{paramsHash}:{userID}
	// - leaderboard:count:project:{projectID}:{entityType}:{paramsHash}
	if strings.HasPrefix(key, PrefixLeaderboard) {
		parts := strings.Split(key, ":")
		if len(parts) >= 4 {
			// Register under the base prefix for invalidation
			// e.g., "leaderboard:full:project:PROJ123" or "leaderboard:project:PROJ123"
			if parts[1] == "full" && len(parts) >= 5 {
				// leaderboard:full:{context}:{contextID}:...
				basePrefix := strings.Join(parts[:4], ":")
				prefixes = append(prefixes, basePrefix)
			} else if parts[1] == "position" || parts[1] == "count" {
				// leaderboard:position:{context}:{contextID}:... or leaderboard:count:{context}:{contextID}:...
				if len(parts) >= 5 {
					basePrefix := strings.Join(parts[:4], ":")
					prefixes = append(prefixes, basePrefix)
				}
			} else {
				// leaderboard:{context}:{contextID}:...
				basePrefix := strings.Join(parts[:3], ":")
				prefixes = append(prefixes, basePrefix)
			}
		}
		return prefixes
	}

	// Add the main entity prefix
	// Note: More specific prefixes must come before general ones (e.g., PrefixTeamLeaderboardTags before PrefixTeam)
	for _, prefix := range []string{
		PrefixTeamLeaderboardTags,   // Must be before PrefixTeamMemberLeaderboard and PrefixTeam
		PrefixTeamMemberLeaderboard, // Must be before PrefixTeam
		PrefixUser, PrefixChurch, PrefixProject, PrefixEvent, PrefixTeam,
		PrefixSuperTeam, PrefixChallenge, PrefixAchievement,
		PrefixUserProjects, PrefixUserEvents, PrefixTeamMembers, PrefixUserRoles,
		PrefixUserChallengeEnrollments, PrefixUserChallengeCompletions,
		PrefixUserContentProgress, PrefixUserAchievements, PrefixUserStreakProgress,
		PrefixUserConsents, PrefixUserProjectPoints, PrefixActiveChallengesCount,
		PrefixUserTeamInProject, PrefixUserEnrolledChallenges, PrefixUserQuizSessionAccess,
		PrefixUserActiveQuizSession,
		PrefixUsersFilter, PrefixUsersCount,
		PrefixProjectsFilter, PrefixProjectsCount,
		PrefixEventsFilter, PrefixEventsCount,
		PrefixTeamsFilter, PrefixTeamsCount,
		PrefixSuperTeamsFilter, PrefixSuperTeamsCount,
		PrefixAchievementsFilter, PrefixAchievementsCount,
		PrefixChallengesFilter, PrefixChallengesCount,
		PrefixChurchesFilter, PrefixChurchesCount,
	} {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			prefixes = append(prefixes, prefix)
			break
		}
	}

	// Add project/event/team tags if present in the key
	if projectID, ok := ExtractProjectTag(key); ok {
		prefixes = append(prefixes, "project:"+projectID)
	}
	if eventID, ok := ExtractEventTag(key); ok {
		prefixes = append(prefixes, "event:"+eventID)
	}
	if userID, ok := ExtractUserTag(key); ok {
		prefixes = append(prefixes, "user:"+userID)
	}

	return prefixes
}

// CacheWithRegistry extends Cache with key registry for tag-based invalidation
type CacheWithRegistry struct {
	*Cache
	registry *KeyRegistry
	sync     *CacheSync
	pool     *pgxpool.Pool

	// evictedKeys feeds the pruneEvictedKeys worker; closeOnce guards the
	// stop channel against double Close.
	evictedKeys chan string
	stop        chan struct{}
	closeOnce   sync.Once
}

// NewCacheWithRegistry creates a cache with key registry support
func NewCacheWithRegistry(cfg Config) (*CacheWithRegistry, error) {
	registry := NewKeyRegistry()

	// Prune keys from the registry when ristretto evicts them on its own (TTL
	// expiry / cost eviction / admission rejection). Without this the registry
	// grows unbounded and DeletePrefix slows down over time.
	// Ristretto fires OnReject not only for admission rejections but also for
	// a duplicate Set of a key it already holds (two concurrent cache misses
	// storing the same key: the second item is rejected while the first one's
	// value stays cached). Unregistering unconditionally would strand such a
	// live entry outside the registry, silently breaking prefix invalidation
	// for it — so keys must be re-checked against the cache before pruning.
	// The check cannot run inside the callback itself: ristretto invokes it
	// while holding internal shard locks, and a Get there deadlocks. The
	// callback only enqueues; pruneEvictedKeys does the check. A full queue
	// skips pruning (the registry then keeps a dead key until an explicit
	// Delete/DeletePrefix removes it), which is safe.
	evictedKeys := make(chan string, 4096)
	cfg.onEvictKey = func(key string) {
		select {
		case evictedKeys <- key:
		default:
		}
	}

	cache, err := New(cfg)
	if err != nil {
		return nil, err
	}

	c := &CacheWithRegistry{
		Cache:       cache,
		registry:    registry,
		evictedKeys: evictedKeys,
		stop:        make(chan struct{}),
	}
	go c.pruneEvictedKeys()
	return c, nil
}

// pruneEvictedKeys unregisters keys that ristretto evicted on its own, but
// only after confirming the key is really gone — an eviction callback can
// fire for an item whose key is still live (see NewCacheWithRegistry).
func (c *CacheWithRegistry) pruneEvictedKeys() {
	for {
		select {
		case key := <-c.evictedKeys:
			// Flush pending sets first so a just-stored value for this key
			// is visible to the Get (no-op once the cache is closed).
			c.Cache.Wait()
			if _, ok := c.Cache.Get(key); !ok {
				c.registry.Unregister(key)
			}
		case <-c.stop:
			return
		}
	}
}

// Close stops the registry pruning worker and closes the underlying cache.
func (c *CacheWithRegistry) Close() {
	c.closeOnce.Do(func() { close(c.stop) })
	c.Cache.Close()
}

// SetSync configures the cache sync for cross-instance invalidation
func (c *CacheWithRegistry) SetSync(sync *CacheSync, pool *pgxpool.Pool) {
	c.sync = sync
	c.pool = pool
}

// broadcast sends an invalidation message to other instances if sync is configured
func (c *CacheWithRegistry) broadcast(msg InvalidationMessage) {
	if c.sync != nil && c.pool != nil {
		c.sync.BroadcastWithPool(context.Background(), c.pool, msg)
	}
}

// Set stores a value and registers the key
func (c *CacheWithRegistry) Set(key string, value interface{}) bool {
	if ok := c.Cache.Set(key, value); ok {
		c.registry.Register(key)
		return true
	}
	return false
}

// SetWithTTL stores a value with custom TTL and registers the key
func (c *CacheWithRegistry) SetWithTTL(key string, value interface{}, ttl time.Duration) bool {
	if ok := c.Cache.SetWithTTL(key, value, ttl); ok {
		c.registry.Register(key)
		return true
	}
	return false
}

// Delete removes a value and unregisters the key
func (c *CacheWithRegistry) Delete(key string) {
	c.Cache.Delete(key)
	c.registry.Unregister(key)
}

// DeletePrefix removes all keys matching the given prefix
func (c *CacheWithRegistry) DeletePrefix(prefix string) {
	keys := c.registry.GetKeys(prefix)
	for _, key := range keys {
		c.Delete(key)
	}
}

// Clear removes all entries and clears the registry
func (c *CacheWithRegistry) Clear() {
	c.Cache.Clear()
	c.registry.Clear()
}

// Invalidation helper functions for common operations

// InvalidateUser invalidates all cache entries related to a user and broadcasts to other instances
func (c *CacheWithRegistry) InvalidateUser(userID string) {
	c.invalidateUserLocal(userID)
	c.broadcast(InvalidationMessage{Type: InvalidationTypeUser, ID: userID})
}

// invalidateUserLocal invalidates user cache entries on this instance only
func (c *CacheWithRegistry) invalidateUserLocal(userID string) {
	c.Delete(UserKey(userID))
	c.Delete(ProjectsByUserKey(userID))
	c.Delete(EventsByUserKey(userID))
	c.Delete(UserRolesKey(userID))
	c.DeletePrefix("user:" + userID)
	// Invalidate enrollment and completion data
	c.DeletePrefix(PrefixUserChallengeEnrollments + userID)
	c.DeletePrefix(PrefixUserChallengeCompletions + userID)
	// Invalidate content progress, achievements, and streak activity
	c.DeletePrefix(PrefixUserContentProgress + userID)
	c.DeletePrefix(PrefixUserAchievements + userID)
	c.DeletePrefix(PrefixUserStreakProgress + userID)
	// Invalidate user project points cache (myPoints field)
	c.DeletePrefix(PrefixUserProjectPoints + userID)
	// Invalidate active challenges count for this user
	c.DeletePrefix(PrefixActiveChallengesCount + userID)
	// Invalidate per-user lookup caches (myTeam, enrolled challenges, quiz session access)
	// These keys are registered under the "user:{userID}" tag, which the
	// DeletePrefix("user:"+userID) call above already covers.
	// Invalidate user filter/count queries (gender/church changes affect results)
	c.DeletePrefix(PrefixUsersFilter)
	c.DeletePrefix(PrefixUsersCount)
	// Drop the user's whole-response cache entries (per-user keyed)
	c.DeletePrefix(GQLResponseUserPrefix(userID))
}

// InvalidateProject invalidates all cache entries related to a project and broadcasts to other instances
func (c *CacheWithRegistry) InvalidateProject(projectID string) {
	c.invalidateProjectLocal(projectID)
	c.broadcast(InvalidationMessage{Type: InvalidationTypeProject, ID: projectID})
}

// invalidateProjectLocal invalidates project cache entries on this instance only
func (c *CacheWithRegistry) invalidateProjectLocal(projectID string) {
	// Direct project entity
	c.Delete(ProjectKey(projectID))

	// All related entities (events, teams, challenges, etc. by project)
	c.DeletePrefix("project:" + projectID)

	// All project list/filter queries (any filter combination)
	// These are invalidated globally since filter query cache keys are hashed
	// and don't contain the project ID
	c.DeletePrefix(PrefixProjectsFilter)
	c.DeletePrefix(PrefixProjectsCount)

	// All leaderboard data for this project
	// Leaderboard keys use pattern: leaderboard:{context}:{contextID}:...
	c.DeletePrefix("leaderboard:project:" + projectID)
	c.DeletePrefix("leaderboard:position:project:" + projectID)
	c.DeletePrefix("leaderboard:count:project:" + projectID)
	c.DeletePrefix("leaderboard:full:project:" + projectID)

	// Invalidate all team member leaderboards in this project (scores changed)
	c.DeletePrefix(PrefixTeamMemberLeaderboard)

	// Drop ALL whole-response cache entries: both shared and per-user entries
	// embed project data (branding, info message, leaderboards), so admin
	// project edits must surface immediately rather than after the TTL.
	c.DeletePrefix(PrefixGQLResponse)
}

// InvalidateEvent invalidates all cache entries related to an event and broadcasts to other instances
func (c *CacheWithRegistry) InvalidateEvent(eventID string) {
	c.invalidateEventLocal(eventID)
	c.broadcast(InvalidationMessage{Type: InvalidationTypeEvent, ID: eventID})
}

// invalidateEventLocal invalidates event cache entries on this instance only
func (c *CacheWithRegistry) invalidateEventLocal(eventID string) {
	c.Delete(EventKey(eventID))
	c.DeletePrefix("event:" + eventID)

	// All event list/filter queries (any filter combination)
	// These are invalidated globally since filter query cache keys are hashed
	c.DeletePrefix(PrefixEventsFilter)
	c.DeletePrefix(PrefixEventsCount)

	// All leaderboard data for this event
	// Leaderboard keys use pattern: leaderboard:{context}:{contextID}:...
	c.DeletePrefix("leaderboard:event:" + eventID)
	c.DeletePrefix("leaderboard:position:event:" + eventID)
	c.DeletePrefix("leaderboard:count:event:" + eventID)
	c.DeletePrefix("leaderboard:full:event:" + eventID)
}

// InvalidateTeam invalidates all cache entries related to a team and broadcasts to other instances
func (c *CacheWithRegistry) InvalidateTeam(teamID string) {
	c.invalidateTeamLocal(teamID)
	c.broadcast(InvalidationMessage{Type: InvalidationTypeTeam, ID: teamID})
}

// invalidateTeamLocal invalidates team cache entries on this instance only
func (c *CacheWithRegistry) invalidateTeamLocal(teamID string) {
	c.Delete(TeamKey(teamID))
	c.Delete(TeamMembersByTeamKey(teamID))
	c.Delete(TeamMemberLeaderboardKey(teamID))
	c.Delete(TeamMemberLeaderboardTeamLeadTagsKey(teamID))
	c.Delete(UsersByTeamKey(teamID))
	c.DeletePrefix("team:" + teamID)
}

// InvalidateSuperTeam invalidates all cache entries related to a super team and broadcasts to other instances
func (c *CacheWithRegistry) InvalidateSuperTeam(superTeamID string) {
	c.invalidateSuperTeamLocal(superTeamID)
	c.broadcast(InvalidationMessage{Type: InvalidationTypeSuperTeam, ID: superTeamID})
}

// invalidateSuperTeamLocal invalidates super team cache entries on this instance only
func (c *CacheWithRegistry) invalidateSuperTeamLocal(superTeamID string) {
	c.Delete(SuperTeamKey(superTeamID))
	c.Delete(TeamsBySuperTeamKey(superTeamID))
	c.DeletePrefix("superteam:" + superTeamID)
}

// InvalidateChallenge invalidates all cache entries related to a challenge and broadcasts to other instances
func (c *CacheWithRegistry) InvalidateChallenge(challengeID, projectID string, eventID *string) {
	c.invalidateChallengeLocal(challengeID, projectID, eventID)
	msg := InvalidationMessage{Type: InvalidationTypeChallenge, ID: challengeID, ProjectID: projectID}
	if eventID != nil {
		msg.EventID = *eventID
	}
	c.broadcast(msg)
}

// InvalidateUserChallengeEnrollment invalidates only the cache entries that one
// user's own enrollment change can affect.
//
// Use this for self-service enroll/unenroll instead of InvalidateChallenge.
// InvalidateChallenge exists for changes to the challenge *definition* and,
// because there is no reverse index from a challenge to its enrolled users, it
// falls back to DeletePrefix over five per-user prefixes — which discards those
// caches for every user in the system. That blast radius is fine when an admin
// edits a challenge and ruinous when 4,000 users self-enroll in the same second:
// each enrollment throws away the enrollment, completion, enrolled-challenge and
// quiz-access caches that the very next ChallengePage request needs, so the cache
// never survives long enough to serve anything. Measured on the ramped 10k spike,
// that turned ChallengePage into a guaranteed miss.
//
// A self-enrollment only changes state scoped to (userID, projectID, challengeID),
// and every one of those keys is directly addressable, so no prefix sweep is
// needed. Broadcasts the same narrow invalidation to other instances, so
// multi-replica deployments stay coherent without the InvalidateUser blast
// radius.
func (c *CacheWithRegistry) InvalidateUserChallengeEnrollment(userID, projectID, challengeID string) {
	c.invalidateUserChallengeEnrollmentLocal(userID, projectID, challengeID)
	c.broadcast(InvalidationMessage{
		Type: InvalidationTypeUserEnrollment, ID: userID,
		ProjectID: projectID, ChallengeID: challengeID,
	})
}

// invalidateUserChallengeEnrollmentLocal drops the enrollment-scoped keys on
// this instance only.
func (c *CacheWithRegistry) invalidateUserChallengeEnrollmentLocal(userID, projectID, challengeID string) {
	c.Delete(UserChallengeEnrollmentKey(userID, challengeID))
	c.Delete(UserChallengeCompletionKey(userID, challengeID))
	c.Delete(UserEnrolledChallengesKey(userID, projectID))
	c.Delete(UserQuizSessionAccessKey(userID, projectID))
	c.Delete(ActiveChallengesCountKey(userID, projectID))
	// The user's whole-response entries embed enrollment state (active
	// challenge lists, userEnrolledAt) and must not survive the enrollment.
	c.DeletePrefix(GQLResponseUserPrefix(userID))
}

// invalidateChallengeLocal invalidates challenge cache entries on this instance only
func (c *CacheWithRegistry) invalidateChallengeLocal(challengeID, projectID string, eventID *string) {
	c.Delete(ChallengeKey(challengeID))

	// Invalidate challenge list caches for project and event
	c.Delete(ChallengesByProjectKey(projectID))
	if eventID != nil {
		c.Delete(ChallengesByEventKey(*eventID))
	}

	// All challenge list/filter queries (any filter combination)
	// These are invalidated globally since filter query cache keys are hashed
	c.DeletePrefix(PrefixChallengesFilter)
	c.DeletePrefix(PrefixChallengesCount)

	// Invalidate enrollment and completion data for this challenge
	// This is more aggressive but necessary since we don't track reverse index
	c.DeletePrefix(PrefixUserChallengeEnrollments)
	c.DeletePrefix(PrefixUserChallengeCompletions)
	c.DeletePrefix(PrefixUserEnrolledChallenges)

	// Challenge changes can alter the project's quiz set, which the per-user
	// quiz session access cache is scoped to
	c.DeletePrefix(PrefixUserQuizSessionAccess)

	// Invalidate active challenges count (user-specific, keyed by project)
	c.DeletePrefix(PrefixActiveChallengesCount)
}

// InvalidateAchievement invalidates all cache entries related to an achievement and broadcasts to other instances
func (c *CacheWithRegistry) InvalidateAchievement(achievementID string) {
	c.invalidateAchievementLocal(achievementID)
	c.broadcast(InvalidationMessage{Type: InvalidationTypeAchievement, ID: achievementID})
}

// invalidateAchievementLocal invalidates achievement cache entries on this instance only
func (c *CacheWithRegistry) invalidateAchievementLocal(achievementID string) {
	c.Delete(AchievementKey(achievementID))

	// All achievement list/filter queries (any filter combination)
	// These are invalidated globally since filter query cache keys are hashed
	c.DeletePrefix(PrefixAchievementsFilter)
	c.DeletePrefix(PrefixAchievementsCount)
}

// InvalidateQuiz invalidates all cache entries related to a quiz and broadcasts to other instances
func (c *CacheWithRegistry) InvalidateQuiz(quizID string) {
	c.invalidateQuizLocal(quizID, "")
	c.broadcast(InvalidationMessage{Type: InvalidationTypeQuiz, ID: quizID})
}

// InvalidateQuizWithChallenge invalidates all cache entries related to a quiz including the challenge lookup cache
func (c *CacheWithRegistry) InvalidateQuizWithChallenge(quizID, challengeID string) {
	c.invalidateQuizLocal(quizID, challengeID)
	c.broadcast(InvalidationMessage{Type: InvalidationTypeQuiz, ID: quizID, ChallengeID: challengeID})
}

// invalidateQuizLocal invalidates quiz cache entries on this instance only
func (c *CacheWithRegistry) invalidateQuizLocal(quizID, challengeID string) {
	c.Delete(QuizKey(quizID))

	// Invalidate quiz by challenge lookup if challenge ID is provided
	if challengeID != "" {
		c.Delete(QuizByChallengeKey(challengeID))
	}

	// All quiz list/filter queries
	c.DeletePrefix(PrefixQuizzesFilter)
	c.DeletePrefix(PrefixQuizzesCount)

	// Per-user quiz session access and active-session caches may include this quiz
	c.DeletePrefix(PrefixUserQuizSessionAccess)
	c.DeletePrefix(PrefixUserActiveQuizSession)

	// Invalidate quiz questions and answers for this quiz
	c.Delete(QuizQuestionsByQuizKey(quizID))

	// Invalidate quiz achievement criteria for this quiz
	c.Delete(QuizAchievementsByQuizKey(quizID))

	// Invalidate submissions for this quiz
	c.Delete(QuizSubmissionsByQuizKey(quizID))
}

// InvalidateQuizSessionAccess invalidates all per-user quiz session access caches
// and broadcasts to other instances. Call this when session state changes
// (open/lock/finish/reopen/delete) or access is granted/revoked.
func (c *CacheWithRegistry) InvalidateQuizSessionAccess() {
	c.invalidateQuizSessionAccessLocal()
	c.broadcast(InvalidationMessage{Type: InvalidationTypeQuizSessionAccess})
}

// invalidateQuizSessionAccessLocal invalidates quiz session access caches on this instance only
func (c *CacheWithRegistry) invalidateQuizSessionAccessLocal() {
	c.DeletePrefix(PrefixUserQuizSessionAccess)
	c.DeletePrefix(PrefixUserActiveQuizSession)
}

// InvalidateQuizSession invalidates the cached session row and broadcasts to
// other instances. Call this whenever session state changes (open/lock/finish/
// reopen/update/delete), alongside InvalidateQuizSessionAccess for the
// per-user visibility caches.
func (c *CacheWithRegistry) InvalidateQuizSession(sessionID string) {
	c.invalidateQuizSessionLocal(sessionID)
	c.broadcast(InvalidationMessage{Type: InvalidationTypeQuizSession, ID: sessionID})
}

func (c *CacheWithRegistry) invalidateQuizSessionLocal(sessionID string) {
	c.Delete(QuizSessionKey(sessionID))
}

// InvalidateQuizAnswers invalidates cached answers/ordering items for a question
func (c *CacheWithRegistry) InvalidateQuizAnswers(questionID string) {
	c.Delete(QuizAnswersByQuestionKey(questionID))
}

// InvalidateQuizSubmission invalidates all cache entries related to a quiz submission
func (c *CacheWithRegistry) InvalidateQuizSubmission(submissionID string) {
	c.Delete(QuizSubmissionKey(submissionID))

	// Invalidate responses for this submission
	c.Delete(QuizResponsesBySubmissionKey(submissionID))

	// All quiz submission list/filter queries
	c.DeletePrefix(PrefixQuizSubmissionsFilter)
	c.DeletePrefix(PrefixQuizSubmissionsCount)
}

// InvalidateUserQuizSubmissions invalidates cached quiz submissions for a user
func (c *CacheWithRegistry) InvalidateUserQuizSubmissions(userID string) {
	c.Delete(QuizSubmissionsByUserKey(userID))
}

// InvalidateTeamMemberLeaderboardTags invalidates all tag caches for team member leaderboards.
// Call this when user roles change, as TEAM_LEAD tags depend on role assignments.
// Note: TEAM_LEAD tags are cached per team (viewer-independent), ME tags are computed on-the-fly.
func (c *CacheWithRegistry) InvalidateTeamMemberLeaderboardTags() {
	c.DeletePrefix(PrefixTeamLeaderboardTags)
}

// Hits returns the total number of cache hits
func (c *CacheWithRegistry) Hits() uint64 {
	return c.Cache.Metrics().Hits()
}
