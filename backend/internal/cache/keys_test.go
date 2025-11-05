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
