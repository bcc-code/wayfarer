package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

func TestConsentProjectForNewVersion(t *testing.T) {
	projectA := "PR01ARZ3NDEKTSV4RRFFQ69G5FA"
	projectB := "PR01ARZ3NDEKTSV4RRFFQ69G5FB"

	t.Run("inherits the project when none is requested", func(t *testing.T) {
		got, err := consentProjectForNewVersion("photo_consent", &projectA, nil)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, projectA, *got)
	})

	t.Run("stays global when the key is global and none is requested", func(t *testing.T) {
		got, err := consentProjectForNewVersion("privacy_policy", nil, nil)

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("accepts a request that matches the existing project", func(t *testing.T) {
		got, err := consentProjectForNewVersion("photo_consent", &projectA, &projectA)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, projectA, *got)
	})

	t.Run("rejects moving a consent to a different project", func(t *testing.T) {
		got, err := consentProjectForNewVersion("photo_consent", &projectA, &projectB)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), projectA)
	})

	t.Run("rejects narrowing a global consent to one project", func(t *testing.T) {
		// Narrowing would silently stop the consent applying to every other
		// project, so it has to be an explicit new key instead.
		got, err := consentProjectForNewVersion("privacy_policy", nil, &projectA)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "all projects")
	})
}

func TestLatestConsentVersionError(t *testing.T) {
	v1 := "CN01ARZ3NDEKTSV4RRFFQ69G5FV1"
	v2 := "CN01ARZ3NDEKTSV4RRFFQ69G5FV2"

	t.Run("allows the latest version", func(t *testing.T) {
		assert.NoError(t, latestConsentVersionError(v2, v2))
	})

	t.Run("refuses an older version", func(t *testing.T) {
		assert.Error(t, latestConsentVersionError(v1, v2))
	})

	t.Run("refuses when no version is published", func(t *testing.T) {
		assert.Error(t, latestConsentVersionError(v1, ""))
	})
}


func TestConsentBlockChanges(t *testing.T) {
	stays := projectUser{ProjectID: "PR1", UserID: "US1"}
	unblocked := projectUser{ProjectID: "PR1", UserID: "US2"}
	blocked := projectUser{ProjectID: "PR1", UserID: "US3"}
	otherProject := projectUser{ProjectID: "PR2", UserID: "US1"}

	t.Run("returns users blocked or unblocked by the change", func(t *testing.T) {
		got := consentBlockChanges(
			[]projectUser{stays, unblocked},
			[]projectUser{stays, blocked},
		)

		assert.Equal(t, map[projectUser]bool{unblocked: true, blocked: true}, got)
	})

	t.Run("tells projects of the same user apart", func(t *testing.T) {
		got := consentBlockChanges([]projectUser{stays}, []projectUser{stays, otherProject})

		assert.Equal(t, map[projectUser]bool{otherProject: true}, got)
	})

	t.Run("returns nothing when nobody changed", func(t *testing.T) {
		assert.Empty(t, consentBlockChanges([]projectUser{stays}, []projectUser{stays}))
		assert.Empty(t, consentBlockChanges(nil, nil))
	})
}

func TestTeamsOfChangedUsers(t *testing.T) {
	changed := map[projectUser]bool{{ProjectID: "PR1", UserID: "US1"}: true}
	memberships := []*sqlc.GetTeamsByUserIDsRow{
		{ID: "TM1", ProjectID: "PR1", UserID: "US1"},
		{ID: "TM2", ProjectID: "PR1", UserID: "US1"}, // a second team in the project
		{ID: "TM3", ProjectID: "PR2", UserID: "US1"}, // unchanged in this project
		{ID: "TM4", ProjectID: "PR1", UserID: "US2"}, // another user
	}

	got := teamsOfChangedUsers(memberships, changed)

	assert.Equal(t, map[string]string{"TM1": "PR1", "TM2": "PR1"}, got)
}
