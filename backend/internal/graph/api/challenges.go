package api

import (
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/jackc/pgx/v5/pgtype"
)

// ChallengeType constants
const (
	ChallengeTypeSimple   = "SIMPLE"
	ChallengeTypeQuiz     = "QUIZ"
	ChallengeTypeExternal = "EXTERNAL"
)

// buildChallengeFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildChallengeFilterParamsCursor(filter *model.ChallengeFilter, first *int, after *string, last *int, before *string) (sqlc.GetChallengesFilteredCursorParams, error) {
	params := sqlc.GetChallengesFilteredCursorParams{}

	// Apply filters if provided
	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.EventID != nil {
			params.Eventid = *filter.EventID
		}

		if filter.ChallengeType != nil {
			params.Challengetype = string(*filter.ChallengeType)
		}

		if filter.PublishedAfter != nil {
			params.Publishedafter = pgtype.Timestamptz{
				Time:  filter.PublishedAfter.Time,
				Valid: true,
			}
		}

		if filter.PublishedBefore != nil {
			params.Publishedbefore = pgtype.Timestamptz{
				Time:  filter.PublishedBefore.Time,
				Valid: true,
			}
		}
	}

	// Handle cursor pagination
	isBackward := false
	var limit int

	if first != nil && last != nil {
		return params, fmt.Errorf("cannot specify both first and last")
	}

	if first != nil {
		limit = *first + 1 // Fetch one extra to determine hasNextPage
		isBackward = false
	} else if last != nil {
		limit = *last + 1 // Fetch one extra to determine hasPreviousPage
		isBackward = true
	} else {
		// Default page size
		limit = 11 // 10 items + 1 to check for next page
		isBackward = false
	}

	params.Querylimit = int32(limit)
	params.Isbackward = isBackward

	// Set cursors
	if after != nil && *after != "" {
		params.Aftercursor = *after
	}

	if before != nil && *before != "" {
		params.Beforecursor = *before
	}

	return params, nil
}

// buildCountChallengesFilterParams converts GraphQL filter to database count query parameters
func buildCountChallengesFilterParams(filter *model.ChallengeFilter) sqlc.CountChallengesFilteredParams {
	params := sqlc.CountChallengesFilteredParams{}

	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.EventID != nil {
			params.Eventid = *filter.EventID
		}

		if filter.ChallengeType != nil {
			params.Challengetype = string(*filter.ChallengeType)
		}

		if filter.PublishedAfter != nil {
			params.Publishedafter = pgtype.Timestamptz{
				Time:  filter.PublishedAfter.Time,
				Valid: true,
			}
		}

		if filter.PublishedBefore != nil {
			params.Publishedbefore = pgtype.Timestamptz{
				Time:  filter.PublishedBefore.Time,
				Valid: true,
			}
		}
	}

	return params
}

// buildChallengeCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildChallengeCacheKeyParams(filter *model.ChallengeFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if len(filter.Ids) > 0 {
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
		}
		if filter.ProjectID != nil {
			params["projectid"] = *filter.ProjectID
		}
		if filter.EventID != nil {
			params["eventid"] = *filter.EventID
		}
		if filter.ChallengeType != nil {
			params["challengetype"] = string(*filter.ChallengeType)
		}
		if filter.PublishedAfter != nil {
			params["publishedafter"] = filter.PublishedAfter.Format(time.RFC3339)
		}
		if filter.PublishedBefore != nil {
			params["publishedbefore"] = filter.PublishedBefore.Format(time.RFC3339)
		}
	}

	// Add pagination parameters
	if first != nil {
		params["first"] = fmt.Sprintf("%d", *first)
	}
	if after != nil && *after != "" {
		params["after"] = *after
	}
	if last != nil {
		params["last"] = fmt.Sprintf("%d", *last)
	}
	if before != nil && *before != "" {
		params["before"] = *before
	}

	return params
}

// convertRowToChallenge converts a database row to the appropriate Challenge implementation
// based on the challenge_type column. Returns the model.Challenge interface.
func convertRowToChallenge(row *sqlc.Challenge) model.Challenge {
	// Set common timestamp fields
	var publishedAt, visibleAt, startedAt, endTime *scalars.DateTime
	if row.PublishedAt.Valid {
		dt := scalars.DateTime{Time: row.PublishedAt.Time}
		publishedAt = &dt
	}
	if row.VisibleAt.Valid {
		dt := scalars.DateTime{Time: row.VisibleAt.Time}
		visibleAt = &dt
	}
	if row.StartedAt.Valid {
		dt := scalars.DateTime{Time: row.StartedAt.Time}
		startedAt = &dt
	}
	if row.EndTime.Valid {
		dt := scalars.DateTime{Time: row.EndTime.Time}
		endTime = &dt
	}

	switch row.ChallengeType {
	case ChallengeTypeQuiz:
		return &model.QuizChallenge{
			ID:                          row.ID,
			Name:                        row.Name,
			Description:                 scalars.HTML(row.Description),
			Image:                       row.ImageUrl,
			ButtonText:                  row.ButtonText,
			ProjectID:                   row.ProjectID,
			EventID:                     row.EventID,
			PublishedAt:                 publishedAt,
			VisibleAt:                   visibleAt,
			StartedAt:                   startedAt,
			EndTime:                     endTime,
			RequiresTeamMembership:      row.RequiresTeamMembership,
			RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		}
	case ChallengeTypeExternal:
		url := ""
		if row.Url != nil {
			url = *row.Url
		}
		return &model.ExternalChallenge{
			ID:                          row.ID,
			Name:                        row.Name,
			Description:                 scalars.HTML(row.Description),
			Image:                       row.ImageUrl,
			URL:                         url,
			ButtonText:                  row.ButtonText,
			ProjectID:                   row.ProjectID,
			EventID:                     row.EventID,
			PublishedAt:                 publishedAt,
			VisibleAt:                   visibleAt,
			StartedAt:                   startedAt,
			EndTime:                     endTime,
			RequiresTeamMembership:      row.RequiresTeamMembership,
			RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		}
	default: // SIMPLE
		return &model.SimpleChallenge{
			ID:                          row.ID,
			Name:                        row.Name,
			Description:                 scalars.HTML(row.Description),
			Image:                       row.ImageUrl,
			ButtonText:                  row.ButtonText,
			ProjectID:                   row.ProjectID,
			EventID:                     row.EventID,
			PublishedAt:                 publishedAt,
			VisibleAt:                   visibleAt,
			StartedAt:                   startedAt,
			EndTime:                     endTime,
			AllowSelfCompletion:         row.AllowSelfCompletion,
			RequiresTeamMembership:      row.RequiresTeamMembership,
			RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		}
	}
}

// Type-specific conversion functions for returning concrete types

func convertRowToSimpleChallenge(row *sqlc.Challenge) *model.SimpleChallenge {
	var publishedAt, visibleAt, startedAt, endTime *scalars.DateTime
	if row.PublishedAt.Valid {
		dt := scalars.DateTime{Time: row.PublishedAt.Time}
		publishedAt = &dt
	}
	if row.VisibleAt.Valid {
		dt := scalars.DateTime{Time: row.VisibleAt.Time}
		visibleAt = &dt
	}
	if row.StartedAt.Valid {
		dt := scalars.DateTime{Time: row.StartedAt.Time}
		startedAt = &dt
	}
	if row.EndTime.Valid {
		dt := scalars.DateTime{Time: row.EndTime.Time}
		endTime = &dt
	}

	return &model.SimpleChallenge{
		ID:                          row.ID,
		Name:                        row.Name,
		Description:                 scalars.HTML(row.Description),
		Image:                       row.ImageUrl,
		ButtonText:                  row.ButtonText,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		PublishedAt:                 publishedAt,
		VisibleAt:                   visibleAt,
		StartedAt:                   startedAt,
		EndTime:                     endTime,
		AllowSelfCompletion:         row.AllowSelfCompletion,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
	}
}

func convertRowToQuizChallenge(row *sqlc.Challenge) *model.QuizChallenge {
	var publishedAt, visibleAt, startedAt, endTime *scalars.DateTime
	if row.PublishedAt.Valid {
		dt := scalars.DateTime{Time: row.PublishedAt.Time}
		publishedAt = &dt
	}
	if row.VisibleAt.Valid {
		dt := scalars.DateTime{Time: row.VisibleAt.Time}
		visibleAt = &dt
	}
	if row.StartedAt.Valid {
		dt := scalars.DateTime{Time: row.StartedAt.Time}
		startedAt = &dt
	}
	if row.EndTime.Valid {
		dt := scalars.DateTime{Time: row.EndTime.Time}
		endTime = &dt
	}

	return &model.QuizChallenge{
		ID:                          row.ID,
		Name:                        row.Name,
		Description:                 scalars.HTML(row.Description),
		Image:                       row.ImageUrl,
		ButtonText:                  row.ButtonText,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		PublishedAt:                 publishedAt,
		VisibleAt:                   visibleAt,
		StartedAt:                   startedAt,
		EndTime:                     endTime,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
	}
}

func convertRowToExternalChallenge(row *sqlc.Challenge) *model.ExternalChallenge {
	var publishedAt, visibleAt, startedAt, endTime *scalars.DateTime
	if row.PublishedAt.Valid {
		dt := scalars.DateTime{Time: row.PublishedAt.Time}
		publishedAt = &dt
	}
	if row.VisibleAt.Valid {
		dt := scalars.DateTime{Time: row.VisibleAt.Time}
		visibleAt = &dt
	}
	if row.StartedAt.Valid {
		dt := scalars.DateTime{Time: row.StartedAt.Time}
		startedAt = &dt
	}
	if row.EndTime.Valid {
		dt := scalars.DateTime{Time: row.EndTime.Time}
		endTime = &dt
	}

	url := ""
	if row.Url != nil {
		url = *row.Url
	}

	return &model.ExternalChallenge{
		ID:                          row.ID,
		Name:                        row.Name,
		Description:                 scalars.HTML(row.Description),
		Image:                       row.ImageUrl,
		URL:                         url,
		ButtonText:                  row.ButtonText,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		PublishedAt:                 publishedAt,
		VisibleAt:                   visibleAt,
		StartedAt:                   startedAt,
		EndTime:                     endTime,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
	}
}

// Helper conversion functions for different row types

func convertCreateChallengeRowToChallenge(row *sqlc.CreateChallengeRow) model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		ChallengeType:               row.ChallengeType,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		AllowSelfCompletion:         row.AllowSelfCompletion,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
}

func convertUpdateChallengeRowToChallenge(row *sqlc.UpdateChallengeRow) model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		ChallengeType:               row.ChallengeType,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		AllowSelfCompletion:         row.AllowSelfCompletion,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
}

func convertPublishChallengeRowToChallenge(row *sqlc.PublishChallengeRow) model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		ChallengeType:               row.ChallengeType,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		AllowSelfCompletion:         row.AllowSelfCompletion,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
}

func convertAssignChallengeToEventRowToChallenge(row *sqlc.AssignChallengeToEventRow) model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		ChallengeType:               row.ChallengeType,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		AllowSelfCompletion:         row.AllowSelfCompletion,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
}

func convertBulkPublishChallengesRowToChallenge(row *sqlc.BulkPublishChallengesRow) model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		ChallengeType:               row.ChallengeType,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		AllowSelfCompletion:         row.AllowSelfCompletion,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
}

func convertBulkCreateChallengesRowToChallenge(row *sqlc.BulkCreateChallengesRow) model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		ChallengeType:               row.ChallengeType,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		AllowSelfCompletion:         row.AllowSelfCompletion,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
}

func convertGetChallengesFilteredCursorRowToChallenge(row *sqlc.GetChallengesFilteredCursorRow) model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		ChallengeType:               row.ChallengeType,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		AllowSelfCompletion:         row.AllowSelfCompletion,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
}

func convertUpdateChallengeTimestampsRowToChallenge(row *sqlc.UpdateChallengeTimestampsRow) model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		ChallengeType:               row.ChallengeType,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		AllowSelfCompletion:         row.AllowSelfCompletion,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
}

func convertUpdateChallengeRequirementsRowToChallenge(row *sqlc.UpdateChallengeRequirementsRow) model.Challenge {
	return convertRowToChallenge(&sqlc.Challenge{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		EventID:                     row.EventID,
		ChallengeType:               row.ChallengeType,
		Name:                        row.Name,
		Description:                 row.Description,
		ImageUrl:                    row.ImageUrl,
		Url:                         row.Url,
		ButtonText:                  row.ButtonText,
		PublishedAt:                 row.PublishedAt,
		VisibleAt:                   row.VisibleAt,
		StartedAt:                   row.StartedAt,
		EndTime:                     row.EndTime,
		AllowSelfCompletion:         row.AllowSelfCompletion,
		RequiresTeamMembership:      row.RequiresTeamMembership,
		RequiresSuperTeamMembership: row.RequiresSuperTeamMembership,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
}

// Helper to get ProjectID from any Challenge implementation
func getChallengeProjectID(c model.Challenge) string {
	switch v := c.(type) {
	case *model.SimpleChallenge:
		return v.ProjectID
	case *model.QuizChallenge:
		return v.ProjectID
	case *model.ExternalChallenge:
		return v.ProjectID
	default:
		return ""
	}
}

// Helper to get EventID from any Challenge implementation
func getChallengeEventID(c model.Challenge) *string {
	switch v := c.(type) {
	case *model.SimpleChallenge:
		return v.EventID
	case *model.QuizChallenge:
		return v.EventID
	case *model.ExternalChallenge:
		return v.EventID
	default:
		return nil
	}
}

// Helper to get ID from any Challenge implementation
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

// Helper to get PublishedAt from any Challenge implementation
func getChallengePublishedAt(c model.Challenge) *scalars.DateTime {
	switch v := c.(type) {
	case *model.SimpleChallenge:
		return v.PublishedAt
	case *model.QuizChallenge:
		return v.PublishedAt
	case *model.ExternalChallenge:
		return v.PublishedAt
	default:
		return nil
	}
}

// Helper to get EndTime from any Challenge implementation
func getChallengeEndTime(c model.Challenge) *scalars.DateTime {
	switch v := c.(type) {
	case *model.SimpleChallenge:
		return v.EndTime
	case *model.QuizChallenge:
		return v.EndTime
	case *model.ExternalChallenge:
		return v.EndTime
	default:
		return nil
	}
}

// Helper to get RequiresTeamMembership from any Challenge implementation
func getChallengeRequiresTeamMembership(c model.Challenge) bool {
	switch v := c.(type) {
	case *model.SimpleChallenge:
		return v.RequiresTeamMembership
	case *model.QuizChallenge:
		return v.RequiresTeamMembership
	case *model.ExternalChallenge:
		return v.RequiresTeamMembership
	default:
		return false
	}
}

// Helper to get RequiresSuperTeamMembership from any Challenge implementation
func getChallengeRequiresSuperTeamMembership(c model.Challenge) bool {
	switch v := c.(type) {
	case *model.SimpleChallenge:
		return v.RequiresSuperTeamMembership
	case *model.QuizChallenge:
		return v.RequiresSuperTeamMembership
	case *model.ExternalChallenge:
		return v.RequiresSuperTeamMembership
	default:
		return false
	}
}

// Helper to get ChallengeType from any Challenge implementation
func getChallengeType(c model.Challenge) model.ChallengeType {
	switch c.(type) {
	case *model.SimpleChallenge:
		return model.ChallengeTypeSimple
	case *model.QuizChallenge:
		return model.ChallengeTypeQuiz
	case *model.ExternalChallenge:
		return model.ChallengeTypeExternal
	default:
		return model.ChallengeTypeSimple
	}
}

// validateCreateChallengeInput validates type-specific fields for CreateChallengeInput
func validateCreateChallengeInput(input model.CreateChallengeInput) error {
	switch input.Type {
	case model.ChallengeTypeSimple:
		// allowSelfCompletion is valid (optional, defaults true)
		if input.URL != nil {
			return fmt.Errorf("url is not allowed for SIMPLE challenges")
		}
	case model.ChallengeTypeQuiz:
		if input.AllowSelfCompletion != nil {
			return fmt.Errorf("allowSelfCompletion is not valid for QUIZ challenges")
		}
		if input.URL != nil {
			return fmt.Errorf("url is not valid for QUIZ challenges")
		}
	case model.ChallengeTypeExternal:
		if input.URL == nil || *input.URL == "" {
			return fmt.Errorf("url is required for EXTERNAL challenges")
		}
		if input.AllowSelfCompletion != nil {
			return fmt.Errorf("allowSelfCompletion is not valid for EXTERNAL challenges")
		}
	}
	return nil
}

// validateUpdateChallengeInput validates type-specific fields for UpdateChallengeInput
func validateUpdateChallengeInput(input model.UpdateChallengeInput, challengeType model.ChallengeType) error {
	switch challengeType {
	case model.ChallengeTypeSimple:
		// allowSelfCompletion is valid
		if input.URL != nil {
			return fmt.Errorf("url is not allowed for SIMPLE challenges")
		}
	case model.ChallengeTypeQuiz:
		if input.AllowSelfCompletion != nil {
			return fmt.Errorf("allowSelfCompletion is not valid for QUIZ challenges")
		}
		if input.URL != nil {
			return fmt.Errorf("url is not valid for QUIZ challenges")
		}
	case model.ChallengeTypeExternal:
		// url is valid (but must be non-empty if provided)
		if input.URL != nil && *input.URL == "" {
			return fmt.Errorf("url cannot be empty for EXTERNAL challenges")
		}
		if input.AllowSelfCompletion != nil {
			return fmt.Errorf("allowSelfCompletion is not valid for EXTERNAL challenges")
		}
	}
	return nil
}
