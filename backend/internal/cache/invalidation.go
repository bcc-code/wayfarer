package cache

import (
	"sync"
	"time"
)

// KeyRegistry tracks all keys in the cache for tag-based bulk invalidation
// This is necessary because ristretto doesn't natively support prefix-based deletion
type KeyRegistry struct {
	mu   sync.RWMutex
	keys map[string][]string // prefix -> list of keys with that prefix
}

// NewKeyRegistry creates a new key registry
func NewKeyRegistry() *KeyRegistry {
	return &KeyRegistry{
		keys: make(map[string][]string),
	}
}

// Register adds a key to the registry under its prefixes
func (kr *KeyRegistry) Register(key string) {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	// Extract all relevant prefixes/tags from the key
	prefixes := extractPrefixes(key)
	for _, prefix := range prefixes {
		kr.keys[prefix] = append(kr.keys[prefix], key)
	}
}

// Unregister removes a key from the registry
func (kr *KeyRegistry) Unregister(key string) {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	prefixes := extractPrefixes(key)
	for _, prefix := range prefixes {
		keys := kr.keys[prefix]
		for i, k := range keys {
			if k == key {
				// Remove key from slice
				kr.keys[prefix] = append(keys[:i], keys[i+1:]...)
				break
			}
		}
	}
}

// GetKeys returns all keys matching the given prefix
func (kr *KeyRegistry) GetKeys(prefix string) []string {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	keys := kr.keys[prefix]
	result := make([]string, len(keys))
	copy(result, keys)
	return result
}

// Clear removes all keys from the registry
func (kr *KeyRegistry) Clear() {
	kr.mu.Lock()
	defer kr.mu.Unlock()
	kr.keys = make(map[string][]string)
}

// extractPrefixes extracts all relevant prefixes from a cache key
// For example, "challenge:project:PROJ123:CH456" would extract:
// - "challenge:" (entity type)
// - "challenge:project:PROJ123" (all challenges in project)
func extractPrefixes(key string) []string {
	prefixes := []string{}

	// Add the main entity prefix
	// Note: More specific prefixes must come before general ones (e.g., PrefixTeamLeaderboardTags before PrefixTeam)
	for _, prefix := range []string{
		PrefixTeamLeaderboardTags, // Must be before PrefixTeam
		PrefixUser, PrefixChurch, PrefixProject, PrefixEvent, PrefixTeam,
		PrefixSuperTeam, PrefixChallenge, PrefixAchievement, PrefixStreak,
		PrefixUserProjects, PrefixUserEvents, PrefixTeamMembers, PrefixUserRoles,
		PrefixUserChallengeEnrollments, PrefixUserChallengeCompletions,
		PrefixUsersFilter, PrefixUsersCount,
		PrefixProjectsFilter, PrefixProjectsCount,
		PrefixEventsFilter, PrefixEventsCount,
		PrefixTeamsFilter, PrefixTeamsCount,
		PrefixSuperTeamsFilter, PrefixSuperTeamsCount,
		PrefixAchievementsFilter, PrefixAchievementsCount,
		PrefixChallengesFilter, PrefixChallengesCount,
		PrefixChurchesFilter, PrefixChurchesCount,
		PrefixStreaksFilter, PrefixStreaksCount,
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
}

// NewCacheWithRegistry creates a cache with key registry support
func NewCacheWithRegistry(cfg Config) (*CacheWithRegistry, error) {
	cache, err := New(cfg)
	if err != nil {
		return nil, err
	}

	return &CacheWithRegistry{
		Cache:    cache,
		registry: NewKeyRegistry(),
	}, nil
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

// InvalidateUser invalidates all cache entries related to a user
func (c *CacheWithRegistry) InvalidateUser(userID string) {
	c.Delete(UserKey(userID))
	c.Delete(ProjectsByUserKey(userID))
	c.Delete(UserRolesKey(userID))
	c.DeletePrefix("user:" + userID)
	// Invalidate enrollment and completion data
	c.DeletePrefix(PrefixUserChallengeEnrollments + userID)
	c.DeletePrefix(PrefixUserChallengeCompletions + userID)
}

// InvalidateProject invalidates all cache entries related to a project
func (c *CacheWithRegistry) InvalidateProject(projectID string) {
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

	// Invalidate all team leaderboards in this project (scores changed)
	// Team leaderboard keys are "team:leaderboard:{teamID}"
	c.DeletePrefix("team:leaderboard:")
}

// InvalidateEvent invalidates all cache entries related to an event
func (c *CacheWithRegistry) InvalidateEvent(eventID string) {
	c.Delete(EventKey(eventID))
	c.DeletePrefix("event:" + eventID)

	// All event list/filter queries (any filter combination)
	// These are invalidated globally since filter query cache keys are hashed
	c.DeletePrefix(PrefixEventsFilter)
	c.DeletePrefix(PrefixEventsCount)
}

// InvalidateTeam invalidates all cache entries related to a team
func (c *CacheWithRegistry) InvalidateTeam(teamID string) {
	c.Delete(TeamKey(teamID))
	c.Delete(TeamMembersByTeamKey(teamID))
	c.Delete(TeamMemberLeaderboardKey(teamID))
	c.Delete(UsersByTeamKey(teamID))
	c.DeletePrefix("team:" + teamID)
}

// InvalidateSuperTeam invalidates all cache entries related to a super team
func (c *CacheWithRegistry) InvalidateSuperTeam(superTeamID string) {
	c.Delete(SuperTeamKey(superTeamID))
	c.Delete(TeamsBySuperTeamKey(superTeamID))
	c.DeletePrefix("superteam:" + superTeamID)
}

// InvalidateChallenge invalidates all cache entries related to a challenge
func (c *CacheWithRegistry) InvalidateChallenge(challengeID, projectID string, eventID *string) {
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
}

// InvalidateAchievement invalidates all cache entries related to an achievement
func (c *CacheWithRegistry) InvalidateAchievement(achievementID string) {
	c.Delete(AchievementKey(achievementID))

	// All achievement list/filter queries (any filter combination)
	// These are invalidated globally since filter query cache keys are hashed
	c.DeletePrefix(PrefixAchievementsFilter)
	c.DeletePrefix(PrefixAchievementsCount)
}

// InvalidateQuiz invalidates all cache entries related to a quiz
func (c *CacheWithRegistry) InvalidateQuiz(quizID string) {
	c.Delete(QuizKey(quizID))

	// All quiz list/filter queries
	c.DeletePrefix(PrefixQuizzesFilter)
	c.DeletePrefix(PrefixQuizzesCount)

	// Invalidate quiz questions and answers for this quiz
	c.Delete(QuizQuestionsByQuizKey(quizID))

	// Invalidate submissions for this quiz
	c.Delete(QuizSubmissionsByQuizKey(quizID))
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

// InvalidateTeamMemberLeaderboardTags invalidates all tag caches for team member leaderboards.
// Call this when user roles change, as TEAM_LEAD tags depend on role assignments.
// Note: TEAM_LEAD tags are cached per team (viewer-independent), ME tags are computed on-the-fly.
func (c *CacheWithRegistry) InvalidateTeamMemberLeaderboardTags() {
	c.DeletePrefix(PrefixTeamLeaderboardTags)
}
