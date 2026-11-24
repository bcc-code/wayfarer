package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const adminConsentsQuery = `
	query AdminConsents($projectId: ID) {
		adminConsents(projectId: $projectId) {
			id
			key
			version
			publishedAt
			project { id }
		}
	}
`

type adminConsentRow struct {
	ID          string  `json:"id"`
	Key         string  `json:"key"`
	Version     int     `json:"version"`
	PublishedAt *string `json:"publishedAt"`
	Project     *struct {
		ID string `json:"id"`
	} `json:"project"`
}

// TestAdminConsentsListing covers the admin listing, which — unlike the
// `consents` query — has to show drafts: they are published from that page.
func TestAdminConsentsListing(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(data.ProjectIDs), 2, "scoping needs two projects")

	projectA, projectB := data.ProjectIDs[0], data.ProjectIDs[1]
	adminUserID, plainUserID := data.UserIDs[0], data.UserIDs[1]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	// A live v1 with a draft v2 on top: the case the old listing got wrong.
	liveV1, err := dbMgr.CreateTestConsentVersion(ctx, "privacy_policy", 1, nil, &past)
	require.NoError(t, err)
	draftV2, err := dbMgr.CreateTestConsentVersion(ctx, "privacy_policy", 2, nil, nil)
	require.NoError(t, err)
	// A superseded version that should stay hidden.
	_, err = dbMgr.CreateTestConsentVersion(ctx, "terms", 1, nil, &past)
	require.NoError(t, err)
	liveTermsV2, err := dbMgr.CreateTestConsentVersion(ctx, "terms", 2, nil, &past)
	require.NoError(t, err)
	scheduled, err := dbMgr.CreateTestConsentVersion(ctx, "scheduled_consent", 1, nil, &future)
	require.NoError(t, err)
	projectAConsent, err := dbMgr.CreateTestConsentVersion(ctx, "photo_consent_a", 1, &projectA, &past)
	require.NoError(t, err)
	projectBConsent, err := dbMgr.CreateTestConsentVersion(ctx, "photo_consent_b", 1, &projectB, nil)
	require.NoError(t, err)

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)
	userToken, err := testutil.GenerateUserToken(plainUserID)
	require.NoError(t, err)

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	list := func(t *testing.T, projectID *string) []adminConsentRow {
		t.Helper()
		vars := map[string]any{}
		if projectID != nil {
			vars["projectId"] = *projectID
		}

		resp := client.WithAuth(adminToken).MustExecute(t, adminConsentsQuery, vars)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AdminConsents []adminConsentRow `json:"adminConsents"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.AdminConsents
	}

	idsOf := func(rows []adminConsentRow) []string {
		ids := make([]string, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
		return ids
	}

	t.Run("shows the live version and the draft stacked on it", func(t *testing.T) {
		ids := idsOf(list(t, nil))

		assert.Contains(t, ids, liveV1, "the live version must stay visible under a draft")
		assert.Contains(t, ids, draftV2, "the draft is the whole point of the admin listing")
	})

	t.Run("includes consents that are only scheduled", func(t *testing.T) {
		assert.Contains(t, idsOf(list(t, nil)), scheduled)
	})

	t.Run("hides superseded published versions", func(t *testing.T) {
		rows := list(t, nil)

		var termsVersions []int
		for _, row := range rows {
			if row.Key == "terms" {
				termsVersions = append(termsVersions, row.Version)
			}
		}

		assert.Equal(t, []int{2}, termsVersions, "only the live version of a key is listed")
		assert.Contains(t, idsOf(rows), liveTermsV2)
	})

	t.Run("lists every project's consents when unfiltered", func(t *testing.T) {
		ids := idsOf(list(t, nil))

		assert.Contains(t, ids, projectAConsent)
		assert.Contains(t, ids, projectBConsent)
	})

	t.Run("a project filter keeps its own consents and the global ones", func(t *testing.T) {
		rows := list(t, &projectA)
		ids := idsOf(rows)

		assert.Contains(t, ids, projectAConsent, "the project's own consent")
		assert.Contains(t, ids, liveV1, "a global consent applies to this project too")
		assert.Contains(t, ids, draftV2, "and so does a global draft")
		assert.NotContains(t, ids, projectBConsent, "another project's consent must not leak in")

		for _, row := range rows {
			if row.ID == projectAConsent {
				require.NotNil(t, row.Project)
				assert.Equal(t, projectA, row.Project.ID)
			}
		}
	})

	t.Run("is refused to non-admins", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, adminConsentsQuery, map[string]any{})

		assert.True(t, resp.HasErrors(), "drafts are unpublished text, not public")
	})
}

// TestArchiveConsent covers retiring a consent, the only way one leaves
// circulation: user_consent_history has to outlive it.
func TestArchiveConsent(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	projectID := data.ProjectIDs[0]
	require.NoError(t, dbMgr.SetCurrentProject(ctx, projectID))

	adminUserID, plainUserID := data.UserIDs[0], data.UserIDs[1]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	past := time.Now().Add(-24 * time.Hour)
	olderPast := time.Now().Add(-48 * time.Hour)

	draft, err := dbMgr.CreateTestConsentVersion(ctx, "draft_to_retire", 1, nil, nil)
	require.NoError(t, err)
	// Archiving the live v2 must retire the key, not promote v1.
	_, err = dbMgr.CreateTestConsentVersion(ctx, "retire_me", 1, nil, &olderPast)
	require.NoError(t, err)
	liveV2, err := dbMgr.CreateTestConsentVersion(ctx, "retire_me", 2, nil, &past)
	require.NoError(t, err)
	keepLive, err := dbMgr.CreateTestConsentVersion(ctx, "keep_me", 1, nil, &past)
	require.NoError(t, err)

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)
	userToken, err := testutil.GenerateUserToken(plainUserID)
	require.NoError(t, err)

	// With the cache: Ristretto's Set is async, and one that has not landed
	// would survive the invalidation under test. testCache.Wait() flushes it.
	router, testCache, cleanup, err := testutil.SetupTestServerWithCache(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	const archiveMutation = `
		mutation ArchiveConsent($id: ID!) {
			archiveConsent(id: $id) { id archivedAt }
		}
	`

	archive := func(t *testing.T, id string) *testutil.GraphQLResponse {
		t.Helper()
		return client.WithAuth(adminToken).MustExecute(t, archiveMutation, map[string]any{"id": id})
	}

	list := func(t *testing.T, includeArchived bool) []adminConsentRow {
		t.Helper()
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query AdminConsents($includeArchived: Boolean) {
				adminConsents(includeArchived: $includeArchived) {
					id
					key
					version
					archivedAt
				}
			}
		`, map[string]any{"includeArchived": includeArchived})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AdminConsents []adminConsentRow `json:"adminConsents"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.AdminConsents
	}

	listedIDs := func(rows []adminConsentRow) []string {
		ids := make([]string, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
		return ids
	}

	pendingKeys := func(t *testing.T) []string {
		t.Helper()
		resp := client.WithAuth(userToken).MustExecute(t, `
			query PendingConsents { pendingConsents { key version } }
		`, map[string]any{})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			PendingConsents []struct {
				Key     string `json:"key"`
				Version int    `json:"version"`
			} `json:"pendingConsents"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		keys := make([]string, 0, len(result.PendingConsents))
		for _, consent := range result.PendingConsents {
			keys = append(keys, consent.Key)
		}
		return keys
	}

	// The rendered gate reads ConsentStatus, a cached path of its own.
	gatedKeys := func(t *testing.T) []string {
		t.Helper()
		resp := client.WithAuth(userToken).MustExecute(t, `
			query ConsentGate { me { consentStatus { pendingConsents { key } } } }
		`, map[string]any{})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Me struct {
				ConsentStatus struct {
					PendingConsents []struct {
						Key string `json:"key"`
					} `json:"pendingConsents"`
				} `json:"consentStatus"`
			} `json:"me"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		keys := make([]string, 0, len(result.Me.ConsentStatus.PendingConsents))
		for _, consent := range result.Me.ConsentStatus.PendingConsents {
			keys = append(keys, consent.Key)
		}
		return keys
	}

	t.Run("a live consent is pending before it is archived", func(t *testing.T) {
		assert.Contains(t, pendingKeys(t), "retire_me")
		// Warms the cache, so the next subtest reads a warm one.
		assert.Contains(t, gatedKeys(t), "retire_me")
		testCache.Wait()
	})

	t.Run("archives a draft", func(t *testing.T) {
		resp := archive(t, draft)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			ArchiveConsent struct {
				ArchivedAt *string `json:"archivedAt"`
			} `json:"archiveConsent"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.NotNil(t, result.ArchiveConsent.ArchivedAt)

		assert.NotContains(t, listedIDs(list(t, false)), draft)
		assert.Contains(t, listedIDs(list(t, true)), draft, "includeArchived shows it again")
	})

	t.Run("archiving the live version retires the whole key", func(t *testing.T) {
		resp := archive(t, liveV2)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		keys := pendingKeys(t)
		assert.NotContains(t, keys, "retire_me", "an archived consent stops being asked for")
		assert.Contains(t, keys, "keep_me", "other consents are untouched")

		gated := gatedKeys(t)
		assert.NotContains(t, gated, "retire_me", "the cached gate must be invalidated too")
		assert.Contains(t, gated, "keep_me")
	})

	t.Run("does not resurrect the version underneath", func(t *testing.T) {
		for _, row := range list(t, false) {
			assert.NotEqual(t, "retire_me", row.Key,
				"archiving v2 must not promote v1 back into the listing")
		}
	})

	t.Run("refuses to accept an archived consent", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation AcceptConsent($consentId: ID!) {
				acceptConsent(consentId: $consentId) { id }
			}
		`, map[string]any{"consentId": liveV2})

		require.True(t, resp.HasErrors(), "a retired consent takes no new answers")
		assert.Contains(t, resp.ErrorMessage(), "archived")
	})

	t.Run("refuses to archive the same consent twice", func(t *testing.T) {
		resp := archive(t, liveV2)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "already archived")
	})

	t.Run("unarchiving puts a consent back into circulation", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation UnarchiveConsent($id: ID!) {
				unarchiveConsent(id: $id) { id archivedAt }
			}
		`, map[string]any{"id": liveV2})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			UnarchiveConsent struct {
				ArchivedAt *string `json:"archivedAt"`
			} `json:"unarchiveConsent"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Nil(t, result.UnarchiveConsent.ArchivedAt)

		assert.Contains(t, pendingKeys(t), "retire_me")
	})

	t.Run("keeps the consent history of an archived consent", func(t *testing.T) {
		accept := client.WithAuth(userToken).MustExecute(t, `
			mutation AcceptConsent($consentId: ID!) {
				acceptConsent(consentId: $consentId) { id }
			}
		`, map[string]any{"consentId": keepLive})
		require.False(t, accept.HasErrors(), "unexpected error: %s", accept.ErrorMessage())

		resp := archive(t, keepLive)
		require.False(t, resp.HasErrors(), "an answered consent is archivable, unlike deletable")

		count, err := dbMgr.GetUserConsentHistoryCount(ctx, plainUserID, "keep_me")
		require.NoError(t, err)
		assert.Equal(t, 1, count, "the audit trail outlives the consent")
	})

	t.Run("is refused to non-admins", func(t *testing.T) {
		other, err := dbMgr.CreateTestConsentVersion(ctx, "not_yours_to_retire", 1, nil, nil)
		require.NoError(t, err)

		resp := client.WithAuth(userToken).MustExecute(t, archiveMutation, map[string]any{"id": other})

		assert.True(t, resp.HasErrors())
	})
}
