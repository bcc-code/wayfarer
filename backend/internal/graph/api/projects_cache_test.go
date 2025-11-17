package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// TestCacheInvalidation tests that cache is properly invalidated for project operations
func TestCacheInvalidationOnUpdate(t *testing.T) {
	projectID := "PR01K9Q3TQYGR8W5JHW4GMVPWS44"

	// Create cache
	testCache, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	// Add project to cache
	testCache.Set(cache.ProjectKey(projectID), &model.Project{
		ID:   projectID,
		Name: "Test Project",
	})
	testCache.Wait()

	// Verify it's in cache
	cached, ok := testCache.Get(cache.ProjectKey(projectID))
	require.True(t, ok, "project should be in cache")
	require.NotNil(t, cached)

	// Invalidate cache (simulating what UpdateProject does)
	testCache.InvalidateProject(projectID)

	// Verify it's no longer in cache
	_, ok = testCache.Get(cache.ProjectKey(projectID))
	assert.False(t, ok, "project should be removed from cache after invalidation")
}

func TestCacheInvalidationOnDelete(t *testing.T) {
	projectID := "PR01K9Q3TQYGR8W5JHW4GMVPWS44"

	testCache, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	// Add project to cache
	testCache.Set(cache.ProjectKey(projectID), &model.Project{
		ID:   projectID,
		Name: "Test Project",
	})

	// Add related data to cache (following the codebase's key pattern: <entity>:project:<id>)
	relatedKey := "challenges:project:" + projectID
	testCache.Set(relatedKey, "some data")
	testCache.Wait()

	// Verify both are in cache
	_, ok := testCache.Get(cache.ProjectKey(projectID))
	require.True(t, ok)
	_, ok = testCache.Get(relatedKey)
	require.True(t, ok)

	// Invalidate cache (simulating what DeleteProject does)
	testCache.InvalidateProject(projectID)

	// Verify main project is removed
	_, ok = testCache.Get(cache.ProjectKey(projectID))
	assert.False(t, ok, "project should be removed from cache")

	// Verify related data is also removed (prefix invalidation)
	_, ok = testCache.Get(relatedKey)
	assert.False(t, ok, "related project data should be removed from cache")
}

func TestCacheInvalidationOnArchive(t *testing.T) {
	projectID := "PR01K9Q3TQYGR8W5JHW4GMVPWS44"

	testCache, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	// Add project to cache
	testCache.Set(cache.ProjectKey(projectID), &model.Project{
		ID:         projectID,
		Name:       "Test Project",
		ArchivedAt: boolPtr(false),
	})
	testCache.Wait()

	// Verify it's in cache
	_, ok := testCache.Get(cache.ProjectKey(projectID))
	require.True(t, ok)

	// Invalidate cache (simulating what ArchiveProject does)
	testCache.InvalidateProject(projectID)

	// Verify it's no longer in cache
	_, ok = testCache.Get(cache.ProjectKey(projectID))
	assert.False(t, ok, "project should be removed from cache after archiving")
}

func TestCacheInvalidationDoesNotAffectOtherProjects(t *testing.T) {
	project1ID := "PR01K9Q3TQYGR8W5JHW4GMVPWS44"
	project2ID := "PR01K9Q3TQYGR8W5JHW4GMVPWS45"

	testCache, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	// Add both projects to cache
	testCache.Set(cache.ProjectKey(project1ID), &model.Project{
		ID:   project1ID,
		Name: "Test Project 1",
	})
	testCache.Set(cache.ProjectKey(project2ID), &model.Project{
		ID:   project2ID,
		Name: "Test Project 2",
	})
	testCache.Wait()

	// Invalidate only project 1
	testCache.InvalidateProject(project1ID)

	// Verify project 1 is removed
	_, ok := testCache.Get(cache.ProjectKey(project1ID))
	assert.False(t, ok, "project 1 should be removed from cache")

	// Verify project 2 is still in cache
	cached, ok := testCache.Get(cache.ProjectKey(project2ID))
	assert.True(t, ok, "project 2 should still be in cache")
	assert.NotNil(t, cached)
	project2 := cached.(*model.Project)
	assert.Equal(t, "Test Project 2", project2.Name)
}
