package ulid

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIDs(t *testing.T) {
	tests := []struct {
		name           string
		generateFunc   func() string
		expectedPrefix string
		validationFunc func(string) bool
	}{
		{"NewChurchID", NewChurchID, PrefixChurch, IsChurchID},
		{"NewUserID", NewUserID, PrefixUser, IsUserID},
		{"NewProjectID", NewProjectID, PrefixProject, IsProjectID},
		{"NewEventID", NewEventID, PrefixEvent, IsEventID},
		{"NewSuperTeamID", NewSuperTeamID, PrefixSuperTeam, IsSuperTeamID},
		{"NewTeamID", NewTeamID, PrefixTeam, IsTeamID},
		{"NewStreakID", NewStreakID, PrefixStreak, IsStreakID},
		{"NewChallengeID", NewChallengeID, PrefixChallenge, IsChallengeID},
		{"NewAchievementID", NewAchievementID, PrefixAchievement, IsAchievementID},
		{"NewContentItemID", NewContentItemID, PrefixContentItem, IsContentItemID},
		{"NewScoreJournalID", NewScoreJournalID, PrefixScoreJournal, IsScoreJournalID},
		{"NewUserFeedbackID", NewUserFeedbackID, PrefixUserFeedback, IsUserFeedbackID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.generateFunc()

			// Check length (2 prefix + 26 ULID = 28 total)
			assert.Equal(t, 28, len(id), "ID should be exactly 28 characters")

			// Check prefix
			assert.True(t, strings.HasPrefix(id, tt.expectedPrefix), "ID should have correct prefix")

			// Check validation function
			assert.True(t, tt.validationFunc(id), "ID should pass its validation function")

			// Generate multiple IDs and ensure they're unique
			id2 := tt.generateFunc()
			assert.NotEqual(t, id, id2, "Generated IDs should be unique")
		})
	}
}

func TestGetPrefix(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{"valid church ID", "CH01ARZ3NDEKTSV4RRFFQ69G5FAV", PrefixChurch},
		{"valid user ID", "US01ARZ3NDEKTSV4RRFFQ69G5FAV", PrefixUser},
		{"short ID", "X", ""},
		{"empty ID", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPrefix(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTimestamp(t *testing.T) {
	t.Run("extracts timestamp from valid ID", func(t *testing.T) {
		before := time.Now()
		id := NewUserID()
		after := time.Now()

		ts := GetTimestamp(id)
		assert.False(t, ts.IsZero(), "Timestamp should not be zero")
		assert.True(t, ts.After(before.Add(-time.Second)), "Timestamp should be after 'before'")
		assert.True(t, ts.Before(after.Add(time.Second)), "Timestamp should be before 'after'")
	})

	t.Run("returns zero time for invalid ID", func(t *testing.T) {
		ts := GetTimestamp("invalid")
		assert.True(t, ts.IsZero(), "Timestamp should be zero for invalid ID")
	})
}

func TestIsValidID(t *testing.T) {
	validID := NewUserID()

	tests := []struct {
		name           string
		id             string
		expectedPrefix string
		expected       bool
	}{
		{"valid ID with matching prefix", validID, PrefixUser, true},
		{"valid ID with wrong prefix", validID, PrefixChurch, false},
		{"too short", "US123", PrefixUser, false},
		{"too long", "US01ARZ3NDEKTSV4RRFFQ69G5FAVEXTRA", PrefixUser, false},
		{"invalid ULID portion", "USinvalid_ulid_portion_here", PrefixUser, false},
		{"empty string", "", PrefixUser, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidID(tt.id, tt.expectedPrefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationFunctions(t *testing.T) {
	churchID := NewChurchID()
	userID := NewUserID()

	t.Run("valid IDs pass their validation", func(t *testing.T) {
		assert.True(t, IsChurchID(churchID))
		assert.True(t, IsUserID(userID))
	})

	t.Run("IDs fail validation with wrong validator", func(t *testing.T) {
		assert.False(t, IsChurchID(userID))
		assert.False(t, IsUserID(churchID))
	})

	t.Run("invalid IDs fail all validations", func(t *testing.T) {
		invalid := "invalid"
		assert.False(t, IsChurchID(invalid))
		assert.False(t, IsUserID(invalid))
		assert.False(t, IsProjectID(invalid))
	})
}

func TestIDsAreLexicographicallySortable(t *testing.T) {
	// Generate IDs with small time delay
	id1 := NewUserID()
	time.Sleep(2 * time.Millisecond)
	id2 := NewUserID()

	// Later ID should be greater when compared as strings
	assert.True(t, id2 > id1, "Later ID should be lexicographically greater")
}

func TestConcurrentIDGeneration(t *testing.T) {
	// Test that concurrent ID generation produces unique IDs
	const numGoroutines = 100
	const idsPerGoroutine = 100

	idChan := make(chan string, numGoroutines*idsPerGoroutine)
	done := make(chan bool)

	// Generate IDs concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < idsPerGoroutine; j++ {
				idChan <- NewUserID()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(idChan)

	// Collect all IDs and check for uniqueness
	ids := make(map[string]bool)
	for id := range idChan {
		require.False(t, ids[id], "Duplicate ID generated: %s", id)
		ids[id] = true
	}

	assert.Equal(t, numGoroutines*idsPerGoroutine, len(ids), "All IDs should be unique")
}
