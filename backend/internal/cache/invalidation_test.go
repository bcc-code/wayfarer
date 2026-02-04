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
		name   string
		key    string
		wantID string
		wantOK bool
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
		name   string
		key    string
		wantID string
		wantOK bool
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

func TestExtractUserTag(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		wantID string
		wantOK bool
	}{
		{
			name:   "direct user key",
			key:    "user:US123",
			wantID: "US123",
			wantOK: true,
		},
		{
			name:   "user projects key",
			key:    "userprojects:US123",
			wantID: "US123",
			wantOK: true,
		},
		{
			name:   "user events key",
			key:    "userevents:US123",
			wantID: "US123",
			wantOK: true,
		},
		{
			name:   "user content progress key",
			key:    "usercontent:US123:AC456",
			wantID: "US123",
			wantOK: true,
		},
		{
			name:   "user achievements key",
			key:    "userachievements:US123:AC456",
			wantID: "US123",
			wantOK: true,
		},
		{
			name:   "user streak activity key",
			key:    "userstreak:US123:SK789",
			wantID: "US123",
			wantOK: true,
		},
		{
			name:   "user challenge enrollments key",
			key:    "userchallengeenrollments:US123:CL456",
			wantID: "US123",
			wantOK: true,
		},
		{
			name:   "user challenge completions key",
			key:    "userchallenges:US123:CL456",
			wantID: "US123",
			wantOK: true,
		},
		{
			name:   "user consents key",
			key:    "userconsents:US123",
			wantID: "US123",
			wantOK: true,
		},
		{
			name:   "non-user key",
			key:    "challenge:CL123",
			wantID: "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := ExtractUserTag(tt.key)
			assert.Equal(t, tt.wantID, gotID)
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}

func TestCacheWithRegistry_UserInvalidation(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	assert.NoError(t, err)
	defer c.Close()

	userID := "US123"

	// Set various user-related cache entries
	c.Set(UserKey(userID), "user-data")
	c.Set(ProjectsByUserKey(userID), "projects")
	c.Set(EventsByUserKey(userID), "events")
	c.Set(UserRolesKey(userID), "roles")
	c.Set(UserContentProgressKey(userID, "AC001"), "progress-1")
	c.Set(UserContentProgressKey(userID, "AC002"), "progress-2")
	c.Set(UserAchievementTimestampKey(userID, "AC001"), "timestamp")
	c.Set(UserStreakActivityKey(userID, "SK001"), "streak")
	c.Set(UserChallengeEnrollmentKey(userID, "CL001"), "enrolled")
	c.Set(UserChallengeCompletionKey(userID, "CL001"), "completed")

	// Set a different user's data (should not be affected)
	c.Set(UserKey("US999"), "other-user")
	c.Set(UserContentProgressKey("US999", "AC001"), "other-progress")

	c.cache.Wait()

	// Verify all keys are set
	_, found := c.Get(UserKey(userID))
	assert.True(t, found, "user key should exist before invalidation")
	_, found = c.Get(UserContentProgressKey(userID, "AC001"))
	assert.True(t, found, "content progress should exist before invalidation")
	_, found = c.Get(UserAchievementTimestampKey(userID, "AC001"))
	assert.True(t, found, "achievement timestamp should exist before invalidation")

	// Invalidate user
	c.invalidateUserLocal(userID)
	c.cache.Wait()

	// Verify all user keys are deleted
	_, found = c.Get(UserKey(userID))
	assert.False(t, found, "user key should be deleted")
	_, found = c.Get(ProjectsByUserKey(userID))
	assert.False(t, found, "projects key should be deleted")
	_, found = c.Get(EventsByUserKey(userID))
	assert.False(t, found, "events key should be deleted")
	_, found = c.Get(UserRolesKey(userID))
	assert.False(t, found, "roles key should be deleted")
	_, found = c.Get(UserContentProgressKey(userID, "AC001"))
	assert.False(t, found, "content progress 1 should be deleted")
	_, found = c.Get(UserContentProgressKey(userID, "AC002"))
	assert.False(t, found, "content progress 2 should be deleted")
	_, found = c.Get(UserAchievementTimestampKey(userID, "AC001"))
	assert.False(t, found, "achievement timestamp should be deleted")
	_, found = c.Get(UserStreakActivityKey(userID, "SK001"))
	assert.False(t, found, "streak activity should be deleted")
	_, found = c.Get(UserChallengeEnrollmentKey(userID, "CL001"))
	assert.False(t, found, "challenge enrollment should be deleted")
	_, found = c.Get(UserChallengeCompletionKey(userID, "CL001"))
	assert.False(t, found, "challenge completion should be deleted")

	// Verify other user's data is NOT deleted
	_, found = c.Get(UserKey("US999"))
	assert.True(t, found, "other user's key should still exist")
	_, found = c.Get(UserContentProgressKey("US999", "AC001"))
	assert.True(t, found, "other user's content progress should still exist")
}

func TestCacheWithRegistry_TeamMemberLeaderboardInvalidationViaProject(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	assert.NoError(t, err)
	defer c.Close()

	teamID1 := "TM01AAAAAAAAAAAAAAAAAAAAAA01"
	teamID2 := "TM01AAAAAAAAAAAAAAAAAAAAAA02"

	// Set team member leaderboard entries for two teams
	c.Set(TeamMemberLeaderboardKey(teamID1), "leaderboard-team1")
	c.Set(TeamMemberLeaderboardKey(teamID2), "leaderboard-team2")
	c.Set(TeamMemberLeaderboardTeamLeadTagsKey(teamID1), "tags-team1")

	// Set an unrelated key that should not be affected
	c.Set(TeamKey(teamID1), "team-data")

	c.cache.Wait()

	// Verify keys exist
	_, found := c.Get(TeamMemberLeaderboardKey(teamID1))
	assert.True(t, found, "team1 leaderboard should exist before invalidation")
	_, found = c.Get(TeamMemberLeaderboardKey(teamID2))
	assert.True(t, found, "team2 leaderboard should exist before invalidation")
	_, found = c.Get(TeamMemberLeaderboardTeamLeadTagsKey(teamID1))
	assert.True(t, found, "team1 lead tags should exist before invalidation")

	// Simulate what invalidateProjectLocal does for team member leaderboards
	c.DeletePrefix(PrefixTeamMemberLeaderboard)
	c.cache.Wait()

	// All team member leaderboard keys should be deleted
	_, found = c.Get(TeamMemberLeaderboardKey(teamID1))
	assert.False(t, found, "team1 leaderboard should be deleted after project invalidation")
	_, found = c.Get(TeamMemberLeaderboardKey(teamID2))
	assert.False(t, found, "team2 leaderboard should be deleted after project invalidation")
	// Tags keys have a more specific prefix, so they are registered under PrefixTeamLeaderboardTags, not PrefixTeamMemberLeaderboard
	_, found = c.Get(TeamMemberLeaderboardTeamLeadTagsKey(teamID1))
	assert.True(t, found, "team1 lead tags should NOT be deleted by PrefixTeamMemberLeaderboard")

	// Unrelated team key should still exist
	_, found = c.Get(TeamKey(teamID1))
	assert.True(t, found, "team data should still exist")
}

func TestExtractPrefixes_TeamMemberLeaderboardKey(t *testing.T) {
	teamID := "TM01AAAAAAAAAAAAAAAAAAAAAA01"
	key := TeamMemberLeaderboardKey(teamID)

	// Key should be "team:leaderboard:TM01AAAAAAAAAAAAAAAAAAAAAA01"
	assert.Equal(t, "team:leaderboard:"+teamID, key)

	prefixes := extractPrefixes(key)
	assert.Contains(t, prefixes, PrefixTeamMemberLeaderboard, "should be registered under PrefixTeamMemberLeaderboard")
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
	cache.cache.Wait()

	// Verify all keys are set
	_, found1 := cache.Get(key1)
	_, found2 := cache.Get(key2)
	_, found3 := cache.Get(key3)
	assert.True(t, found1, "key1 should exist")
	assert.True(t, found2, "key2 should exist")
	assert.True(t, found3, "key3 should exist")

	// Invalidate only PROJ123 leaderboards using the prefix pattern
	cache.DeletePrefix("leaderboard:full:project:PROJ123")
	cache.cache.Wait()

	// Verify PROJ123 keys are deleted, but PROJ456 remains
	_, found1 = cache.Get(key1)
	_, found2 = cache.Get(key2)
	_, found3 = cache.Get(key3)
	assert.False(t, found1, "key1 should be deleted")
	assert.False(t, found2, "key2 should be deleted")
	assert.True(t, found3, "key3 should still exist")
}
