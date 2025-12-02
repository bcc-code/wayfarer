package ulid

import (
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// Entity prefixes (2 characters each)
const (
	PrefixChurch               = "CH" // Churches
	PrefixUser                 = "US" // Users
	PrefixUserRole             = "UR" // User Roles
	PrefixProject              = "PR" // Projects
	PrefixEvent                = "EV" // Events
	PrefixSuperTeam            = "ST" // SuperTeams
	PrefixTeam                 = "TM" // Teams
	PrefixStreak               = "SK" // Streaks
	PrefixStreakRelevantDay    = "SD" // Streak Relevant Days
	PrefixChallenge            = "CL" // Challenges
	PrefixAchievement          = "AC" // Achievements
	PrefixReadingAchievement   = "RA" // Reading Achievement Articles
	PrefixListeningAchievement = "LT" // Listening Achievement Tracks
	PrefixScoreJournal         = "SJ" // Score Journal
	PrefixContentEvent         = "CE" // External Content Events
	PrefixConsent              = "CN" // Consents
	PrefixUserConsent          = "UC" // User Consent Acceptances (deprecated, use PrefixUserConsentHistory)
	PrefixUserConsentHistory   = "UH" // User Consent History
)

// Total ID length: 2 (prefix) + 26 (ULID) = 28 characters

// entropy is the source of randomness for ULID generation
var entropy = ulid.Monotonic(rand.Reader, 0)
var entropyMutex sync.Mutex

// newID generates a new ULID with the given prefix
// Uses a mutex to ensure thread-safe access to the entropy source
func newID(prefix string) string {
	if len(prefix) != 2 {
		panic(fmt.Sprintf("prefix must be exactly 2 characters, got %d", len(prefix)))
	}
	entropyMutex.Lock()
	defer entropyMutex.Unlock()
	return prefix + ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

// Type-safe ID generators for each entity type

// NewChurchID generates a new ID for a church (CH prefix)
func NewChurchID() string {
	return newID(PrefixChurch)
}

// NewUserID generates a new ID for a user (US prefix)
func NewUserID() string {
	return newID(PrefixUser)
}

// NewUserRoleID generates a new ID for a user role (UR prefix)
func NewUserRoleID() string {
	return newID(PrefixUserRole)
}

// NewProjectID generates a new ID for a project (PR prefix)
func NewProjectID() string {
	return newID(PrefixProject)
}

// NewEventID generates a new ID for an event (EV prefix)
func NewEventID() string {
	return newID(PrefixEvent)
}

// NewSuperTeamID generates a new ID for a super team (ST prefix)
func NewSuperTeamID() string {
	return newID(PrefixSuperTeam)
}

// NewTeamID generates a new ID for a team (TM prefix)
func NewTeamID() string {
	return newID(PrefixTeam)
}

// NewStreakID generates a new ID for a streak (SK prefix)
func NewStreakID() string {
	return newID(PrefixStreak)
}

// NewStreakRelevantDayID generates a new ID for a streak relevant day (SD prefix)
func NewStreakRelevantDayID() string {
	return newID(PrefixStreakRelevantDay)
}

// NewChallengeID generates a new ID for a challenge (CL prefix)
func NewChallengeID() string {
	return newID(PrefixChallenge)
}

// NewAchievementID generates a new ID for an achievement (AC prefix)
func NewAchievementID() string {
	return newID(PrefixAchievement)
}

// NewReadingAchievementID generates a new ID for a reading achievement article (RA prefix)
func NewReadingAchievementID() string {
	return newID(PrefixReadingAchievement)
}

// NewListeningAchievementID generates a new ID for a listening achievement track (LT prefix)
func NewListeningAchievementID() string {
	return newID(PrefixListeningAchievement)
}

// NewScoreJournalID generates a new ID for a score journal entry (SJ prefix)
func NewScoreJournalID() string {
	return newID(PrefixScoreJournal)
}

// NewContentEventID generates a new ID for an external content event (CE prefix)
func NewContentEventID() string {
	return newID(PrefixContentEvent)
}

// NewConsentID generates a new ID for a consent (CN prefix)
func NewConsentID() string {
	return newID(PrefixConsent)
}

// NewUserConsentID generates a new ID for a user consent acceptance (UC prefix)
// Deprecated: use NewUserConsentHistoryID instead
func NewUserConsentID() string {
	return newID(PrefixUserConsent)
}

// NewUserConsentHistoryID generates a new ID for user consent history (UH prefix)
func NewUserConsentHistoryID() string {
	return newID(PrefixUserConsentHistory)
}

// Validation functions

// IsValidID checks if an ID has the correct format and prefix
func IsValidID(id string, expectedPrefix string) bool {
	if len(id) != 28 {
		return false
	}
	if !strings.HasPrefix(id, expectedPrefix) {
		return false
	}
	// Try to parse the ULID portion (characters 2-28)
	_, err := ulid.Parse(id[2:])
	return err == nil
}

// GetPrefix extracts the 2-character prefix from an ID
func GetPrefix(id string) string {
	if len(id) < 2 {
		return ""
	}
	return id[:2]
}

// GetTimestamp extracts the timestamp from an ID
// Returns zero time if the ID is invalid
func GetTimestamp(id string) time.Time {
	if len(id) != 28 {
		return time.Time{}
	}
	parsed, err := ulid.Parse(id[2:])
	if err != nil {
		return time.Time{}
	}
	return ulid.Time(parsed.Time())
}

// Entity type validation functions

// IsChurchID validates a church ID
func IsChurchID(id string) bool {
	return IsValidID(id, PrefixChurch)
}

// IsUserID validates a user ID
func IsUserID(id string) bool {
	return IsValidID(id, PrefixUser)
}

// IsProjectID validates a project ID
func IsProjectID(id string) bool {
	return IsValidID(id, PrefixProject)
}

// IsEventID validates an event ID
func IsEventID(id string) bool {
	return IsValidID(id, PrefixEvent)
}

// IsSuperTeamID validates a super team ID
func IsSuperTeamID(id string) bool {
	return IsValidID(id, PrefixSuperTeam)
}

// IsTeamID validates a team ID
func IsTeamID(id string) bool {
	return IsValidID(id, PrefixTeam)
}

// IsStreakID validates a streak ID
func IsStreakID(id string) bool {
	return IsValidID(id, PrefixStreak)
}

// IsStreakRelevantDayID validates a streak relevant day ID
func IsStreakRelevantDayID(id string) bool {
	return IsValidID(id, PrefixStreakRelevantDay)
}

// IsChallengeID validates a challenge ID
func IsChallengeID(id string) bool {
	return IsValidID(id, PrefixChallenge)
}

// IsAchievementID validates an achievement ID
func IsAchievementID(id string) bool {
	return IsValidID(id, PrefixAchievement)
}

// IsReadingAchievementID validates a reading achievement ID
func IsReadingAchievementID(id string) bool {
	return IsValidID(id, PrefixReadingAchievement)
}

// IsListeningAchievementID validates a listening achievement ID
func IsListeningAchievementID(id string) bool {
	return IsValidID(id, PrefixListeningAchievement)
}

// IsScoreJournalID validates a score journal ID
func IsScoreJournalID(id string) bool {
	return IsValidID(id, PrefixScoreJournal)
}

// IsContentEventID validates a content event ID
func IsContentEventID(id string) bool {
	return IsValidID(id, PrefixContentEvent)
}

// IsConsentID validates a consent ID
func IsConsentID(id string) bool {
	return IsValidID(id, PrefixConsent)
}

// IsUserConsentID validates a user consent ID
// Deprecated: use IsUserConsentHistoryID instead
func IsUserConsentID(id string) bool {
	return IsValidID(id, PrefixUserConsent)
}

// IsUserConsentHistoryID validates a user consent history ID
func IsUserConsentHistoryID(id string) bool {
	return IsValidID(id, PrefixUserConsentHistory)
}
