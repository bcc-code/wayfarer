package cache

import (
	"fmt"
	"strings"
)

// Key prefixes for different entity types
// These prefixes enable tag-based invalidation by matching prefix patterns
const (
	// Core entities
	PrefixUser       = "user:"
	PrefixChurch     = "church:"
	PrefixProject    = "project:"
	PrefixEvent      = "event:"
	PrefixTeam       = "team:"
	PrefixSuperTeam  = "superteam:"
	PrefixChallenge  = "challenge:"
	PrefixAchievement = "achievement:"
	PrefixStreak     = "streak:"

	// Relationship/Junction tables
	PrefixUserProjects = "userprojects:"
	PrefixUserEvents   = "userevents:"
	PrefixTeamMembers  = "teammembers:"
	PrefixUserRoles    = "userroles:"

	// Progress tracking
	PrefixUserAchievements         = "userachievements:"
	PrefixTeamAchievements         = "teamachievements:"
	PrefixSuperTeamAchievements    = "superteamachievements:"
	PrefixUserChallengeCompletions = "userchallenges:"
	PrefixUserReadingProgress      = "userreading:"
	PrefixUserListeningProgress    = "userlistening:"
	PrefixUserStreakActivity       = "userstreak:"

	// Computed data (future use)
	PrefixLeaderboard = "leaderboard:"
	PrefixScore       = "score:"
)

// Key builders for different entity types

// UserKey builds a cache key for a user by ID
func UserKey(userID string) string {
	return PrefixUser + userID
}

// ChurchKey builds a cache key for a church by ID
func ChurchKey(churchID string) string {
	return PrefixChurch + churchID
}

// ProjectKey builds a cache key for a project by ID
func ProjectKey(projectID string) string {
	return PrefixProject + projectID
}

// ProjectsByUserKey builds a cache key for projects associated with a user
func ProjectsByUserKey(userID string) string {
	return PrefixUserProjects + userID
}

// EventKey builds a cache key for an event by ID
func EventKey(eventID string) string {
	return PrefixEvent + eventID
}

// EventsByProjectKey builds a cache key for events in a project
func EventsByProjectKey(projectID string) string {
	return fmt.Sprintf("%s:project:%s", PrefixEvent, projectID)
}

// TeamKey builds a cache key for a team by ID
func TeamKey(teamID string) string {
	return PrefixTeam + teamID
}

// TeamsByProjectKey builds a cache key for teams in a project
func TeamsByProjectKey(projectID string) string {
	return fmt.Sprintf("%s:project:%s", PrefixTeam, projectID)
}

// SuperTeamKey builds a cache key for a super team by ID
func SuperTeamKey(superTeamID string) string {
	return PrefixSuperTeam + superTeamID
}

// SuperTeamsByProjectKey builds a cache key for super teams in a project
func SuperTeamsByProjectKey(projectID string) string {
	return fmt.Sprintf("%s:project:%s", PrefixSuperTeam, projectID)
}

// ChallengeKey builds a cache key for a challenge by ID
func ChallengeKey(challengeID string) string {
	return PrefixChallenge + challengeID
}

// ChallengesByProjectKey builds a cache key for challenges in a project
func ChallengesByProjectKey(projectID string) string {
	return fmt.Sprintf("%s:project:%s", PrefixChallenge, projectID)
}

// ChallengesByEventKey builds a cache key for challenges in an event
func ChallengesByEventKey(eventID string) string {
	return fmt.Sprintf("%s:event:%s", PrefixChallenge, eventID)
}

// AchievementKey builds a cache key for an achievement by ID
func AchievementKey(achievementID string) string {
	return PrefixAchievement + achievementID
}

// AchievementsByProjectKey builds a cache key for achievements in a project
func AchievementsByProjectKey(projectID string) string {
	return fmt.Sprintf("%s:project:%s", PrefixAchievement, projectID)
}

// StreakKey builds a cache key for a streak by ID
func StreakKey(streakID string) string {
	return PrefixStreak + streakID
}

// StreaksByProjectKey builds a cache key for streaks in a project
func StreaksByProjectKey(projectID string) string {
	return fmt.Sprintf("%s:project:%s", PrefixStreak, projectID)
}

// TeamMembersByTeamKey builds a cache key for team members
func TeamMembersByTeamKey(teamID string) string {
	return PrefixTeamMembers + teamID
}

// UserRolesKey builds a cache key for user roles
func UserRolesKey(userID string) string {
	return PrefixUserRoles + userID
}

// Tag extraction helpers for invalidation

// ExtractProjectTag extracts project ID from a key for tag-based invalidation
func ExtractProjectTag(key string) (string, bool) {
	if strings.Contains(key, ":project:") {
		parts := strings.Split(key, ":project:")
		if len(parts) == 2 {
			return parts[1], true
		}
	}
	// Check if key is a direct project key
	if strings.HasPrefix(key, PrefixProject) {
		return strings.TrimPrefix(key, PrefixProject), true
	}
	return "", false
}

// ExtractEventTag extracts event ID from a key for tag-based invalidation
func ExtractEventTag(key string) (string, bool) {
	if strings.Contains(key, ":event:") {
		parts := strings.Split(key, ":event:")
		if len(parts) == 2 {
			return parts[1], true
		}
	}
	if strings.HasPrefix(key, PrefixEvent) {
		return strings.TrimPrefix(key, PrefixEvent), true
	}
	return "", false
}

// ExtractUserTag extracts user ID from a key for tag-based invalidation
func ExtractUserTag(key string) (string, bool) {
	// Check for direct user keys
	if strings.HasPrefix(key, PrefixUser) {
		return strings.TrimPrefix(key, PrefixUser), true
	}
	// Check for user-prefixed relationship keys
	prefixes := []string{PrefixUserProjects, PrefixUserEvents, PrefixUserRoles}
	for _, prefix := range prefixes {
		if strings.HasPrefix(key, prefix) {
			return strings.TrimPrefix(key, prefix), true
		}
	}
	return "", false
}
