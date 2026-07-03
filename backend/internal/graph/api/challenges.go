package api

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/pagination"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services/push"
	"github.com/jackc/pgx/v5/pgtype"
)

// challengeFilterMode controls which challenges are returned by getFilteredChallenges.
type challengeFilterMode int

const (
	challengeFilterAll       challengeFilterMode = iota // All visible challenges (same as current Challenges resolver)
	challengeFilterActive                               // Not completed by user AND not past end time
	challengeFilterCompleted                            // Completed by user OR past end time
)

// getFilteredChallenges loads visible challenges for a project, optionally filtering by completion status.
// Uses batch queries for session access and enrollment to avoid N+1 sequential calls.
func (r *Resolver) getFilteredChallenges(ctx context.Context, projectID string, mode challengeFilterMode) ([]model.Challenge, error) {
	thunk := r.Loaders.ChallengesByProjectLoader.Load(ctx, projectID)
	challenges, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load challenges: %w", err)
	}

	userID, ok := middleware.GetUserID(ctx)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user not authenticated")
	}
	now := time.Now()

	// Step 1: Categorize challenges and batch-load quiz data
	quizChallengeIDs := make([]string, 0)
	for _, ch := range challenges {
		if _, ok := ch.(*model.QuizChallenge); ok {
			quizChallengeIDs = append(quizChallengeIDs, ch.GetID())
		}
	}

	// Batch load quizzes for all quiz challenges at once
	quizByChallenge := make(map[string]*model.Quiz)
	if len(quizChallengeIDs) > 0 {
		quizThunks := r.Loaders.QuizByChallengeIDLoader.LoadMany(ctx, quizChallengeIDs)
		quizResults, _ := quizThunks()
		for i, q := range quizResults {
			if q != nil {
				quizByChallenge[quizChallengeIDs[i]] = q
			}
		}
	}

	// Step 2: Batch check session access for all quiz challenges
	sessionAccessQuizIDs := make(map[string]bool)
	if len(quizByChallenge) > 0 {
		quizIDs := make([]string, 0, len(quizByChallenge))
		for _, q := range quizByChallenge {
			quizIDs = append(quizIDs, q.ID)
		}
		accessibleQuizIDs, err := r.DB.Queries.GetBulkUserSessionAccessQuizIDs(ctx, sqlc.GetBulkUserSessionAccessQuizIDsParams{
			Quizids: quizIDs,
			Userid:  userID,
		})
		if err == nil {
			for _, qid := range accessibleQuizIDs {
				sessionAccessQuizIDs[qid] = true
			}
		}
	}

	// Step 3: Batch load enrolled challenge IDs for this user+project
	enrolledChallengeIDs := make(map[string]bool)
	enrolledIDs, err := r.DB.Queries.GetUserEnrolledChallengeIDsInProject(ctx, sqlc.GetUserEnrolledChallengeIDsInProjectParams{
		Userid:    userID,
		Projectid: projectID,
	})
	if err == nil {
		for _, id := range enrolledIDs {
			enrolledChallengeIDs[id] = true
		}
	}

	// Step 4: Filter using pre-fetched batch data (no translations yet)
	visible := make([]model.Challenge, 0, len(challenges))
	for _, ch := range challenges {
		if _, ok := ch.(*model.QuizChallenge); ok {
			quiz := quizByChallenge[ch.GetID()]
			if quiz != nil && sessionAccessQuizIDs[quiz.ID] {
				visible = append(visible, ch)
			}
			continue
		}

		publishedAt := getChallengePublishedAt(ch)
		if publishedAt == nil || publishedAt.After(now) {
			continue
		}

		visibleAt := getChallengeVisibleAt(ch)
		isVisible := visibleAt != nil && !visibleAt.After(now)

		if !isVisible && !enrolledChallengeIDs[ch.GetID()] {
			continue
		}

		visible = append(visible, ch)
	}

	// Step 5: Filter by completion status if needed
	var finalChallenges []model.Challenge
	if mode == challengeFilterAll {
		finalChallenges = visible
	} else {
		// Load completion timestamps in batch for all visible challenges
		completionKeys := make([]loaders.UserChallengeKey, len(visible))
		for i, ch := range visible {
			completionKeys[i] = loaders.UserChallengeKey{UserID: userID, ChallengeID: ch.GetID()}
		}

		completionThunk := r.Loaders.UserChallengeCompletionTimestampLoader.LoadMany(ctx, completionKeys)
		completionTimestamps, _ := completionThunk()

		finalChallenges = make([]model.Challenge, 0, len(visible))
		for i, ch := range visible {
			endTime := ch.GetEndTime()
			pastEndTime := endTime != nil && endTime.Time.Before(now)

			var completed bool
			if i < len(completionTimestamps) && completionTimestamps[i] != nil {
				completed = true
			}

			isCompleted := completed || pastEndTime

			switch mode {
			case challengeFilterActive:
				if !isCompleted {
					finalChallenges = append(finalChallenges, ch)
				}
			case challengeFilterCompleted:
				if isCompleted {
					finalChallenges = append(finalChallenges, ch)
				}
			}
		}
	}

	// Step 6: Apply translations in batch after all filtering is done
	r.applyTranslationsToChallenges(ctx, finalChallenges)

	r.sortChallengesByEnrollment(ctx, userID, finalChallenges)
	return finalChallenges, nil
}

// getVisibleEventChallenges loads the visible challenges for an event using batch queries
// for quiz session access and enrollment, avoiding the per-challenge N+1 the naive resolver
// would issue. It mirrors the visibility rules of the original Event.challenges resolver,
// including support for unauthenticated viewers (userID == "").
func (r *Resolver) getVisibleEventChallenges(ctx context.Context, obj *model.Event) ([]model.Challenge, error) {
	thunk := r.Loaders.ChallengesByEventLoader.Load(ctx, obj.ID)
	challenges, err := thunk()
	if err != nil {
		return nil, err
	}

	userID, _ := middleware.GetUserID(ctx)
	now := time.Now()

	// Batch-load quiz session access for quiz challenges (authenticated viewers only).
	quizByChallenge := make(map[string]*model.Quiz)
	sessionAccessQuizIDs := make(map[string]bool)
	if userID != "" {
		quizChallengeIDs := make([]string, 0)
		for _, ch := range challenges {
			if _, ok := ch.(*model.QuizChallenge); ok {
				quizChallengeIDs = append(quizChallengeIDs, ch.GetID())
			}
		}
		if len(quizChallengeIDs) > 0 {
			quizThunk := r.Loaders.QuizByChallengeIDLoader.LoadMany(ctx, quizChallengeIDs)
			quizResults, _ := quizThunk()
			quizIDs := make([]string, 0, len(quizChallengeIDs))
			for i, q := range quizResults {
				if q != nil {
					quizByChallenge[quizChallengeIDs[i]] = q
					quizIDs = append(quizIDs, q.ID)
				}
			}
			if len(quizIDs) > 0 {
				accessibleQuizIDs, err := r.DB.Queries.GetBulkUserSessionAccessQuizIDs(ctx, sqlc.GetBulkUserSessionAccessQuizIDsParams{
					Quizids: quizIDs,
					Userid:  userID,
				})
				if err == nil {
					for _, qid := range accessibleQuizIDs {
						sessionAccessQuizIDs[qid] = true
					}
				}
			}
		}
	}

	// Batch-load which of this event's challenges the viewer is enrolled in (authenticated only).
	enrolledChallengeIDs := make(map[string]bool)
	if userID != "" && len(challenges) > 0 {
		challengeIDs := make([]string, len(challenges))
		for i, ch := range challenges {
			challengeIDs[i] = ch.GetID()
		}
		enrolled, err := r.DB.Queries.GetUserEnrollmentTimestamps(ctx, sqlc.GetUserEnrollmentTimestampsParams{
			Userid:       userID,
			Challengeids: challengeIDs,
		})
		if err == nil {
			for _, row := range enrolled {
				enrolledChallengeIDs[row.ChallengeID] = true
			}
		}
	}

	result := make([]model.Challenge, 0, len(challenges))
	for _, ch := range challenges {
		// Quiz challenges (authenticated): visible only when session access is granted,
		// regardless of publishedAt/visibleAt.
		if _, ok := ch.(*model.QuizChallenge); ok && userID != "" {
			if quiz := quizByChallenge[ch.GetID()]; quiz != nil && sessionAccessQuizIDs[quiz.ID] {
				result = append(result, ch)
			}
			continue
		}

		publishedAt := getChallengePublishedAt(ch)
		if publishedAt == nil || publishedAt.After(now) {
			continue // Skip unpublished
		}

		visibleAt := getChallengeVisibleAt(ch)
		isVisible := visibleAt != nil && !visibleAt.After(now)
		if !isVisible {
			// Not publicly visible yet: only enrolled (authenticated) users may see it.
			if userID == "" || !enrolledChallengeIDs[ch.GetID()] {
				continue
			}
		}

		result = append(result, ch)
	}

	// Apply translations in batch after filtering.
	r.applyTranslationsToChallenges(ctx, result)
	return result, nil
}

// sortChallengesByEnrollment sorts challenges with enrolled first (by enrolled_at DESC),
// then non-enrolled (by published_at DESC).
func (r *Resolver) sortChallengesByEnrollment(ctx context.Context, userID string, challenges []model.Challenge) {
	if len(challenges) == 0 {
		return
	}

	keys := make([]loaders.UserChallengeKey, len(challenges))
	for i, ch := range challenges {
		keys[i] = loaders.UserChallengeKey{UserID: userID, ChallengeID: ch.GetID()}
	}

	thunk := r.Loaders.UserChallengeEnrollmentTimestampLoader.LoadMany(ctx, keys)
	timestamps, _ := thunk()
	enrollmentTimes := make(map[string]*time.Time, len(challenges))
	for i, ts := range timestamps {
		enrollmentTimes[challenges[i].GetID()] = ts
	}

	sort.Slice(challenges, func(i, j int) bool {
		tsI := enrollmentTimes[challenges[i].GetID()]
		tsJ := enrollmentTimes[challenges[j].GetID()]

		if (tsI != nil) != (tsJ != nil) {
			return tsI != nil
		}
		if tsI != nil && tsJ != nil {
			return tsI.After(*tsJ)
		}
		pubI := challenges[i].GetPublishedAt()
		pubJ := challenges[j].GetPublishedAt()
		if pubI != nil && pubJ != nil {
			return pubI.Time.After(pubJ.Time)
		}
		return pubI != nil
	})
}

// ChallengeType constants
const (
	ChallengeTypeSimple   = "SIMPLE"
	ChallengeTypeQuiz     = "QUIZ"
	ChallengeTypeExternal = "EXTERNAL"
	ChallengeTypePlugin   = "PLUGIN"
)

// buildChallengeFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildChallengeFilterParamsCursor(filter *model.ChallengeFilter, first *int, afterCursor *pagination.ChallengeCursor, last *int, beforeCursor *pagination.ChallengeCursor) (sqlc.GetChallengesFilteredCursorParams, error) {
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

	// Set composite cursors (publishedAt + id)
	if afterCursor != nil && afterCursor.ID != "" {
		params.Aftercursorpublishedat = pgtype.Timestamptz{
			Time:  afterCursor.PublishedAt,
			Valid: true,
		}
		params.Aftercursorid = afterCursor.ID
	}

	if beforeCursor != nil && beforeCursor.ID != "" {
		params.Beforecursorpublishedat = pgtype.Timestamptz{
			Time:  beforeCursor.PublishedAt,
			Valid: true,
		}
		params.Beforecursorid = beforeCursor.ID
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
	case ChallengeTypePlugin:
		pluginChallengeID := ""
		if row.PluginChallengeID != nil {
			pluginChallengeID = *row.PluginChallengeID
		}
		var buttonText *string
		if row.ButtonText != "" {
			buttonText = &row.ButtonText
		}
		return &model.PluginChallenge{
			ID:                          row.ID,
			Name:                        row.Name,
			Description:                 scalars.HTML(row.Description),
			Image:                       row.ImageUrl,
			PluginChallengeID:           pluginChallengeID,
			ButtonText:                  buttonText,
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
		PluginChallengeID:           row.PluginChallengeID,
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
		PluginChallengeID:           row.PluginChallengeID,
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
		PluginChallengeID:           row.PluginChallengeID,
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
		PluginChallengeID:           row.PluginChallengeID,
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
		PluginChallengeID:           row.PluginChallengeID,
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
		PluginChallengeID:           row.PluginChallengeID,
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
		PluginChallengeID:           row.PluginChallengeID,
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
		PluginChallengeID:           row.PluginChallengeID,
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
	case *model.PluginChallenge:
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
	case *model.PluginChallenge:
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
	case *model.PluginChallenge:
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
	case *model.PluginChallenge:
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
	case *model.PluginChallenge:
		return v.EndTime
	default:
		return nil
	}
}

// Helper to get VisibleAt from any Challenge implementation
func getChallengeVisibleAt(c model.Challenge) *scalars.DateTime {
	switch v := c.(type) {
	case *model.SimpleChallenge:
		return v.VisibleAt
	case *model.QuizChallenge:
		return v.VisibleAt
	case *model.ExternalChallenge:
		return v.VisibleAt
	case *model.PluginChallenge:
		return v.VisibleAt
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
	case *model.PluginChallenge:
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
	case *model.PluginChallenge:
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
	case *model.PluginChallenge:
		return model.ChallengeTypePlugin
	default:
		return model.ChallengeTypeSimple
	}
}

// getChallengePushInfo extracts the minimal information needed for push notifications from a Challenge
func getChallengePushInfo(c model.Challenge) push.ChallengeInfo {
	switch v := c.(type) {
	case *model.SimpleChallenge:
		return push.ChallengeInfo{
			ID:               v.ID,
			Name:             v.Name,
			NotificationText: v.NotificationText,
			Image:            derefString(v.Image),
		}
	case *model.QuizChallenge:
		return push.ChallengeInfo{
			ID:               v.ID,
			Name:             v.Name,
			NotificationText: v.NotificationText,
			Image:            derefString(v.Image),
		}
	case *model.ExternalChallenge:
		return push.ChallengeInfo{
			ID:               v.ID,
			Name:             v.Name,
			NotificationText: v.NotificationText,
			Image:            derefString(v.Image),
		}
	case *model.PluginChallenge:
		return push.ChallengeInfo{
			ID:               v.ID,
			Name:             v.Name,
			NotificationText: v.NotificationText,
			Image:            derefString(v.Image),
		}
	default:
		return push.ChallengeInfo{}
	}
}

// derefString safely dereferences a string pointer, returning empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// validateCreateChallengeInput validates type-specific fields for CreateChallengeInput
func validateCreateChallengeInput(input model.CreateChallengeInput) error {
	switch input.Type {
	case model.ChallengeTypeSimple:
		// buttonText is required for SIMPLE challenges
		if input.ButtonText == nil || *input.ButtonText == "" {
			return fmt.Errorf("buttonText is required for SIMPLE challenges")
		}
		// allowSelfCompletion is valid (optional, defaults true)
		if input.URL != nil {
			return fmt.Errorf("url is not allowed for SIMPLE challenges")
		}
		if input.PluginChallengeID != nil {
			return fmt.Errorf("pluginChallengeId is not allowed for SIMPLE challenges")
		}
	case model.ChallengeTypeQuiz:
		// buttonText is required for QUIZ challenges
		if input.ButtonText == nil || *input.ButtonText == "" {
			return fmt.Errorf("buttonText is required for QUIZ challenges")
		}
		if input.AllowSelfCompletion != nil {
			return fmt.Errorf("allowSelfCompletion is not valid for QUIZ challenges")
		}
		if input.URL != nil {
			return fmt.Errorf("url is not valid for QUIZ challenges")
		}
		if input.PluginChallengeID != nil {
			return fmt.Errorf("pluginChallengeId is not valid for QUIZ challenges")
		}
	case model.ChallengeTypeExternal:
		// buttonText is required for EXTERNAL challenges
		if input.ButtonText == nil || *input.ButtonText == "" {
			return fmt.Errorf("buttonText is required for EXTERNAL challenges")
		}
		if input.URL == nil || *input.URL == "" {
			return fmt.Errorf("url is required for EXTERNAL challenges")
		}
		if input.AllowSelfCompletion != nil {
			return fmt.Errorf("allowSelfCompletion is not valid for EXTERNAL challenges")
		}
		if input.PluginChallengeID != nil {
			return fmt.Errorf("pluginChallengeId is not valid for EXTERNAL challenges")
		}
	case model.ChallengeTypePlugin:
		// buttonText is optional for PLUGIN challenges
		if input.PluginChallengeID == nil || *input.PluginChallengeID == "" {
			return fmt.Errorf("pluginChallengeId is required for PLUGIN challenges")
		}
		if input.AllowSelfCompletion != nil {
			return fmt.Errorf("allowSelfCompletion is not valid for PLUGIN challenges")
		}
		if input.URL != nil {
			return fmt.Errorf("url is not valid for PLUGIN challenges")
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
		if input.PluginChallengeID != nil {
			return fmt.Errorf("pluginChallengeId is not allowed for SIMPLE challenges")
		}
	case model.ChallengeTypeQuiz:
		if input.AllowSelfCompletion != nil {
			return fmt.Errorf("allowSelfCompletion is not valid for QUIZ challenges")
		}
		if input.URL != nil {
			return fmt.Errorf("url is not valid for QUIZ challenges")
		}
		if input.PluginChallengeID != nil {
			return fmt.Errorf("pluginChallengeId is not valid for QUIZ challenges")
		}
	case model.ChallengeTypeExternal:
		// url is valid (but must be non-empty if provided)
		if input.URL != nil && *input.URL == "" {
			return fmt.Errorf("url cannot be empty for EXTERNAL challenges")
		}
		if input.AllowSelfCompletion != nil {
			return fmt.Errorf("allowSelfCompletion is not valid for EXTERNAL challenges")
		}
		if input.PluginChallengeID != nil {
			return fmt.Errorf("pluginChallengeId is not valid for EXTERNAL challenges")
		}
	case model.ChallengeTypePlugin:
		// pluginChallengeId is valid (but must be non-empty if provided)
		if input.PluginChallengeID != nil && *input.PluginChallengeID == "" {
			return fmt.Errorf("pluginChallengeId cannot be empty for PLUGIN challenges")
		}
		if input.AllowSelfCompletion != nil {
			return fmt.Errorf("allowSelfCompletion is not valid for PLUGIN challenges")
		}
		if input.URL != nil {
			return fmt.Errorf("url is not valid for PLUGIN challenges")
		}
	}
	return nil
}
