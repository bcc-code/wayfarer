package pagination

import "github.com/bcc-media/wayfarer/internal/graph/api/model"

// BuildUserConnectionParams holds parameters for building a UserConnection
type BuildUserConnectionParams struct {
	Users           []model.User
	RequestedFirst  *int
	RequestedLast   *int
	RequestedAfter  *string
	RequestedBefore *string
	TotalCount      int
	HasMore         bool // Indicates if there are more results beyond what was fetched
}

// BuildUserConnection constructs a Relay-style connection from query results
func BuildUserConnection(params BuildUserConnectionParams) *model.UserConnection {
	edges := make([]model.UserEdge, len(params.Users))
	for i, user := range params.Users {
		edges[i] = model.UserEdge{
			Cursor: EncodeCursor(user.ID),
			Node:   &user,
		}
	}

	pageInfo := buildPageInfo(params, edges)

	return &model.UserConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: params.TotalCount,
	}
}

// buildPageInfo constructs the PageInfo based on the request and results
func buildPageInfo(params BuildUserConnectionParams, edges []model.UserEdge) *model.PageInfo {
	pageInfo := &model.PageInfo{
		HasNextPage:     false,
		HasPreviousPage: false,
		StartCursor:     nil,
		EndCursor:       nil,
	}

	if len(edges) == 0 {
		return pageInfo
	}

	// Set start and end cursors
	startCursor := edges[0].Cursor
	endCursor := edges[len(edges)-1].Cursor
	pageInfo.StartCursor = &startCursor
	pageInfo.EndCursor = &endCursor

	// Determine hasNextPage and hasPreviousPage based on pagination direction
	if params.RequestedFirst != nil {
		// Forward pagination
		pageInfo.HasNextPage = params.HasMore
		// If we're paginating forward with an 'after' cursor, there must be a previous page
		if params.RequestedAfter != nil && *params.RequestedAfter != "" {
			pageInfo.HasPreviousPage = true
		}
	} else if params.RequestedLast != nil {
		// Backward pagination
		pageInfo.HasPreviousPage = params.HasMore
		// If we're paginating backward with a 'before' cursor, there must be a next page
		if params.RequestedBefore != nil && *params.RequestedBefore != "" {
			pageInfo.HasNextPage = true
		}
	}

	return pageInfo
}
