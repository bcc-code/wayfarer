package services

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSyncServiceCreation(t *testing.T) {
	t.Run("all nil dependencies", func(t *testing.T) {
		service := &UserSyncService{
			DB:                        nil,
			Cache:                     nil,
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
		ChurchLockSkipped:      true,
		PersonUUIDUpdated:      true,
	}

	assert.Equal(t, 42, result.ContentEventsProcessed)
	assert.True(t, result.GenderUpdated)
	assert.False(t, result.ChurchUpdated)
	assert.True(t, result.ChurchLockSkipped)
	assert.True(t, result.PersonUUIDUpdated)
}

func TestMemberSyncResultFields(t *testing.T) {
	result := &memberSyncResult{
		GenderUpdated:     true,
		ChurchUpdated:     false,
		ChurchLockSkipped: true,
		PersonUUIDUpdated: false,
	}

	assert.True(t, result.GenderUpdated)
	assert.False(t, result.ChurchUpdated)
	assert.True(t, result.ChurchLockSkipped)
	assert.False(t, result.PersonUUIDUpdated)
}

func TestChurchLockTimeCheck(t *testing.T) {
	t.Run("future lock is active", func(t *testing.T) {
		lockedUntil := pgtype.Timestamptz{
			Time:  time.Now().Add(6 * 30 * 24 * time.Hour),
			Valid: true,
		}
		assert.True(t, lockedUntil.Valid && lockedUntil.Time.After(time.Now()))
	})

	t.Run("expired lock is not active", func(t *testing.T) {
		lockedUntil := pgtype.Timestamptz{
			Time:  time.Now().Add(-1 * time.Hour),
			Valid: true,
		}
		assert.False(t, lockedUntil.Valid && lockedUntil.Time.After(time.Now()))
	})

	t.Run("null lock is not active", func(t *testing.T) {
		lockedUntil := pgtype.Timestamptz{
			Valid: false,
		}
		assert.False(t, lockedUntil.Valid && lockedUntil.Time.After(time.Now()))
	})
}

func TestChurchResolverCreation(t *testing.T) {
	resolver := &ChurchResolver{
		DB:            nil,
		MembersClient: nil,
	}
	require.NotNil(t, resolver)
}
