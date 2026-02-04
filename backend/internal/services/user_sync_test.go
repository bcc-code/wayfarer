package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSyncServiceCreation(t *testing.T) {
	t.Run("all nil dependencies", func(t *testing.T) {
		service := &UserSyncService{
			DB:                        nil,
			SSFClient:                 nil,
			MembersClient:             nil,
			ChurchResolver:            nil,
			ContentAchievementService: nil,
		}
		require.NotNil(t, service)
	})

	t.Run("nil SSF client skips content sync", func(t *testing.T) {
		service := &UserSyncService{
			SSFClient: nil,
		}
		assert.Nil(t, service.SSFClient)
	})

	t.Run("nil Members client skips member sync", func(t *testing.T) {
		service := &UserSyncService{
			MembersClient: nil,
		}
		assert.Nil(t, service.MembersClient)
	})
}

func TestSyncUserResultFields(t *testing.T) {
	result := &SyncUserResult{
		ContentEventsProcessed: 42,
		GenderUpdated:          true,
		ChurchUpdated:          false,
		PersonUUIDUpdated:      true,
	}

	assert.Equal(t, 42, result.ContentEventsProcessed)
	assert.True(t, result.GenderUpdated)
	assert.False(t, result.ChurchUpdated)
	assert.True(t, result.PersonUUIDUpdated)
}

func TestChurchResolverCreation(t *testing.T) {
	resolver := &ChurchResolver{
		DB:            nil,
		MembersClient: nil,
	}
	require.NotNil(t, resolver)
}
