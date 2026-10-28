package api

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestConsentBlockChangesBetween(t *testing.T) {
	stays := projectUser{ProjectID: "PR1", UserID: "US1"}
	unblocked := projectUser{ProjectID: "PR1", UserID: "US2"}

	t.Run("returns the change when both listings are known", func(t *testing.T) {
		got, known := consentBlockChangesBetween(
			blockedSet{Users: []projectUser{stays, unblocked}, Known: true},
			blockedSet{Users: []projectUser{stays}, Known: true},
		)

		assert.True(t, known)
		assert.Equal(t, map[projectUser]bool{unblocked: true}, got)
	})

	t.Run("an empty listing is known, not failed", func(t *testing.T) {
		got, known := consentBlockChangesBetween(
			blockedSet{Known: true},
			blockedSet{Users: []projectUser{stays}, Known: true},
		)

		assert.True(t, known)
		assert.Equal(t, map[projectUser]bool{stays: true}, got)
	})

	t.Run("is unknown when either listing failed", func(t *testing.T) {
		known := blockedSet{Users: []projectUser{stays}, Known: true}

		_, beforeKnown := consentBlockChangesBetween(blockedSet{}, known)
		_, afterKnown := consentBlockChangesBetween(known, blockedSet{})

		assert.False(t, beforeKnown)
		assert.False(t, afterKnown)
	})
}

func TestConsentFromAdminRow(t *testing.T) {
	published := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FA"
	url := "https://example.com/personvern"
	managedBy := "members"

	t.Run("maps a published local consent", func(t *testing.T) {
		got := consentFromAdminRow(&sqlc.GetAdminConsentsRow{
			ID:          "CN01ARZ3NDEKTSV4RRFFQ69G5FA",
			Key:         "privacy_policy",
			Version:     2,
			Title:       "Personvern",
			ShortText:   "Kort",
			Body:        "# Personvern",
			Url:         &url,
			PublishedAt: pgtype.Timestamptz{Time: published, Valid: true},
			ProjectID:   &projectID,
			IsRequired:  true,
		})

		assert.Equal(t, "CN01ARZ3NDEKTSV4RRFFQ69G5FA", got.ID)
		assert.Equal(t, "privacy_policy", got.Key)
		assert.Equal(t, 2, got.Version)
		assert.Equal(t, "# Personvern", got.BodyMarkdown)
		assert.Equal(t, model.ConsentManagementTypeLocal, got.ManagementType)
		require.NotNil(t, got.PublishedAt)
		assert.Equal(t, published, got.PublishedAt.Time)
		require.NotNil(t, got.ProjectID)
		assert.Equal(t, projectID, *got.ProjectID)
		assert.True(t, got.IsRequired)
	})

	t.Run("leaves publishedAt nil for a draft", func(t *testing.T) {
		got := consentFromAdminRow(&sqlc.GetAdminConsentsRow{
			ID:          "CN01ARZ3NDEKTSV4RRFFQ69G5FB",
			Key:         "photo_consent",
			Version:     1,
			PublishedAt: pgtype.Timestamptz{},
		})

		assert.Nil(t, got.PublishedAt, "a draft has no publication date to show")
		assert.Nil(t, got.ArchivedAt, "nor an archive date")
		assert.Nil(t, got.ProjectID)
	})

	t.Run("carries the archive date of a retired consent", func(t *testing.T) {
		archived := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
		got := consentFromAdminRow(&sqlc.GetAdminConsentsRow{
			ID:          "CN01ARZ3NDEKTSV4RRFFQ69G5FD",
			Key:         "old_terms",
			Version:     1,
			PublishedAt: pgtype.Timestamptz{Time: published, Valid: true},
			ArchivedAt:  pgtype.Timestamptz{Time: archived, Valid: true},
		})

		require.NotNil(t, got.ArchivedAt)
		assert.Equal(t, archived, got.ArchivedAt.Time)
	})

	t.Run("marks a remotely managed consent", func(t *testing.T) {
		got := consentFromAdminRow(&sqlc.GetAdminConsentsRow{
			ID:        "CN01ARZ3NDEKTSV4RRFFQ69G5FC",
			Key:       "members_terms",
			IsRemote:  true,
			ManagedBy: &managedBy,
		})

		assert.Equal(t, model.ConsentManagementTypeRemote, got.ManagementType)
		require.NotNil(t, got.ManagedBy)
		assert.Equal(t, managedBy, *got.ManagedBy)
	})
}
