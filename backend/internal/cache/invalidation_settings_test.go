package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidateSettings_RunsRegisteredRefresher(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	defer c.Close()

	calls := 0
	c.SetSettingsRefresher(func() { calls++ })

	c.InvalidateSettings()

	assert.Equal(t, 1, calls)
}

// Without sync configured there is no refresher either; invalidating must be a
// no-op rather than a nil dereference.
func TestInvalidateSettings_WithoutRefresherIsSafe(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	defer c.Close()

	assert.NotPanics(t, c.InvalidateSettings)
}

// A settings message arriving from another instance must reach the refresher,
// since that is the only thing that reloads the in-memory map there.
func TestApplyInvalidation_SettingsReachesRefresher(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	defer c.Close()

	calls := 0
	c.SetSettingsRefresher(func() { calls++ })

	sync := NewCacheSync(c, "")
	sync.applyInvalidation(InvalidationMessage{Type: InvalidationTypeSettings})

	assert.Equal(t, 1, calls)
}
