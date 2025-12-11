package pagination

import "github.com/bcc-media/wayfarer/internal/graph/api/model"

// getChallengeID extracts the ID from any Challenge implementation
func getChallengeID(c model.Challenge) string {
	switch v := c.(type) {
	case *model.SimpleChallenge:
		return v.ID
	case *model.QuizChallenge:
		return v.ID
	case *model.ExternalChallenge:
		return v.ID
	default:
		return ""
	}
}

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

// BuildProjectConnectionParams holds parameters for building a ProjectConnection
type BuildProjectConnectionParams struct {
	Projects        []model.Project
	RequestedFirst  *int
	RequestedLast   *int
	RequestedAfter  *string
	RequestedBefore *string
	TotalCount      int
	HasMore         bool
}

// BuildProjectConnection constructs a Relay-style connection from query results
func BuildProjectConnection(params BuildProjectConnectionParams) *model.ProjectConnection {
	edges := make([]model.ProjectEdge, len(params.Projects))
	for i, project := range params.Projects {
		edges[i] = model.ProjectEdge{
			Cursor: EncodeCursor(project.ID),
			Node:   &project,
		}
	}

	pageInfo := buildProjectPageInfo(params, edges)

	return &model.ProjectConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: params.TotalCount,
	}
}

// buildProjectPageInfo constructs the PageInfo for projects
func buildProjectPageInfo(params BuildProjectConnectionParams, edges []model.ProjectEdge) *model.PageInfo {
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

	// Determine hasNextPage
	if params.RequestedFirst != nil {
		pageInfo.HasNextPage = params.HasMore
	}

	// Determine hasPreviousPage
	if params.RequestedLast != nil {
		pageInfo.HasPreviousPage = params.HasMore
	} else if params.RequestedAfter != nil && *params.RequestedAfter != "" {
		pageInfo.HasPreviousPage = true
	}

	return pageInfo
}

// BuildTeamConnectionParams holds parameters for building a TeamConnection
type BuildTeamConnectionParams struct {
	Teams           []model.Team
	RequestedFirst  *int
	RequestedLast   *int
	RequestedAfter  *string
	RequestedBefore *string
	TotalCount      int
	HasMore         bool
}

// BuildTeamConnection constructs a Relay-style connection from query results
func BuildTeamConnection(params BuildTeamConnectionParams) *model.TeamConnection {
	edges := make([]model.TeamEdge, len(params.Teams))
	for i, team := range params.Teams {
		edges[i] = model.TeamEdge{
			Cursor: EncodeCursor(team.ID),
			Node:   &team,
		}
	}

	pageInfo := buildTeamPageInfo(params, edges)

	return &model.TeamConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: params.TotalCount,
	}
}

// buildTeamPageInfo constructs the PageInfo for teams
func buildTeamPageInfo(params BuildTeamConnectionParams, edges []model.TeamEdge) *model.PageInfo {
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

	// Determine hasNextPage
	if params.RequestedFirst != nil {
		pageInfo.HasNextPage = params.HasMore
	}

	// Determine hasPreviousPage
	if params.RequestedLast != nil {
		pageInfo.HasPreviousPage = params.HasMore
	} else if params.RequestedAfter != nil && *params.RequestedAfter != "" {
		pageInfo.HasPreviousPage = true
	}

	return pageInfo
}

// BuildSuperTeamConnectionParams holds parameters for building a SuperTeamConnection
type BuildSuperTeamConnectionParams struct {
	SuperTeams      []model.SuperTeam
	RequestedFirst  *int
	RequestedLast   *int
	RequestedAfter  *string
	RequestedBefore *string
	TotalCount      int
	HasMore         bool
}

// BuildSuperTeamConnection constructs a Relay-style connection from query results
func BuildSuperTeamConnection(params BuildSuperTeamConnectionParams) *model.SuperTeamConnection {
	edges := make([]model.SuperTeamEdge, len(params.SuperTeams))
	for i, superTeam := range params.SuperTeams {
		edges[i] = model.SuperTeamEdge{
			Cursor: EncodeCursor(superTeam.ID),
			Node:   &superTeam,
		}
	}

	pageInfo := buildSuperTeamPageInfo(params, edges)

	return &model.SuperTeamConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: params.TotalCount,
	}
}

// buildSuperTeamPageInfo constructs the PageInfo for super teams
func buildSuperTeamPageInfo(params BuildSuperTeamConnectionParams, edges []model.SuperTeamEdge) *model.PageInfo {
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

	// Determine hasNextPage
	if params.RequestedFirst != nil {
		pageInfo.HasNextPage = params.HasMore
	}

	// Determine hasPreviousPage
	if params.RequestedLast != nil {
		pageInfo.HasPreviousPage = params.HasMore
		// If we're paginating backward with a 'before' cursor, there must be a next page
		if params.RequestedBefore != nil && *params.RequestedBefore != "" {
			pageInfo.HasNextPage = true
		}
	} else if params.RequestedAfter != nil && *params.RequestedAfter != "" {
		pageInfo.HasPreviousPage = true
	}

	return pageInfo
}

// BuildAchievementConnectionParams holds parameters for building an AchievementConnection
type BuildAchievementConnectionParams struct {
	Achievements    []model.Achievement
	RequestedFirst  *int
	RequestedLast   *int
	RequestedAfter  *string
	RequestedBefore *string
	TotalCount      int
	HasMore         bool
}

// BuildAchievementConnection constructs a Relay-style connection from query results
func BuildAchievementConnection(params BuildAchievementConnectionParams) *model.AchievementConnection {
	edges := make([]model.AchievementEdge, len(params.Achievements))
	for i, achievement := range params.Achievements {
		edges[i] = model.AchievementEdge{
			Cursor: EncodeCursor(achievement.GetID()),
			Node:   achievement,
		}
	}

	pageInfo := buildAchievementPageInfo(params, edges)

	return &model.AchievementConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: params.TotalCount,
	}
}

// buildAchievementPageInfo constructs the PageInfo for achievements
func buildAchievementPageInfo(params BuildAchievementConnectionParams, edges []model.AchievementEdge) *model.PageInfo {
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

	// Determine hasNextPage
	if params.RequestedFirst != nil {
		pageInfo.HasNextPage = params.HasMore
	}

	// Determine hasPreviousPage
	if params.RequestedLast != nil {
		pageInfo.HasPreviousPage = params.HasMore
		// If we're paginating backward with a 'before' cursor, there must be a next page
		if params.RequestedBefore != nil && *params.RequestedBefore != "" {
			pageInfo.HasNextPage = true
		}
	} else if params.RequestedAfter != nil && *params.RequestedAfter != "" {
		pageInfo.HasPreviousPage = true
	}

	return pageInfo
}

// BuildChallengeConnectionParams holds parameters for building a ChallengeConnection
type BuildChallengeConnectionParams struct {
	Challenges      []model.Challenge
	RequestedFirst  *int
	RequestedLast   *int
	RequestedAfter  *string
	RequestedBefore *string
	TotalCount      int
	HasMore         bool
}

// BuildChallengeConnection constructs a Relay-style connection from query results
func BuildChallengeConnection(params BuildChallengeConnectionParams) *model.ChallengeConnection {
	edges := make([]model.ChallengeEdge, len(params.Challenges))
	for i, challenge := range params.Challenges {
		edges[i] = model.ChallengeEdge{
			Cursor: EncodeCursor(getChallengeID(challenge)),
			Node:   challenge,
		}
	}

	pageInfo := buildChallengePageInfo(params, edges)

	return &model.ChallengeConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: params.TotalCount,
	}
}

// buildChallengePageInfo constructs the PageInfo for challenges
func buildChallengePageInfo(params BuildChallengeConnectionParams, edges []model.ChallengeEdge) *model.PageInfo {
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

	// Determine hasNextPage
	if params.RequestedFirst != nil {
		pageInfo.HasNextPage = params.HasMore
	}

	// Determine hasPreviousPage
	if params.RequestedLast != nil {
		pageInfo.HasPreviousPage = params.HasMore
		// If we're paginating backward with a 'before' cursor, there must be a next page
		if params.RequestedBefore != nil && *params.RequestedBefore != "" {
			pageInfo.HasNextPage = true
		}
	} else if params.RequestedAfter != nil && *params.RequestedAfter != "" {
		pageInfo.HasPreviousPage = true
	}

	return pageInfo
}

// BuildChurchConnectionParams holds parameters for building a church connection
type BuildChurchConnectionParams struct {
	Churches        []model.Church
	RequestedFirst  *int
	RequestedLast   *int
	RequestedAfter  *string
	RequestedBefore *string
	TotalCount      int
	HasMore         bool
}

// BuildChurchConnection builds a GraphQL Church connection with edges and page info
func BuildChurchConnection(params BuildChurchConnectionParams) *model.ChurchConnection {
	edges := make([]model.ChurchEdge, len(params.Churches))
	for i, church := range params.Churches {
		edges[i] = model.ChurchEdge{
			Cursor: EncodeCursor(church.ID),
			Node:   &church,
		}
	}

	pageInfo := buildChurchPageInfo(params, edges)

	return &model.ChurchConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: params.TotalCount,
	}
}

// buildChurchPageInfo constructs the PageInfo for churches
func buildChurchPageInfo(params BuildChurchConnectionParams, edges []model.ChurchEdge) *model.PageInfo {
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

	// Determine hasNextPage
	if params.RequestedFirst != nil {
		pageInfo.HasNextPage = params.HasMore
	}

	// Determine hasPreviousPage
	if params.RequestedLast != nil {
		pageInfo.HasPreviousPage = params.HasMore
		// If we're paginating backward with a 'before' cursor, there must be a next page
		if params.RequestedBefore != nil && *params.RequestedBefore != "" {
			pageInfo.HasNextPage = true
		}
	} else if params.RequestedAfter != nil && *params.RequestedAfter != "" {
		pageInfo.HasPreviousPage = true
	}

	return pageInfo
}

// BuildStreakConnectionParams holds parameters for building a streak connection
type BuildStreakConnectionParams struct {
	Streaks         []model.Streak
	RequestedFirst  *int
	RequestedLast   *int
	RequestedAfter  *string
	RequestedBefore *string
	TotalCount      int
	HasMore         bool
}

// BuildStreakConnection builds a GraphQL Streak connection with edges and page info
func BuildStreakConnection(params BuildStreakConnectionParams) *model.StreakConnection {
	edges := make([]model.StreakEdge, len(params.Streaks))
	for i, streak := range params.Streaks {
		edges[i] = model.StreakEdge{
			Cursor: EncodeCursor(streak.ID),
			Node:   &streak,
		}
	}

	pageInfo := buildStreakPageInfo(params, edges)

	return &model.StreakConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: params.TotalCount,
	}
}

// buildStreakPageInfo constructs the PageInfo for streaks
func buildStreakPageInfo(params BuildStreakConnectionParams, edges []model.StreakEdge) *model.PageInfo {
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

	// Determine hasNextPage
	if params.RequestedFirst != nil {
		pageInfo.HasNextPage = params.HasMore
	}

	// Determine hasPreviousPage
	if params.RequestedLast != nil {
		pageInfo.HasPreviousPage = params.HasMore
		// If we're paginating backward with a 'before' cursor, there must be a next page
		if params.RequestedBefore != nil && *params.RequestedBefore != "" {
			pageInfo.HasNextPage = true
		}
	} else if params.RequestedAfter != nil && *params.RequestedAfter != "" {
		pageInfo.HasPreviousPage = true
	}

	return pageInfo
}

// BuildScoreJournalConnectionParams holds parameters for building a score journal connection
type BuildScoreJournalConnectionParams struct {
	ScoreJournals   []*model.ScoreJournal
	RequestedFirst  *int
	RequestedLast   *int
	RequestedAfter  *string
	RequestedBefore *string
	TotalCount      int
	HasMore         bool
}

// BuildScoreJournalConnection builds a GraphQL ScoreJournal connection with edges and page info
func BuildScoreJournalConnection(params BuildScoreJournalConnectionParams) *model.ScoreJournalConnection {
	edges := make([]model.ScoreJournalEdge, len(params.ScoreJournals))
	for i, entry := range params.ScoreJournals {
		edges[i] = model.ScoreJournalEdge{
			Cursor: EncodeCursor(entry.ID),
			Node:   entry,
		}
	}

	pageInfo := buildScoreJournalPageInfo(params, edges)

	return &model.ScoreJournalConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: params.TotalCount,
	}
}

// buildScoreJournalPageInfo constructs the PageInfo for score journal
func buildScoreJournalPageInfo(params BuildScoreJournalConnectionParams, edges []model.ScoreJournalEdge) *model.PageInfo {
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

	// Determine hasNextPage
	if params.RequestedFirst != nil {
		pageInfo.HasNextPage = params.HasMore
	}

	// Determine hasPreviousPage
	if params.RequestedLast != nil {
		pageInfo.HasPreviousPage = params.HasMore
		// If we're paginating backward with a 'before' cursor, there must be a next page
		if params.RequestedBefore != nil && *params.RequestedBefore != "" {
			pageInfo.HasNextPage = true
		}
	} else if params.RequestedAfter != nil && *params.RequestedAfter != "" {
		pageInfo.HasPreviousPage = true
	}

	return pageInfo
}

// BuildExternalContentConnectionParams holds parameters for building an ExternalContentConnection
type BuildExternalContentConnectionParams struct {
	ExternalContents []*model.ExternalContent
	RequestedFirst   *int
	RequestedLast    *int
	RequestedAfter   *string
	RequestedBefore  *string
	TotalCount       int
	HasMore          bool
}

// BuildExternalContentConnection constructs a Relay-style connection from query results
func BuildExternalContentConnection(params BuildExternalContentConnectionParams) *model.ExternalContentConnection {
	edges := make([]model.ExternalContentEdge, len(params.ExternalContents))
	for i, ec := range params.ExternalContents {
		edges[i] = model.ExternalContentEdge{
			Cursor: EncodeCursor(ec.ID),
			Node:   ec,
		}
	}

	pageInfo := buildExternalContentPageInfo(params, edges)

	return &model.ExternalContentConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: params.TotalCount,
	}
}

// buildExternalContentPageInfo constructs the PageInfo for external content
func buildExternalContentPageInfo(params BuildExternalContentConnectionParams, edges []model.ExternalContentEdge) *model.PageInfo {
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

	// Determine hasNextPage
	if params.RequestedFirst != nil {
		pageInfo.HasNextPage = params.HasMore
	}

	// Determine hasPreviousPage
	if params.RequestedLast != nil {
		pageInfo.HasPreviousPage = params.HasMore
		// If we're paginating backward with a 'before' cursor, there must be a next page
		if params.RequestedBefore != nil && *params.RequestedBefore != "" {
			pageInfo.HasNextPage = true
		}
	} else if params.RequestedAfter != nil && *params.RequestedAfter != "" {
		pageInfo.HasPreviousPage = true
	}

	return pageInfo
}
