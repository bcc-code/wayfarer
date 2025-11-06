package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersFilterKey_EmptyParams(t *testing.T) {
	params := make(map[string]string)
	key := UsersFilterKey(params)

	assert.Equal(t, "usersfilter:all", key)
}

func TestUsersFilterKey_SingleParam(t *testing.T) {
	params := map[string]string{
		"churchid": "CH123",
	}
	key := UsersFilterKey(params)

	assert.Contains(t, key, "usersfilter:")
	assert.NotEqual(t, "usersfilter:all", key)
	assert.Len(t, key, len("usersfilter:")+16) // prefix + 16 char hash
}

func TestUsersFilterKey_MultipleParams(t *testing.T) {
	params := map[string]string{
		"churchid":  "CH123",
		"gender":    "MALE",
		"minage":    "18",
		"maxage":    "30",
		"projectid": "PR456",
	}
	key := UsersFilterKey(params)

	assert.Contains(t, key, "usersfilter:")
	assert.Len(t, key, len("usersfilter:")+16) // prefix + 16 char hash
}

func TestUsersFilterKey_Deterministic(t *testing.T) {
	params1 := map[string]string{
		"churchid": "CH123",
		"gender":   "MALE",
		"minage":   "18",
	}
	params2 := map[string]string{
		"minage":   "18",
		"churchid": "CH123",
		"gender":   "MALE",
	}

	key1 := UsersFilterKey(params1)
	key2 := UsersFilterKey(params2)

	// Keys should be identical even though map order may differ
	assert.Equal(t, key1, key2, "Same parameters should produce same cache key regardless of order")
}

func TestUsersFilterKey_DifferentParams(t *testing.T) {
	params1 := map[string]string{
		"churchid": "CH123",
		"gender":   "MALE",
	}
	params2 := map[string]string{
		"churchid": "CH123",
		"gender":   "FEMALE",
	}

	key1 := UsersFilterKey(params1)
	key2 := UsersFilterKey(params2)

	assert.NotEqual(t, key1, key2, "Different parameters should produce different cache keys")
}

func TestUsersFilterKey_WithPaginationParams(t *testing.T) {
	params1 := map[string]string{
		"churchid": "CH123",
		"first":    "10",
		"after":    "US001",
	}
	params2 := map[string]string{
		"churchid": "CH123",
		"first":    "20",
		"after":    "US001",
	}

	key1 := UsersFilterKey(params1)
	key2 := UsersFilterKey(params2)

	// Different pagination should produce different keys
	assert.NotEqual(t, key1, key2, "Different pagination parameters should produce different cache keys")
}

func TestUsersCountKey_EmptyParams(t *testing.T) {
	params := make(map[string]string)
	key := UsersCountKey(params)

	assert.Equal(t, "userscount:all", key)
}

func TestUsersCountKey_SingleParam(t *testing.T) {
	params := map[string]string{
		"churchid": "CH123",
	}
	key := UsersCountKey(params)

	assert.Contains(t, key, "userscount:")
	assert.NotEqual(t, "userscount:all", key)
	assert.Len(t, key, len("userscount:")+16) // prefix + 16 char hash
}

func TestUsersCountKey_Deterministic(t *testing.T) {
	params1 := map[string]string{
		"churchid": "CH123",
		"gender":   "MALE",
		"minage":   "18",
	}
	params2 := map[string]string{
		"minage":   "18",
		"churchid": "CH123",
		"gender":   "MALE",
	}

	key1 := UsersCountKey(params1)
	key2 := UsersCountKey(params2)

	// Keys should be identical even though map order may differ
	assert.Equal(t, key1, key2, "Same parameters should produce same cache key regardless of order")
}

func TestUsersCountKey_DifferentFromFilterKey(t *testing.T) {
	params := map[string]string{
		"churchid": "CH123",
		"gender":   "MALE",
	}

	filterKey := UsersFilterKey(params)
	countKey := UsersCountKey(params)

	// Filter and count keys should have different prefixes
	assert.Contains(t, filterKey, "usersfilter:")
	assert.Contains(t, countKey, "userscount:")
	assert.NotEqual(t, filterKey, countKey, "Filter and count keys should be different")
}

func TestUsersCountKey_IgnoresPaginationParams(t *testing.T) {
	// Count shouldn't change based on pagination, but we include pagination in the key
	// because we want different cache entries for different result sets
	params1 := map[string]string{
		"churchid": "CH123",
		"first":    "10",
	}
	params2 := map[string]string{
		"churchid": "CH123",
		"first":    "20",
	}

	key1 := UsersCountKey(params1)
	key2 := UsersCountKey(params2)

	// Different params should produce different keys (even though count might be the same)
	assert.NotEqual(t, key1, key2)
}

func TestCacheKeyConsistency(t *testing.T) {
	// Test that calling the same function multiple times produces the same result
	params := map[string]string{
		"churchid":  "CH123",
		"gender":    "MALE",
		"minage":    "18",
		"maxage":    "30",
		"projectid": "PR456",
		"eventid":   "EV789",
		"teamid":    "TM999",
	}

	// Generate keys multiple times
	keys := make([]string, 10)
	for i := 0; i < 10; i++ {
		keys[i] = UsersFilterKey(params)
	}

	// All keys should be identical
	for i := 1; i < len(keys); i++ {
		assert.Equal(t, keys[0], keys[i], "Cache key generation should be consistent")
	}
}

func TestCacheKeyLength(t *testing.T) {
	// Test with extremely long parameter values
	params := map[string]string{
		"churchid":  "CH" + string(make([]byte, 1000)),
		"projectid": "PR" + string(make([]byte, 1000)),
	}

	key := UsersFilterKey(params)

	// Key should still be reasonable length (hash keeps it short)
	assert.Less(t, len(key), 50, "Cache key should remain short even with long parameters")
}

func TestSuperTeamsFilterKey_EmptyParams(t *testing.T) {
	params := make(map[string]string)
	key := SuperTeamsFilterKey(params)

	assert.Equal(t, "superteamsfilter:all", key)
}

func TestSuperTeamsFilterKey_SingleParam(t *testing.T) {
	params := map[string]string{
		"projectid": "PR123",
	}
	key := SuperTeamsFilterKey(params)

	assert.Contains(t, key, "superteamsfilter:")
	assert.NotEqual(t, "superteamsfilter:all", key)
	assert.Len(t, key, len("superteamsfilter:")+16) // prefix + 16 char hash
}

func TestSuperTeamsFilterKey_MultipleParams(t *testing.T) {
	params := map[string]string{
		"projectid":  "PR123",
		"minteams":   "2",
		"maxteams":   "10",
		"minmembers": "15",
		"maxmembers": "100",
	}
	key := SuperTeamsFilterKey(params)

	assert.Contains(t, key, "superteamsfilter:")
	assert.Len(t, key, len("superteamsfilter:")+16) // prefix + 16 char hash
}

func TestSuperTeamsFilterKey_Deterministic(t *testing.T) {
	params1 := map[string]string{
		"projectid": "PR123",
		"minteams":  "2",
		"maxteams":  "10",
	}
	params2 := map[string]string{
		"maxteams":  "10",
		"projectid": "PR123",
		"minteams":  "2",
	}

	key1 := SuperTeamsFilterKey(params1)
	key2 := SuperTeamsFilterKey(params2)

	// Keys should be identical even though map order may differ
	assert.Equal(t, key1, key2, "Same parameters should produce same cache key regardless of order")
}

func TestSuperTeamsFilterKey_DifferentParams(t *testing.T) {
	params1 := map[string]string{
		"projectid": "PR123",
		"minteams":  "2",
	}
	params2 := map[string]string{
		"projectid": "PR123",
		"minteams":  "3",
	}

	key1 := SuperTeamsFilterKey(params1)
	key2 := SuperTeamsFilterKey(params2)

	assert.NotEqual(t, key1, key2, "Different parameters should produce different cache keys")
}

func TestSuperTeamsFilterKey_WithPaginationParams(t *testing.T) {
	params1 := map[string]string{
		"projectid": "PR123",
		"first":     "10",
		"after":     "ST001",
	}
	params2 := map[string]string{
		"projectid": "PR123",
		"first":     "20",
		"after":     "ST001",
	}

	key1 := SuperTeamsFilterKey(params1)
	key2 := SuperTeamsFilterKey(params2)

	// Different pagination should produce different keys
	assert.NotEqual(t, key1, key2, "Different pagination parameters should produce different cache keys")
}

func TestSuperTeamsFilterKey_WithIdsParam(t *testing.T) {
	params := map[string]string{
		"ids": "[ST001 ST002 ST003]",
	}
	key := SuperTeamsFilterKey(params)

	assert.Contains(t, key, "superteamsfilter:")
	assert.NotEqual(t, "superteamsfilter:all", key)
}

func TestSuperTeamsCountKey_EmptyParams(t *testing.T) {
	params := make(map[string]string)
	key := SuperTeamsCountKey(params)

	assert.Equal(t, "superteamscount:all", key)
}

func TestSuperTeamsCountKey_SingleParam(t *testing.T) {
	params := map[string]string{
		"projectid": "PR123",
	}
	key := SuperTeamsCountKey(params)

	assert.Contains(t, key, "superteamscount:")
	assert.NotEqual(t, "superteamscount:all", key)
	assert.Len(t, key, len("superteamscount:")+16) // prefix + 16 char hash
}

func TestSuperTeamsCountKey_Deterministic(t *testing.T) {
	params1 := map[string]string{
		"projectid":  "PR123",
		"minteams":   "2",
		"minmembers": "15",
	}
	params2 := map[string]string{
		"minmembers": "15",
		"projectid":  "PR123",
		"minteams":   "2",
	}

	key1 := SuperTeamsCountKey(params1)
	key2 := SuperTeamsCountKey(params2)

	// Keys should be identical even though map order may differ
	assert.Equal(t, key1, key2, "Same parameters should produce same cache key regardless of order")
}

func TestSuperTeamsCountKey_DifferentFromFilterKey(t *testing.T) {
	params := map[string]string{
		"projectid": "PR123",
		"minteams":  "2",
	}

	filterKey := SuperTeamsFilterKey(params)
	countKey := SuperTeamsCountKey(params)

	// Filter and count keys should have different prefixes
	assert.Contains(t, filterKey, "superteamsfilter:")
	assert.Contains(t, countKey, "superteamscount:")
	assert.NotEqual(t, filterKey, countKey, "Filter and count keys should be different")
}

func TestSuperTeamsCountKey_WithAllFilters(t *testing.T) {
	params := map[string]string{
		"projectid":  "PR123",
		"ids":        "[ST001 ST002]",
		"minteams":   "2",
		"maxteams":   "10",
		"minmembers": "15",
		"maxmembers": "100",
		"first":      "10",
	}

	key := SuperTeamsCountKey(params)

	assert.Contains(t, key, "superteamscount:")
	assert.Len(t, key, len("superteamscount:")+16) // prefix + 16 char hash
}

func TestSuperTeamsKeyConsistency(t *testing.T) {
	// Test that calling the same function multiple times produces the same result
	params := map[string]string{
		"projectid":  "PR123",
		"minteams":   "2",
		"maxteams":   "10",
		"minmembers": "15",
		"maxmembers": "100",
	}

	// Generate keys multiple times
	keys := make([]string, 10)
	for i := 0; i < 10; i++ {
		keys[i] = SuperTeamsFilterKey(params)
	}

	// All keys should be identical
	for i := 1; i < len(keys); i++ {
		assert.Equal(t, keys[0], keys[i], "Cache key generation should be consistent")
	}
}

func TestSuperTeamsKeyLength(t *testing.T) {
	// Test with extremely long parameter values
	params := map[string]string{
		"projectid": "PR" + string(make([]byte, 1000)),
		"ids":       string(make([]byte, 2000)),
	}

	key := SuperTeamsFilterKey(params)

	// Key should still be reasonable length (hash keeps it short)
	assert.Less(t, len(key), 50, "Cache key should remain short even with long parameters")
}
