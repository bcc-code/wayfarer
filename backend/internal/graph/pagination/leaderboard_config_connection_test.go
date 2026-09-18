package pagination

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
)

func leaderboardConfigWithCreatedAt(id string, createdAt time.Time) *model.LeaderboardConfig {
	return &model.LeaderboardConfig{
		ID:        id,
		Name:      "Config " + id,
		CreatedAt: scalars.DateTime{Time: createdAt},
	}
}

func TestBuildLeaderboardConfigConnection_EmptyResults(t *testing.T) {
	params := BuildLeaderboardConfigConnectionParams{
		Configs:         []*model.LeaderboardConfig{},
		RequestedFirst:  intPtr(10),
		RequestedLast:   nil,
		RequestedAfter:  nil,
		RequestedBefore: nil,
		TotalCount:      0,
		HasMore:         false,
	}

	conn := BuildLeaderboardConfigConnection(params)

	require.NotNil(t, conn)
	assert.Empty(t, conn.Edges)
	assert.Equal(t, 0, conn.TotalCount)
	require.NotNil(t, conn.PageInfo)
	assert.False(t, conn.PageInfo.HasNextPage)
	assert.False(t, conn.PageInfo.HasPreviousPage)
	assert.Nil(t, conn.PageInfo.StartCursor)
	assert.Nil(t, conn.PageInfo.EndCursor)
}

func TestBuildLeaderboardConfigConnection_ForwardPagination(t *testing.T) {
	base := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	configs := []*model.LeaderboardConfig{
		leaderboardConfigWithCreatedAt("LC001", base),
		leaderboardConfigWithCreatedAt("LC002", base.Add(time.Minute)),
		leaderboardConfigWithCreatedAt("LC003", base.Add(2*time.Minute)),
	}

	tests := []struct {
		name                string
		params              BuildLeaderboardConfigConnectionParams
		expectedEdgeCount   int
		expectedHasNextPage bool
		expectedHasPrevPage bool
	}{
		{
			name: "first page with more results",
			params: BuildLeaderboardConfigConnectionParams{
				Configs:         configs,
				RequestedFirst:  intPtr(3),
				RequestedAfter:  nil,
				RequestedBefore: nil,
				TotalCount:      10,
				HasMore:         true,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: true,
			expectedHasPrevPage: false,
		},
		{
			name: "first page with no more results",
			params: BuildLeaderboardConfigConnectionParams{
				Configs:         configs,
				RequestedFirst:  intPtr(3),
				RequestedAfter:  nil,
				RequestedBefore: nil,
				TotalCount:      3,
				HasMore:         false,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: false,
			expectedHasPrevPage: false,
		},
		{
			name: "subsequent page with after cursor and more results",
			params: BuildLeaderboardConfigConnectionParams{
				Configs:        configs,
				RequestedFirst: intPtr(3),
				RequestedAfter: stringPtr(EncodeLeaderboardConfigCursor(base, "LC000")),
				TotalCount:     10,
				HasMore:        true,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: true,
			expectedHasPrevPage: true,
		},
		{
			name: "last page with after cursor and no more results",
			params: BuildLeaderboardConfigConnectionParams{
				Configs:        configs,
				RequestedFirst: intPtr(3),
				RequestedAfter: stringPtr(EncodeLeaderboardConfigCursor(base, "LC000")),
				TotalCount:     6,
				HasMore:        false,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: false,
			expectedHasPrevPage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := BuildLeaderboardConfigConnection(tt.params)

			require.NotNil(t, conn)
			assert.Equal(t, tt.expectedEdgeCount, len(conn.Edges))
			assert.Equal(t, tt.params.TotalCount, conn.TotalCount)

			require.NotNil(t, conn.PageInfo)
			assert.Equal(t, tt.expectedHasNextPage, conn.PageInfo.HasNextPage)
			assert.Equal(t, tt.expectedHasPrevPage, conn.PageInfo.HasPreviousPage)

			if len(conn.Edges) > 0 {
				require.NotNil(t, conn.PageInfo.StartCursor)
				require.NotNil(t, conn.PageInfo.EndCursor)
				assert.Equal(t, conn.Edges[0].Cursor, *conn.PageInfo.StartCursor)
				assert.Equal(t, conn.Edges[len(conn.Edges)-1].Cursor, *conn.PageInfo.EndCursor)
			}
		})
	}
}

func TestBuildLeaderboardConfigConnection_BackwardPagination(t *testing.T) {
	base := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	configs := []*model.LeaderboardConfig{
		leaderboardConfigWithCreatedAt("LC001", base),
		leaderboardConfigWithCreatedAt("LC002", base.Add(time.Minute)),
		leaderboardConfigWithCreatedAt("LC003", base.Add(2*time.Minute)),
	}

	tests := []struct {
		name                string
		params              BuildLeaderboardConfigConnectionParams
		expectedEdgeCount   int
		expectedHasNextPage bool
		expectedHasPrevPage bool
	}{
		{
			name: "last page with more previous results",
			params: BuildLeaderboardConfigConnectionParams{
				Configs:       configs,
				RequestedLast: intPtr(3),
				TotalCount:    10,
				HasMore:       true,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: false,
			expectedHasPrevPage: true,
		},
		{
			name: "last page with no previous results",
			params: BuildLeaderboardConfigConnectionParams{
				Configs:       configs,
				RequestedLast: intPtr(3),
				TotalCount:    3,
				HasMore:       false,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: false,
			expectedHasPrevPage: false,
		},
		{
			name: "previous page with before cursor",
			params: BuildLeaderboardConfigConnectionParams{
				Configs:         configs,
				RequestedLast:   intPtr(3),
				RequestedBefore: stringPtr(EncodeLeaderboardConfigCursor(base.Add(3*time.Minute), "LC004")),
				TotalCount:      10,
				HasMore:         true,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: true,
			expectedHasPrevPage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := BuildLeaderboardConfigConnection(tt.params)

			require.NotNil(t, conn)
			assert.Equal(t, tt.expectedEdgeCount, len(conn.Edges))
			assert.Equal(t, tt.params.TotalCount, conn.TotalCount)

			require.NotNil(t, conn.PageInfo)
			assert.Equal(t, tt.expectedHasNextPage, conn.PageInfo.HasNextPage)
			assert.Equal(t, tt.expectedHasPrevPage, conn.PageInfo.HasPreviousPage)

			if len(conn.Edges) > 0 {
				require.NotNil(t, conn.PageInfo.StartCursor)
				require.NotNil(t, conn.PageInfo.EndCursor)
			}
		})
	}
}

func TestBuildLeaderboardConfigConnection_EdgeContent(t *testing.T) {
	base := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	configs := []*model.LeaderboardConfig{
		leaderboardConfigWithCreatedAt("LC001", base),
		leaderboardConfigWithCreatedAt("LC002", base.Add(time.Minute)),
		leaderboardConfigWithCreatedAt("LC003", base.Add(2*time.Minute)),
	}

	params := BuildLeaderboardConfigConnectionParams{
		Configs:        configs,
		RequestedFirst: intPtr(3),
		TotalCount:     3,
		HasMore:        false,
	}

	conn := BuildLeaderboardConfigConnection(params)

	require.NotNil(t, conn)
	require.Equal(t, 3, len(conn.Edges))

	for i, edge := range conn.Edges {
		expectedCursor := EncodeLeaderboardConfigCursor(configs[i].CreatedAt.Time, configs[i].ID)
		assert.Equal(t, expectedCursor, edge.Cursor)

		require.NotNil(t, edge.Node)
		assert.Equal(t, configs[i].ID, edge.Node.ID)
		assert.Equal(t, configs[i].Name, edge.Node.Name)
	}
}

func TestBuildLeaderboardConfigConnection_NoPaginationParams(t *testing.T) {
	base := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	configs := []*model.LeaderboardConfig{
		leaderboardConfigWithCreatedAt("LC001", base),
		leaderboardConfigWithCreatedAt("LC002", base.Add(time.Minute)),
	}

	params := BuildLeaderboardConfigConnectionParams{
		Configs:         configs,
		RequestedFirst:  nil,
		RequestedLast:   nil,
		RequestedAfter:  nil,
		RequestedBefore: nil,
		TotalCount:      2,
		HasMore:         false,
	}

	conn := BuildLeaderboardConfigConnection(params)

	require.NotNil(t, conn)
	assert.Equal(t, 2, len(conn.Edges))
	assert.Equal(t, 2, conn.TotalCount)

	require.NotNil(t, conn.PageInfo)
	assert.False(t, conn.PageInfo.HasNextPage)
	assert.False(t, conn.PageInfo.HasPreviousPage)
}
