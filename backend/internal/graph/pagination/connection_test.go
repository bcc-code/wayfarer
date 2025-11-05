package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

func TestBuildUserConnection_EmptyResults(t *testing.T) {
	params := BuildUserConnectionParams{
		Users:           []model.User{},
		RequestedFirst:  intPtr(10),
		RequestedLast:   nil,
		RequestedAfter:  nil,
		RequestedBefore: nil,
		TotalCount:      0,
		HasMore:         false,
	}

	conn := BuildUserConnection(params)

	require.NotNil(t, conn)
	assert.Empty(t, conn.Edges)
	assert.Equal(t, 0, conn.TotalCount)
	require.NotNil(t, conn.PageInfo)
	assert.False(t, conn.PageInfo.HasNextPage)
	assert.False(t, conn.PageInfo.HasPreviousPage)
	assert.Nil(t, conn.PageInfo.StartCursor)
	assert.Nil(t, conn.PageInfo.EndCursor)
}

func TestBuildUserConnection_ForwardPagination(t *testing.T) {
	users := []model.User{
		{ID: "US001", Name: "User 1"},
		{ID: "US002", Name: "User 2"},
		{ID: "US003", Name: "User 3"},
	}

	tests := []struct {
		name                string
		params              BuildUserConnectionParams
		expectedEdgeCount   int
		expectedHasNextPage bool
		expectedHasPrevPage bool
	}{
		{
			name: "first page with more results",
			params: BuildUserConnectionParams{
				Users:           users,
				RequestedFirst:  intPtr(3),
				RequestedLast:   nil,
				RequestedAfter:  nil,
				RequestedBefore: nil,
				TotalCount:      100,
				HasMore:         true,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: true,
			expectedHasPrevPage: false,
		},
		{
			name: "first page with no more results",
			params: BuildUserConnectionParams{
				Users:           users,
				RequestedFirst:  intPtr(3),
				RequestedLast:   nil,
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
			name: "subsequent page with after cursor",
			params: BuildUserConnectionParams{
				Users:           users,
				RequestedFirst:  intPtr(3),
				RequestedLast:   nil,
				RequestedAfter:  stringPtr("VVMwMDE="),
				RequestedBefore: nil,
				TotalCount:      100,
				HasMore:         true,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: true,
			expectedHasPrevPage: true,
		},
		{
			name: "last page with after cursor",
			params: BuildUserConnectionParams{
				Users:           users,
				RequestedFirst:  intPtr(3),
				RequestedLast:   nil,
				RequestedAfter:  stringPtr("VVMwOTc="),
				RequestedBefore: nil,
				TotalCount:      100,
				HasMore:         false,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: false,
			expectedHasPrevPage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := BuildUserConnection(tt.params)

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

func TestBuildUserConnection_BackwardPagination(t *testing.T) {
	users := []model.User{
		{ID: "US001", Name: "User 1"},
		{ID: "US002", Name: "User 2"},
		{ID: "US003", Name: "User 3"},
	}

	tests := []struct {
		name                string
		params              BuildUserConnectionParams
		expectedEdgeCount   int
		expectedHasNextPage bool
		expectedHasPrevPage bool
	}{
		{
			name: "last page with more previous results",
			params: BuildUserConnectionParams{
				Users:           users,
				RequestedFirst:  nil,
				RequestedLast:   intPtr(3),
				RequestedAfter:  nil,
				RequestedBefore: nil,
				TotalCount:      100,
				HasMore:         true,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: false,
			expectedHasPrevPage: true,
		},
		{
			name: "last page with no previous results",
			params: BuildUserConnectionParams{
				Users:           users,
				RequestedFirst:  nil,
				RequestedLast:   intPtr(3),
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
			name: "previous page with before cursor",
			params: BuildUserConnectionParams{
				Users:           users,
				RequestedFirst:  nil,
				RequestedLast:   intPtr(3),
				RequestedAfter:  nil,
				RequestedBefore: stringPtr("VVMxMDA="),
				TotalCount:      100,
				HasMore:         true,
			},
			expectedEdgeCount:   3,
			expectedHasNextPage: false,
			expectedHasPrevPage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := BuildUserConnection(tt.params)

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

func TestBuildUserConnection_EdgeContent(t *testing.T) {
	users := []model.User{
		{ID: "US001", Name: "Alice", Email: "alice@example.com"},
		{ID: "US002", Name: "Bob", Email: "bob@example.com"},
		{ID: "US003", Name: "Charlie", Email: "charlie@example.com"},
	}

	params := BuildUserConnectionParams{
		Users:           users,
		RequestedFirst:  intPtr(3),
		RequestedLast:   nil,
		RequestedAfter:  nil,
		RequestedBefore: nil,
		TotalCount:      3,
		HasMore:         false,
	}

	conn := BuildUserConnection(params)

	require.NotNil(t, conn)
	require.Equal(t, 3, len(conn.Edges))

	for i, edge := range conn.Edges {
		// Verify cursor is correctly encoded
		assert.Equal(t, EncodeCursor(users[i].ID), edge.Cursor)

		// Verify node points to the correct user
		require.NotNil(t, edge.Node)
		assert.Equal(t, users[i].ID, edge.Node.ID)
		assert.Equal(t, users[i].Name, edge.Node.Name)
		assert.Equal(t, users[i].Email, edge.Node.Email)
	}
}

func TestBuildUserConnection_SingleResult(t *testing.T) {
	users := []model.User{
		{ID: "US001", Name: "Single User"},
	}

	params := BuildUserConnectionParams{
		Users:           users,
		RequestedFirst:  intPtr(1),
		RequestedLast:   nil,
		RequestedAfter:  nil,
		RequestedBefore: nil,
		TotalCount:      1,
		HasMore:         false,
	}

	conn := BuildUserConnection(params)

	require.NotNil(t, conn)
	assert.Equal(t, 1, len(conn.Edges))
	assert.Equal(t, 1, conn.TotalCount)

	require.NotNil(t, conn.PageInfo)
	assert.False(t, conn.PageInfo.HasNextPage)
	assert.False(t, conn.PageInfo.HasPreviousPage)
	require.NotNil(t, conn.PageInfo.StartCursor)
	require.NotNil(t, conn.PageInfo.EndCursor)
	assert.Equal(t, *conn.PageInfo.StartCursor, *conn.PageInfo.EndCursor)
}

func TestBuildUserConnection_NoPaginationParams(t *testing.T) {
	users := []model.User{
		{ID: "US001", Name: "User 1"},
		{ID: "US002", Name: "User 2"},
	}

	params := BuildUserConnectionParams{
		Users:           users,
		RequestedFirst:  nil,
		RequestedLast:   nil,
		RequestedAfter:  nil,
		RequestedBefore: nil,
		TotalCount:      2,
		HasMore:         false,
	}

	conn := BuildUserConnection(params)

	require.NotNil(t, conn)
	assert.Equal(t, 2, len(conn.Edges))
	assert.Equal(t, 2, conn.TotalCount)

	require.NotNil(t, conn.PageInfo)
	assert.False(t, conn.PageInfo.HasNextPage)
	assert.False(t, conn.PageInfo.HasPreviousPage)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
