# Quiz System Implementation Summary

## Completed Components

### 1. Database Schema (✅ Complete)
- **Migrations**: `00034_add_quiz_system.sql`, `00035_extend_achievement_types.sql`
- **Tables Created**:
  - `quizzes` - Main quiz configuration
  - `quiz_questions` - Questions with type (PREDEFINED, FREE_TEXT, NUMBER, JSON)
  - `quiz_predefined_answers` - Answer options for predefined questions
  - `quiz_submissions` - User quiz attempts with question order and scoring
  - `quiz_responses` - Individual question responses
  - `quiz_achievements` - Achievement integration
  - `quiz_translations`, `quiz_question_translations`, `quiz_answer_translations` - i18n support

### 2. ID Generation (✅ Complete)
- **File**: `backend/internal/ulid/ulid.go`
- **Prefixes**: QZ (Quiz), QQ (Question), QA (Answer), QS (Submission), QR (Response)
- **Functions**: `NewQuizID()`, `NewQuizQuestionID()`, `NewQuizAnswerID()`, `NewQuizSubmissionID()`, `NewQuizResponseID()`
- **Validation**: `IsQuizID()`, etc.

### 3. SQLC Queries (✅ Complete)
All queries use batch operations and named parameters (no SQL in FOR loops):
- **`queries/quizzes.sql`**: CRUD + filtering with cursor pagination
- **`queries/quiz_questions.sql`**: Questions + predefined answers management
- **`queries/quiz_submissions.sql`**: Submission tracking with user/quiz filtering
- **`queries/quiz_responses.sql`**: Response CRUD + score calculation
- **`queries/quiz_achievements.sql`**: Achievement linkage

### 4. GraphQL Schema (✅ Complete)
- **Files**: `gql/quizzes.graphqls`, `gql/shared.graphqls`
- **Types**: Quiz, QuizQuestion, QuizPredefinedAnswer, QuizSubmission, QuizResponse, QuizAchievement
- **Enums**: QuizQuestionType (PREDEFINED, FREE_TEXT, NUMBER, JSON)
- **Inputs**: CreateQuizInput, UpdateQuizInput, CreateQuizQuestionInput, SubmitQuizAnswerInput, QuizFilter
- **Pagination**: QuizConnection, QuizSubmissionConnection
- **Scalar**: JSON (for JSON question responses)

### 5. Cache Functions (✅ Complete)
- **File**: `backend/internal/cache/keys.go`, `backend/internal/cache/invalidation.go`
- **Key Builders**: `QuizKey()`, `QuizQuestionsByQuizKey()`, `QuizAnswersByQuestionKey()`, etc.
- **Invalidation**: `InvalidateQuiz()`, `InvalidateQuizSubmission()`

### 6. Data Loaders (✅ Complete)
All loaders use batch queries with caching to prevent N+1 problems:
- **`loaders/quiz_by_id.go`**: Batch load quizzes by IDs
- **`loaders/quiz_questions_by_quiz.go`**: Batch load questions per quiz
- **`loaders/quiz_answers_by_question.go`**: Batch load predefined answers per question
- **`loaders/quiz_submissions_by_user.go`**: Batch load submissions per user
- **`loaders/quiz_responses_by_submission.go`**: Batch load responses per submission
- **Wired in**: `loaders/loaders.go` NewLoaders() function

## Implementation Patterns to Follow

### Admin Resolvers Pattern

```go
// Example: CreateQuiz
func (r *mutationResolver) CreateQuiz(ctx context.Context, input model.CreateQuizInput) (*model.Quiz, error) {
    // 1. Authenticate
    userID, ok := middleware.GetUserID(ctx)
    if !ok {
        return nil, fmt.Errorf("user not authenticated")
    }

    // 2. Authorize
    if !r.RoleService.CanManageProject(ctx, userID, input.ProjectID) {
        return nil, fmt.Errorf("unauthorized")
    }

    // 3. Validate business rules
    if input.TimeoutSeconds == nil && input.QuestionTimeoutSeconds == nil {
        return nil, fmt.Errorf("must provide either timeoutSeconds or questionTimeoutSeconds")
    }

    // 4. Generate ID
    quizID := ulid.NewQuizID()

    // 5. Build query params
    params := sqlc.CreateQuizParams{
        ID: quizID,
        Projectid: input.ProjectID,
        // ... map all fields
    }

    // 6. Execute query
    row, err := r.DB.Queries.CreateQuiz(ctx, params)
    if err != nil {
        return nil, fmt.Errorf("failed to create quiz: %w", err)
    }

    // 7. Invalidate cache
    r.Cache.InvalidateProject(input.ProjectID)

    // 8. Convert and return
    return convertRowToQuiz(row), nil
}
```

### User Quiz Taking Pattern

```go
// Example: StartQuiz
func (r *mutationResolver) StartQuiz(ctx context.Context, quizID string) (*model.QuizSubmission, error) {
    // 1. Load quiz using data loader
    quiz, err := r.Loaders.QuizByIDLoader.Load(ctx, quizID)()

    // 2. Check retake policy
    if !quiz.AllowRetakes {
        existing, _ := r.DB.Queries.GetCompletedSubmissionsByUserAndQuiz(ctx, userID, quizID)
        if len(existing) > 0 {
            return nil, fmt.Errorf("retakes not allowed")
        }
    }

    // 3. Check for active submission
    active, _ := r.DB.Queries.GetActiveSubmissionByUserAndQuiz(ctx, userID, quizID)
    if active != nil {
        return convertRowToSubmission(active), nil
    }

    // 4. Load questions using data loader
    questions, _ := r.Loaders.QuizQuestionsByQuizLoader.Load(ctx, quizID)()

    // 5. Build question order (randomize if needed)
    questionOrder := make([]string, len(questions))
    for i, q := range questions {
        questionOrder[i] = q.ID
    }
    if quiz.RandomizeQuestions {
        rand.Shuffle(len(questionOrder), func(i, j int) {
            questionOrder[i], questionOrder[j] = questionOrder[j], questionOrder[i]
        })
    }

    // 6. Calculate expiry
    var expiresAt *time.Time
    if quiz.TimeoutSeconds != nil {
        exp := time.Now().Add(time.Duration(*quiz.TimeoutSeconds) * time.Second)
        expiresAt = &exp
    }

    // 7. Create submission
    questionOrderJSON, _ := json.Marshal(questionOrder)
    maxScore := countGradableQuestions(questions)

    submission, err := r.DB.Queries.CreateQuizSubmission(ctx, sqlc.CreateQuizSubmissionParams{
        ID: ulid.NewQuizSubmissionID(),
        Quizid: quizID,
        Userid: userID,
        Expiresat: expiresAt,
        Questionorder: questionOrderJSON,
        Maxscore: int32(maxScore),
    })

    return convertRowToSubmission(submission), nil
}
```

### Field Resolver Pattern (Using Data Loaders)

```go
// Example: Quiz.Questions resolver
func (r *quizResolver) Questions(ctx context.Context, obj *model.Quiz) ([]*model.QuizQuestion, error) {
    // Use data loader to batch load questions
    return r.Loaders.QuizQuestionsByQuizLoader.Load(ctx, obj.ID)()
}

// Example: QuizQuestion.PredefinedAnswers resolver
func (r *quizQuestionResolver) PredefinedAnswers(ctx context.Context, obj *model.QuizQuestion) ([]*model.QuizPredefinedAnswer, error) {
    if obj.QuestionType != model.QuizQuestionTypePredefined {
        return []*model.QuizPredefinedAnswer{}, nil
    }
    // Use data loader to batch load answers
    return r.Loaders.QuizAnswersByQuestionLoader.Load(ctx, obj.ID)()
}

// Example: QuizPredefinedAnswer.IsCorrect resolver (respects reveal setting)
func (r *quizPredefinedAnswerResolver) IsCorrect(ctx context.Context, obj *model.QuizPredefinedAnswer) (*bool, error) {
    // Load question to get quiz ID
    question, err := r.Loaders.QuizQuestionByIDLoader.Load(ctx, obj.QuestionID)()

    // Load quiz to check reveal setting
    quiz, err := r.Loaders.QuizByIDLoader.Load(ctx, question.QuizID)()

    // Check if user has completed the quiz
    userID, _ := middleware.GetUserID(ctx)
    submissions, _ := r.Loaders.QuizSubmissionsByUserLoader.Load(ctx, userID)()

    var userCompleted bool
    for _, sub := range submissions {
        if sub.QuizID == quiz.ID && sub.CompletedAt != nil {
            userCompleted = true
            break
        }
    }

    // Only reveal if setting allows and user completed
    if quiz.RevealCorrectAnswers && userCompleted {
        return &obj.IsCorrectValue, nil
    }
    return nil, nil
}
```

### FinalizeQuiz Pattern (Transaction with Score Calculation)

```go
func (r *mutationResolver) FinalizeQuiz(ctx context.Context, submissionID string) (*model.QuizSubmission, error) {
    // 1. Load and validate submission
    submission, err := r.DB.Queries.GetQuizSubmissionByID(ctx, submissionID)

    // 2. Verify ownership and not expired
    if submission.UserID != userID {
        return nil, fmt.Errorf("unauthorized")
    }
    if submission.ExpiresAt != nil && time.Now().After(*submission.ExpiresAt) {
        return nil, fmt.Errorf("submission expired")
    }

    // 3. Begin transaction
    tx, err := r.DB.Pool.Begin(ctx)
    defer tx.Rollback(ctx)
    qtx := r.DB.Queries.WithTx(tx)

    // 4. Calculate score (single query, no loop)
    scoreResult, err := qtx.CalculateSubmissionScore(ctx, submissionID)
    score := int(scoreResult.Score)
    maxScore := int(scoreResult.MaxScore)

    // 5. Load quiz for points
    quiz, _ := r.Loaders.QuizByIDLoader.Load(ctx, submission.QuizID)()

    // 6. Update submission
    updatedSubmission, err := qtx.UpdateQuizSubmission(ctx, sqlc.UpdateQuizSubmissionParams{
        ID: submissionID,
        Completedat: time.Now(),
        Score: int32(score),
        Pointsawarded: int32(quiz.CompletionPoints),
    })

    // 7. Create score journal entry if points > 0
    if quiz.CompletionPoints > 0 {
        qtx.CreateScoreJournalEntry(ctx, ...)
    }

    // 8. Check quiz achievements (batch load, not in loop)
    achievements, _ := qtx.GetQuizAchievementsByQuizID(ctx, quiz.ID)
    for _, ach := range achievements {
        scorePercentage := (float64(score) / float64(maxScore)) * 100
        if scorePercentage >= float64(ach.MinScorePercentage) {
            qtx.AwardUserAchievement(ctx, userID, ach.AchievementID)
        }
    }

    // 9. Commit
    tx.Commit(ctx)

    // 10. Invalidate caches
    r.Cache.InvalidateQuizSubmission(submissionID)
    r.Cache.InvalidateProject(quiz.ProjectID)

    return convertRowToSubmission(updatedSubmission), nil
}
```

## Key Design Decisions

1. **Timeout Strategy**: Mutually exclusive quiz-level OR question-level timeout
2. **Question Randomization**: Static order stored in quiz_questions, per-submission order in quiz_submissions.question_order JSONB
3. **Response Storage**: Single table with type-specific columns (simpler than multiple junction tables)
4. **Grading Policy**: FREE_TEXT and JSON are NOT auto-graded (`is_correct = NULL`)
5. **Retake Policy**: Boolean flag (single vs unlimited attempts)
6. **Scoring vs Points**: Separate concepts - score = accuracy, completion_points = gamification reward
7. **Achievement Integration**: QuizAchievement type follows StreakAchievement pattern

## Testing Checklist

- [ ] Create quiz with quiz-level timeout
- [ ] Create quiz with question-level timeout
- [ ] Add predefined questions (single and multi-select)
- [ ] Add free text and number questions
- [ ] Start quiz (check question randomization)
- [ ] Submit answers for all question types
- [ ] Finalize quiz (check score calculation)
- [ ] Test retake policy (single attempt should reject second start)
- [ ] Test unlimited retakes
- [ ] Test quiz timeout expiry
- [ ] Test achievement auto-awarding based on score percentage
- [ ] Test reveal_correct_answers setting
- [ ] Verify all data loaders batch correctly (no N+1 queries)
- [ ] Verify cache invalidation on mutations

## Next Steps for Full Implementation

1. Create resolver stub files (gqlgen may have already generated these)
2. Implement admin resolvers (CreateQuiz, UpdateQuiz, AddQuizQuestion, etc.)
3. Implement user resolvers (StartQuiz, SubmitQuizAnswer, FinalizeQuiz)
4. Implement field resolvers with data loaders
5. Add proper error handling and validation
6. Write unit tests for business logic
7. Write integration tests for full quiz flow

## Challenge Visibility vs. Active/Completed Split (2026-09)

A quiz challenge is *visible* to a user whenever they have access to a session
in a visible state (`OPEN`, `LOCKED`, `FINISHED`) — FINISHED stays visible so
users can review their results. The active/completed split in
`getFilteredChallenges` additionally treats a quiz challenge as **completed**
when the user's visible sessions are all FINISHED (no OPEN/LOCKED session
left), even without a `user_challenge_completions` row. This covers
auto-submitted submissions and users who never played before the session
closed — previously such challenges sat in the Active tab forever.

Mechanism: `GetBulkUsersSessionAccessQuizIDsByProject` returns
`has_live_session` per (user, quiz); the `UserAccessibleQuizIDsLoader` map
value carries it (presence in the map = access, value = live session).
