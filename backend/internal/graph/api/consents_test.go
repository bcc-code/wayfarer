package api

import (
	"testing"

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

func TestValidateConsentScope(t *testing.T) {
	projectA := "PR01ARZ3NDEKTSV4RRFFQ69G5FA"

	t.Run("rejects a remote consent scoped to a project", func(t *testing.T) {
		err := validateConsentScope(true, &projectA)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "external system")
	})

	t.Run("allows a remote consent with no project", func(t *testing.T) {
		assert.NoError(t, validateConsentScope(true, nil))
	})

	t.Run("allows a local consent scoped to a project", func(t *testing.T) {
		assert.NoError(t, validateConsentScope(false, &projectA))
	})

	t.Run("allows a local consent with no project", func(t *testing.T) {
		assert.NoError(t, validateConsentScope(false, nil))
	})
}
