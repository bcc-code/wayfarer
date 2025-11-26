package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Key prefixes for different entity types
// These prefixes enable tag-based invalidation by matching prefix patterns
const (
	// Core entities
	PrefixUser        = "user:"
	PrefixChurch      = "church:"
	PrefixProject     = "project:"
	PrefixEvent       = "event:"
	PrefixTeam        = "team:"
	PrefixSuperTeam   = "superteam:"
	PrefixChallenge   = "challenge:"
	PrefixAchievement = "achievement:"
	PrefixStreak      = "streak:"

	// Relationship/Junction tables
	PrefixUserProjects = "userprojects:"
	PrefixUserEvents   = "userevents:"
	PrefixTeamMembers  = "teammembers:"
	PrefixUserRoles    = "userroles:"

	// Progress tracking
	PrefixUserAchievements         = "userachievements:"
	PrefixUserChallengeCompletions = "userchallenges:"
	PrefixUserReadingProgress      = "userreading:"
	PrefixUserListeningProgress    = "userlistening:"
	PrefixUserStreakActivity       = "userstreak:"

	// Computed data
	PrefixLeaderboard             = "leaderboard:"
	PrefixLeaderboardPosition     = "leaderboard:position:"
	PrefixLeaderboardCount        = "leaderboard:count:"
	PrefixTeamLeaderboardTags     = "team:leaderboard:tags:"
	PrefixScore                   = "score:"

	// Query results
	PrefixUsersFilter        = "usersfilter:"
	PrefixUsersCount         = "userscount:"
	PrefixProjectsFilter     = "projectsfilter:"
	PrefixProjectsCount      = "projectscount:"
	PrefixEventsFilter       = "eventsfilter:"
	PrefixEventsCount        = "eventscount:"
	PrefixTeamsFilter        = "teamsfilter:"
	PrefixTeamsCount         = "teamscount:"
	PrefixSuperTeamsFilter   = "superteamsfilter:"
	PrefixSuperTeamsCount    = "superteamscount:"
	PrefixAchievementsFilter = "achievementsfilter:"
	PrefixAchievementsCount  = "achievementscount:"
	PrefixChallengesFilter   = "challengesfilter:"
	PrefixChallengesCount    = "challengescount:"
	PrefixChurchesFilter     = "churchesfilter:"
	PrefixChurchesCount      = "churchescount:"
	PrefixStreaksFilter      = "streaksfilter:"
	PrefixStreaksCount       = "streakscount:"

	// Permissions/Roles
	PrefixHasRole          = "hasrole:"
	PrefixHasRoleInChurch  = "hasroleinchurch:"
	PrefixHasRoleInProject = "hasroleinproject:"
	PrefixHasRoleInTeam    = "hasroleinteam:"
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

// EventsByUserKey builds a cache key for events associated with a user
func EventsByUserKey(userID string) string {
	return PrefixUserEvents + userID
}

// TeamsByUserKey builds a cache key for teams associated with a user
func TeamsByUserKey(userID string) string {
	return PrefixTeamMembers + userID
}

// TeamMemberLeaderboardKey builds a cache key for team member leaderboard
func TeamMemberLeaderboardKey(teamID string) string {
	return fmt.Sprintf("%s:leaderboard:%s", PrefixTeam, teamID)
}

// TeamMemberLeaderboardTagsKey builds a cache key for user-specific leaderboard tags
func TeamMemberLeaderboardTagsKey(teamID, userID string) string {
	return fmt.Sprintf("%s%s:%s", PrefixTeamLeaderboardTags, teamID, userID)
}

// TeamsBySuperTeamKey builds a cache key for teams associated with a super team
func TeamsBySuperTeamKey(superTeamID string) string {
	return fmt.Sprintf("%s:superteam:%s", PrefixTeam, superTeamID)
}

// SuperTeamsByUserKey builds a cache key for super teams associated with a user
func SuperTeamsByUserKey(userID string) string {
	return fmt.Sprintf("%s%s", PrefixSuperTeam, userID)
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

// ArticlesByAchievementKey builds a cache key for articles by achievement ID
func ArticlesByAchievementKey(achievementID string) string {
	return fmt.Sprintf("%s:articles:%s", PrefixAchievement, achievementID)
}

// TracksByAchievementKey builds a cache key for tracks by achievement ID
func TracksByAchievementKey(achievementID string) string {
	return fmt.Sprintf("%s:tracks:%s", PrefixAchievement, achievementID)
}

// StreakKey builds a cache key for a streak by ID
func StreakKey(streakID string) string {
	return PrefixStreak + streakID
}

// StreaksByProjectKey builds a cache key for streaks in a project
func StreaksByProjectKey(projectID string) string {
	return fmt.Sprintf("%s:project:%s", PrefixStreak, projectID)
}

// RelevantDaysByStreakKey builds a cache key for relevant days by streak
func RelevantDaysByStreakKey(streakID string) string {
	return fmt.Sprintf("%s:relevant_days:%s", PrefixStreak, streakID)
}

// UserStreakActivityKey builds a cache key for user streak activity
func UserStreakActivityKey(userID string, streakID string) string {
	return fmt.Sprintf("%s%s:%s", PrefixUserStreakActivity, userID, streakID)
}

// UserAchievementTimestampKey builds a cache key for user achievement timestamp (achievedAt)
func UserAchievementTimestampKey(userID string, achievementID string) string {
	return fmt.Sprintf("%s%s:%s", PrefixUserAchievements, userID, achievementID)
}

// TeamMembersByTeamKey builds a cache key for team members
func TeamMembersByTeamKey(teamID string) string {
	return PrefixTeamMembers + teamID
}

// UsersByTeamKey builds a cache key for users in a team
func UsersByTeamKey(teamID string) string {
	return fmt.Sprintf("%s:team:%s", PrefixUser, teamID)
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

// UsersFilterKey builds a cache key for filtered users query results
// Takes a map of filter parameters and generates a deterministic hash
func UsersFilterKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixUsersFilter + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16] // Use first 16 chars of hash

	return PrefixUsersFilter + hashStr
}

// UsersCountKey builds a cache key for filtered users count query results
// Takes the same parameters as UsersFilterKey to ensure consistency
func UsersCountKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixUsersCount + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16] // Use first 16 chars of hash

	return PrefixUsersCount + hashStr
}

// HasRoleKey builds a cache key for HasRole checks
func HasRoleKey(userID string, role string) string {
	return fmt.Sprintf("%s%s:%s", PrefixHasRole, userID, role)
}

// HasRoleInChurchKey builds a cache key for HasRoleInChurch checks
func HasRoleInChurchKey(userID string, role string, churchID string) string {
	return fmt.Sprintf("%s%s:%s:%s", PrefixHasRoleInChurch, userID, role, churchID)
}

// HasRoleInProjectKey builds a cache key for HasRoleInProject checks
func HasRoleInProjectKey(userID string, role string, projectID string) string {
	return fmt.Sprintf("%s%s:%s:%s", PrefixHasRoleInProject, userID, role, projectID)
}

// HasRoleInTeamKey builds a cache key for HasRoleInTeam checks
func HasRoleInTeamKey(userID string, role string, teamID string) string {
	return fmt.Sprintf("%s%s:%s:%s", PrefixHasRoleInTeam, userID, role, teamID)
}

// ProjectsFilterKey builds a cache key for filtered projects query results
func ProjectsFilterKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixProjectsFilter + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixProjectsFilter + hashStr
}

// ProjectsCountKey builds a cache key for filtered projects count query results
func ProjectsCountKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixProjectsCount + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixProjectsCount + hashStr
}

// EventsFilterKey builds a cache key for filtered events query results
func EventsFilterKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixEventsFilter + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixEventsFilter + hashStr
}

// EventsCountKey builds a cache key for filtered events count query results
func EventsCountKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixEventsCount + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixEventsCount + hashStr
}

// TeamsFilterKey builds a cache key for filtered teams query results
func TeamsFilterKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixTeamsFilter + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixTeamsFilter + hashStr
}

// TeamsCountKey builds a cache key for filtered teams count query results
func TeamsCountKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixTeamsCount + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixTeamsCount + hashStr
}

// SuperTeamsFilterKey builds a cache key for filtered super teams query results
func SuperTeamsFilterKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixSuperTeamsFilter + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixSuperTeamsFilter + hashStr
}

// SuperTeamsCountKey builds a cache key for filtered super teams count query results
func SuperTeamsCountKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixSuperTeamsCount + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixSuperTeamsCount + hashStr
}

// AchievementsFilterKey builds a cache key for filtered achievements query results
func AchievementsFilterKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixAchievementsFilter + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixAchievementsFilter + hashStr
}

// AchievementsCountKey builds a cache key for filtered achievements count query results
func AchievementsCountKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixAchievementsCount + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixAchievementsCount + hashStr
}

func ChallengesFilterKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixChallengesFilter + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixChallengesFilter + hashStr
}

func ChallengesCountKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixChallengesCount + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixChallengesCount + hashStr
}

// ChurchesFilterKey builds a cache key for filtered churches query results
func ChurchesFilterKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixChurchesFilter + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixChurchesFilter + hashStr
}

// ChurchesCountKey builds a cache key for filtered churches count query results
func ChurchesCountKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixChurchesCount + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixChurchesCount + hashStr
}

// StreaksFilterKey builds a cache key for filtered streaks query results
func StreaksFilterKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixStreaksFilter + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixStreaksFilter + hashStr
}

// StreaksCountKey builds a cache key for filtered streaks count query results
func StreaksCountKey(params map[string]string) string {
	if len(params) == 0 {
		return PrefixStreaksCount + "all"
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for a shorter key
	hash := sha256.Sum256([]byte(builder.String()))
	hashStr := hex.EncodeToString(hash[:])[:16]

	return PrefixStreaksCount + hashStr
}

// LeaderboardKey builds a cache key for leaderboard query results (user-agnostic)
// context: "project" or "event"
// contextID: project ID or event ID
// entityType: "persons", "teams", "superteams", or "churches"
// params: filter parameters
// page: pagination cursor/offset identifier
func LeaderboardKey(context, contextID, entityType string, params map[string]string, page string) string {
	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for shorter key
	paramsHash := ""
	if builder.Len() > 0 {
		hash := sha256.Sum256([]byte(builder.String()))
		paramsHash = hex.EncodeToString(hash[:])[:16]
	} else {
		paramsHash = "all"
	}

	return fmt.Sprintf("%s%s:%s:%s:%s:%s", PrefixLeaderboard, context, contextID, entityType, paramsHash, page)
}

// LeaderboardPositionKey builds a cache key for user-specific leaderboard position
// context: "project" or "event"
// contextID: project ID or event ID
// entityType: "persons", "teams", "superteams", or "churches"
// params: filter parameters
// userID: current user's ID
func LeaderboardPositionKey(context, contextID, entityType string, params map[string]string, userID string) string {
	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for shorter key
	paramsHash := ""
	if builder.Len() > 0 {
		hash := sha256.Sum256([]byte(builder.String()))
		paramsHash = hex.EncodeToString(hash[:])[:16]
	} else {
		paramsHash = "all"
	}

	return fmt.Sprintf("%s%s:%s:%s:%s:%s", PrefixLeaderboardPosition, context, contextID, entityType, paramsHash, userID)
}

// LeaderboardCountKey builds a cache key for leaderboard total count (user-agnostic)
// context: "project" or "event"
// contextID: project ID or event ID
// entityType: "persons", "teams", "superteams", or "churches"
// params: filter parameters
func LeaderboardCountKey(context, contextID, entityType string, params map[string]string) string {
	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for shorter key
	paramsHash := ""
	if builder.Len() > 0 {
		hash := sha256.Sum256([]byte(builder.String()))
		paramsHash = hex.EncodeToString(hash[:])[:16]
	} else {
		paramsHash = "all"
	}

	return fmt.Sprintf("%s%s:%s:%s:%s", PrefixLeaderboardCount, context, contextID, entityType, paramsHash)
}

// FullLeaderboardKey builds a cache key for full leaderboard results (no pagination)
// context: "project" or "event"
// contextID: project ID or event ID
// entityType: "persons", "teams", "superteams", or "churches"
// params: filter parameters (includes all filters like churchId, gender, age, score range, etc.)
func FullLeaderboardKey(context, contextID, entityType string, params map[string]string) string {
	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build deterministic string from sorted key-value pairs
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(":")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// Hash the parameter string for shorter key
	paramsHash := ""
	if builder.Len() > 0 {
		hash := sha256.Sum256([]byte(builder.String()))
		paramsHash = hex.EncodeToString(hash[:])[:16]
	} else {
		paramsHash = "all"
	}

	return fmt.Sprintf("%sfull:%s:%s:%s:%s", PrefixLeaderboard, context, contextID, entityType, paramsHash)
}
