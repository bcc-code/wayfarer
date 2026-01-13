package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPrefixes_LeaderboardFullProject(t *testing.T) {
	// Test leaderboard:full:project:{projectID}:{entityType}:{paramsHash}
	key := "leaderboard:full:project:PROJ123:persons:all"
	prefixes := extractPrefixes(key)

	assert.Contains(t, prefixes, "leaderboard:full:project:PROJ123")
	assert.Len(t, prefixes, 1)
}

func TestExtractPrefixes_LeaderboardFullEvent(t *testing.T) {
	// Test leaderboard:full:event:{eventID}:{entityType}:{paramsHash}
	key := "leaderboard:full:event:EV456:teams:abc123"
	prefixes := extractPrefixes(key)

	assert.Contains(t, prefixes, "leaderboard:full:event:EV456")
	assert.Len(t, prefixes, 1)
}

func TestExtractPrefixes_LeaderboardProject(t *testing.T) {
	// Test leaderboard:project:{projectID}:{entityType}:{paramsHash}:{page}
	key := "leaderboard:project:PROJ123:persons:all:page1"
	prefixes := extractPrefixes(key)

	assert.Contains(t, prefixes, "leaderboard:project:PROJ123")
	assert.Len(t, prefixes, 1)
}

func TestExtractPrefixes_LeaderboardPositionProject(t *testing.T) {
	// Test leaderboard:position:project:{projectID}:{entityType}:{paramsHash}:{userID}
	key := "leaderboard:position:project:PROJ123:persons:all:USER001"
	prefixes := extractPrefixes(key)

	assert.Contains(t, prefixes, "leaderboard:position:project:PROJ123")
	assert.Len(t, prefixes, 1)
}

func TestExtractPrefixes_LeaderboardCountProject(t *testing.T) {
	// Test leaderboard:count:project:{projectID}:{entityType}:{paramsHash}
	key := "leaderboard:count:project:PROJ123:persons:all"
	prefixes := extractPrefixes(key)

	assert.Contains(t, prefixes, "leaderboard:count:project:PROJ123")
	assert.Len(t, prefixes, 1)
}

func TestExtractPrefixes_LeaderboardPositionEvent(t *testing.T) {
	// Test leaderboard:position:event:{eventID}:{entityType}:{paramsHash}:{userID}
	key := "leaderboard:position:event:EV789:teams:all:USER001"
	prefixes := extractPrefixes(key)

	assert.Contains(t, prefixes, "leaderboard:position:event:EV789")
	assert.Len(t, prefixes, 1)
}

func TestExtractPrefixes_LeaderboardCountEvent(t *testing.T) {
	// Test leaderboard:count:event:{eventID}:{entityType}:{paramsHash}
	key := "leaderboard:count:event:EV789:teams:all"
	prefixes := extractPrefixes(key)

	assert.Contains(t, prefixes, "leaderboard:count:event:EV789")
	assert.Len(t, prefixes, 1)
}

func TestExtractPrefixes_NonLeaderboardKey(t *testing.T) {
	// Test that non-leaderboard keys still work
	key := "user:US123"
	prefixes := extractPrefixes(key)

	assert.Contains(t, prefixes, PrefixUser)
}

func TestExtractPrefixes_ProjectKey(t *testing.T) {
	// Test project key extraction
	key := "challenge:project:PROJ123:CH456"
	prefixes := extractPrefixes(key)

	assert.Contains(t, prefixes, PrefixChallenge)
	assert.Contains(t, prefixes, "project:PROJ123")
}

func TestExtractPrefixes_EventKey(t *testing.T) {
	// Test event key extraction
	key := "challenge:event:EV789:CH456"
	prefixes := extractPrefixes(key)

	assert.Contains(t, prefixes, PrefixChallenge)
	assert.Contains(t, prefixes, "event:EV789")
}

func TestExtractProjectTag(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantID   string
		wantOK   bool
	}{
		{
			name:   "challenge with project",
			key:    "challenge:project:PROJ123:CH456",
			wantID: "PROJ123",
			wantOK: true,
		},
		{
			name:   "direct project key",
			key:    "project:PROJ123",
			wantID: "PROJ123",
			wantOK: true,
		},
		{
			name:   "no project in key",
			key:    "user:US123",
			wantID: "",
			wantOK: false,
		},
		{
			name:   "team with project",
			key:    "team:project:PROJ456",
			wantID: "PROJ456",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := ExtractProjectTag(tt.key)
			assert.Equal(t, tt.wantID, gotID)
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}

func TestExtractEventTag(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantID   string
		wantOK   bool
	}{
		{
			name:   "challenge with event",
			key:    "challenge:event:EV123:CH456",
			wantID: "EV123",
			wantOK: true,
		},
		{
			name:   "direct event key",
			key:    "event:EV123",
			wantID: "EV123",
			wantOK: true,
		},
		{
			name:   "no event in key",
			key:    "user:US123",
			wantID: "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := ExtractEventTag(tt.key)
			assert.Equal(t, tt.wantID, gotID)
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}

func TestCacheWithRegistry_LeaderboardInvalidation(t *testing.T) {
	cache, err := NewCacheWithRegistry(DefaultConfig())
	assert.NoError(t, err)
	defer cache.Close()

	// Set some leaderboard entries for project PROJ123
	key1 := FullLeaderboardKey("project", "PROJ123", "persons", nil)
	key2 := FullLeaderboardKey("project", "PROJ123", "teams", nil)
	key3 := FullLeaderboardKey("project", "PROJ456", "persons", nil)

	cache.Set(key1, []byte("data1"))
	cache.Set(key2, []byte("data2"))
	cache.Set(key3, []byte("data3"))

	// Wait for cache to process
	cache.Cache.cache.Wait()

	// Verify all keys are set
	_, found1 := cache.Get(key1)
	_, found2 := cache.Get(key2)
	_, found3 := cache.Get(key3)
	assert.True(t, found1, "key1 should exist")
	assert.True(t, found2, "key2 should exist")
	assert.True(t, found3, "key3 should exist")

	// Invalidate only PROJ123 leaderboards using the prefix pattern
	cache.DeletePrefix("leaderboard:full:project:PROJ123")
	cache.Cache.cache.Wait()

	// Verify PROJ123 keys are deleted, but PROJ456 remains
	_, found1 = cache.Get(key1)
	_, found2 = cache.Get(key2)
	_, found3 = cache.Get(key3)
	assert.False(t, found1, "key1 should be deleted")
	assert.False(t, found2, "key2 should be deleted")
	assert.True(t, found3, "key3 should still exist")
}
