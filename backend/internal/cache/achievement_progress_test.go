package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAchievementProgressInvalidation(t *testing.T) {
	for _, remote := range []bool{false, true} {
		name := "local"
		if remote {
			name = "remote notification"
		}
		t.Run(name, func(t *testing.T) {
			c, err := NewCacheWithRegistry(DefaultConfig())
			require.NoError(t, err)
			defer c.Close()
			removed := []string{
				UserContentProgressKey("US1", "AC1"),
				UserContentProgressKey("US1", "AC2"),
				UserStreakProgressKey("US1", "AC1"),
				UserAchievementTimestampKey("US1", "AC1"),
				GQLResponseUserPrefix("US1") + "query",
			}
			retained := []string{
				UserContentProgressKey("US2", "AC1"),
				UserStreakProgressKey("US2", "AC1"),
				UserAchievementTimestampKey("US2", "AC1"),
				GQLResponseUserPrefix("US2") + "query",
				PrefixGQLResponseShared + "query",
				UserRolesKey("US1"), UserKey("US1"), AchievementKey("AC1"),
			}
			for _, key := range append(removed, retained...) {
				require.True(t, c.Set(key, "cached"))
			}
			c.Wait()
			for _, key := range removed {
				_, ok := c.Get(key)
				require.True(t, ok, "precondition: %s", key)
			}
			if remote {
				sync := NewCacheSync(c, "")
				sync.handleNotification(`another-instance:{"t":"userachievementprogress","id":"US1"}`)
			} else {
				c.InvalidateUserAchievementProgress("US1")
			}
			for _, key := range removed {
				_, ok := c.Get(key)
				assert.False(t, ok, "invalidate %s", key)
			}
			for _, key := range retained {
				_, ok := c.Get(key)
				assert.True(t, ok, "retain %s", key)
			}
		})
	}
}

func TestAchievementItemInvalidationAcrossInstances(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	defer c.Close()
	keys := []string{
		AchievementKey("AC1"), ContentItemsByAchievementKey("AC1"),
		StreakItemsByAchievementKey("AC1"), ContentItemCountKey("AC1"),
	}
	for _, key := range keys {
		require.True(t, c.Set(key, "cached"))
	}
	require.True(t, c.Set(StreakItemsByAchievementKey("AC2"), "unrelated"))
	c.Wait()
	for _, key := range keys {
		_, ok := c.Get(key)
		require.True(t, ok, "precondition: %s", key)
	}
	sync := NewCacheSync(c, "")
	sync.handleNotification(`another-instance:{"t":"achievement","id":"AC1"}`)
	for _, key := range keys {
		_, ok := c.Get(key)
		assert.False(t, ok, "invalidate %s", key)
	}
	_, ok := c.Get(StreakItemsByAchievementKey("AC2"))
	assert.True(t, ok)
}
