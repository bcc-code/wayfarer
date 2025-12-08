# Quiz System Testing Guide

## Overview

This document provides comprehensive testing procedures for the quiz system implementation in Wayfarer. The quiz system supports multiple question types, configurable timing, answer correctness tracking, and integration with achievements.

## Prerequisites

Before testing, ensure:
1. Database migrations have been run
2. Server is running locally
3. You have test project and user accounts set up
4. You have admin/superadmin authentication tokens

## Test Data Setup

### 1. Create a Test Project

```graphql
mutation CreateTestProject {
  createProject(input: {
    name: "Quiz Test Project"
    description: "Project for testing quiz functionality"
    startDate: "2024-01-01T00:00:00Z"
  }) {
    id
    name
  }
}
```

Save the project ID for subsequent operations.

### 2. Create a Test Event (Optional)

```graphql
mutation CreateTestEvent {
  createEvent(input: {
    projectId: "PR..."
    name: "Quiz Test Event"
    description: "Event for testing quizzes"
    startDate: "2024-01-01T00:00:00Z"
    endDate: "2024-12-31T23:59:59Z"
  }) {
    id
    name
  }
}
```

## Test Scenarios

### Scenario 1: Basic Quiz Creation with Predefined Answers

#### Step 1: Create a Quiz

```graphql
mutation CreateBasicQuiz {
  createQuiz(input: {
    projectId: "PR..."
    name: "Basic Knowledge Quiz"
    description: "Test your knowledge"
    timeoutSeconds: 300
    randomizeQuestions: false
    revealCorrectAnswers: true
    allowRetakes: false
    completionPoints: 100
  }) {
    id
    name
    timeoutSeconds
    completionPoints
  }
}
```

**Expected**: Quiz created with ID starting with "QZ"

#### Step 2: Add Predefined Question (Single Selection)

```graphql
mutation AddPredefinedQuestion {
  addQuizQuestion(
    quizId: "QZ..."
    input: {
      questionType: PREDEFINED
      questionText: "What is the capital of France?"
      questionOrder: 1
      allowMultipleSelection: false
      predefinedAnswers: [
        { answerText: "Paris", isCorrect: true, answerOrder: 1 }
        { answerText: "London", isCorrect: false, answerOrder: 2 }
        { answerText: "Berlin", isCorrect: false, answerOrder: 3 }
        { answerText: "Madrid", isCorrect: false, answerOrder: 4 }
      ]
    }
  ) {
    id
    questionText
    questionType
    predefinedAnswers {
      id
      answerText
      answerOrder
    }
  }
}
```

**Expected**: Question created with 4 predefined answers

#### Step 3: Add Multiple-Selection Question

```graphql
mutation AddMultipleChoiceQuestion {
  addQuizQuestion(
    quizId: "QZ..."
    input: {
      questionType: PREDEFINED
      questionText: "Which of the following are programming languages?"
      questionOrder: 2
      allowMultipleSelection: true
      predefinedAnswers: [
        { answerText: "Python", isCorrect: true, answerOrder: 1 }
        { answerText: "JavaScript", isCorrect: true, answerOrder: 2 }
        { answerText: "HTML", isCorrect: false, answerOrder: 3 }
        { answerText: "Go", isCorrect: true, answerOrder: 4 }
      ]
    }
  ) {
    id
    questionText
    allowMultipleSelection
  }
}
```

**Expected**: Question created with multiple selection enabled

#### Step 4: Publish the Quiz

```graphql
mutation PublishQuiz {
  publishQuiz(id: "QZ...", publishedAt: "2024-01-01T00:00:00Z") {
    id
    publishedAt
  }
}
```

**Expected**: Quiz published with publishedAt timestamp

### Scenario 2: User Taking the Quiz

#### Step 1: Query Available Quiz (as user)

```graphql
query GetQuiz {
  quiz(id: "QZ...") {
    id
    name
    description
    timeoutSeconds
    revealCorrectAnswers
    allowRetakes
    completionPoints
    questions {
      id
      questionText
      questionType
      questionOrder
      allowMultipleSelection
      predefinedAnswers {
        id
        answerText
        answerOrder
      }
    }
    userCanStart
    userActiveSubmission {
      id
    }
  }
}
```

**Expected**:
- User can see quiz details
- Questions are listed
- `userCanStart` is `true`
- Correct answers are NOT visible (isCorrect field should not be exposed to user before completion)

#### Step 2: Start Quiz

```graphql
mutation StartQuiz {
  startQuiz(quizId: "QZ...") {
    id
    startedAt
    expiresAt
    questionOrder
    orderedQuestions {
      id
      questionText
    }
  }
}
```

**Expected**:
- Submission created with ID starting with "QS"
- `startedAt` is current time
- `expiresAt` is startedAt + 300 seconds
- `questionOrder` contains all question IDs
- If `randomizeQuestions` was true, order would be shuffled

#### Step 3: Submit Answer to First Question

```graphql
mutation SubmitAnswer1 {
  submitQuizAnswer(
    submissionId: "QS..."
    input: {
      questionId: "QQ..."
      selectedAnswerIds: ["QA..."]  # Paris
      timeSpentSeconds: 15
    }
  ) {
    id
    question {
      questionText
    }
    selectedAnswers {
      answerText
    }
    isCorrect
  }
}
```

**Expected**:
- Response created with ID starting with "QR"
- `isCorrect` is `true` (Paris is correct)
- Response is linked to submission and question

#### Step 4: Submit Answer to Second Question

```graphql
mutation SubmitAnswer2 {
  submitQuizAnswer(
    submissionId: "QS..."
    input: {
      questionId: "QQ..."
      selectedAnswerIds: ["QA...", "QA...", "QA..."]  # Python, JavaScript, Go
      timeSpentSeconds: 30
    }
  ) {
    id
    selectedAnswers {
      answerText
    }
    isCorrect
  }
}
```

**Expected**:
- All three correct answers selected
- `isCorrect` is `true` (all selected answers are correct)

#### Step 5: Finalize Quiz

```graphql
mutation FinalizeQuiz {
  finalizeQuiz(submissionId: "QS...") {
    id
    completedAt
    score
    maxScore
    scorePercentage
    pointsAwarded
    responses {
      question {
        questionText
      }
      selectedAnswers {
        answerText
        isCorrect
      }
      isCorrect
    }
  }
}
```

**Expected**:
- `completedAt` is set to current time
- `score` is 2 (both questions correct)
- `maxScore` is 2
- `scorePercentage` is 100.0
- `pointsAwarded` is 100 (completionPoints from quiz)
- Now user can see `isCorrect` on answers (if revealCorrectAnswers is true)

#### Step 6: Verify Points Awarded

```graphql
query CheckUserScore {
  user(id: "US...") {
    totalScore
    scoreHistory {
      amount
      description
      createdAt
    }
  }
}
```

**Expected**:
- Score increased by 100 points
- Score journal entry shows quiz completion

### Scenario 3: Test Retake Policy

#### Step 1: Attempt to Start Quiz Again (allowRetakes: false)

```graphql
mutation TryStartAgain {
  startQuiz(quizId: "QZ...") {
    id
  }
}
```

**Expected**: Error - "Quiz does not allow retakes and user has already completed it"

#### Step 2: Update Quiz to Allow Retakes

```graphql
mutation AllowRetakes {
  updateQuiz(id: "QZ...", input: { allowRetakes: true }) {
    id
    allowRetakes
  }
}
```

#### Step 3: Start Quiz Again

```graphql
mutation StartQuizRetake {
  startQuiz(quizId: "QZ...") {
    id
    startedAt
  }
}
```

**Expected**: New submission created successfully

### Scenario 4: Test Timeout Expiry

#### Step 1: Create Quiz with Short Timeout

```graphql
mutation CreateTimedQuiz {
  createQuiz(input: {
    projectId: "PR..."
    name: "Timed Quiz"
    description: "This quiz expires quickly"
    timeoutSeconds: 10
    revealCorrectAnswers: true
    allowRetakes: true
    completionPoints: 50
  }) {
    id
  }
}
```

#### Step 2: Start Quiz and Wait for Expiry

```graphql
mutation StartTimedQuiz {
  startQuiz(quizId: "QZ...") {
    id
    expiresAt
    isExpired
  }
}
```

**Expected**: `isExpired` is false initially

#### Step 3: Wait 11 seconds, then try to submit

```graphql
mutation TrySubmitAfterExpiry {
  submitQuizAnswer(
    submissionId: "QS..."
    input: {
      questionId: "QQ..."
      selectedAnswerIds: ["QA..."]
    }
  ) {
    id
  }
}
```

**Expected**: Error - "Quiz submission has expired"

### Scenario 5: Number Question Type

#### Step 1: Add Number Question (Slider)

```graphql
mutation AddNumberQuestion {
  addQuizQuestion(
    quizId: "QZ..."
    input: {
      questionType: NUMBER
      questionText: "Rate your satisfaction (1-10)"
      questionOrder: 3
      minValue: 1.0
      maxValue: 10.0
      stepValue: 1.0
    }
  ) {
    id
    questionType
    minValue
    maxValue
    stepValue
  }
}
```

#### Step 2: Submit Number Response

```graphql
mutation SubmitNumberAnswer {
  submitQuizAnswer(
    submissionId: "QS..."
    input: {
      questionId: "QQ..."
      numberResponse: 8.0
    }
  ) {
    id
    numberResponse
    isCorrect
  }
}
```

**Expected**:
- Response saved with `numberResponse: 8.0`
- `isCorrect` is `null` (number questions are informational)

### Scenario 6: Free Text Question Type

#### Step 1: Add Free Text Question

```graphql
mutation AddTextQuestion {
  addQuizQuestion(
    quizId: "QZ..."
    input: {
      questionType: FREE_TEXT
      questionText: "What did you learn from this quiz?"
      questionOrder: 4
    }
  ) {
    id
    questionType
  }
}
```

#### Step 2: Submit Text Response

```graphql
mutation SubmitTextAnswer {
  submitQuizAnswer(
    submissionId: "QS..."
    input: {
      questionId: "QQ..."
      textResponse: "I learned about quiz systems and gamification."
    }
  ) {
    id
    textResponse
    isCorrect
  }
}
```

**Expected**:
- Response saved with text
- `isCorrect` is `null` (free text is not auto-graded)

### Scenario 7: JSON Question Type

#### Step 1: Add JSON Question

```graphql
mutation AddJSONQuestion {
  addQuizQuestion(
    quizId: "QZ..."
    input: {
      questionType: JSON
      questionText: "Provide your demographic information"
      questionOrder: 5
    }
  ) {
    id
    questionType
  }
}
```

#### Step 2: Submit JSON Response

```graphql
mutation SubmitJSONAnswer {
  submitQuizAnswer(
    submissionId: "QS..."
    input: {
      questionId: "QQ..."
      jsonResponse: "{\"age\": 25, \"country\": \"Norway\"}"
    }
  ) {
    id
    jsonResponse
    isCorrect
  }
}
```

**Expected**:
- Response saved with JSON data
- `isCorrect` is `null`

### Scenario 8: Quiz Achievement

#### Step 1: Create Quiz Achievement

```graphql
mutation CreateQuizAchievement {
  createQuizAchievement(input: {
    name: "Quiz Master"
    description: "Score 80% or higher on Basic Knowledge Quiz"
    image: "https://example.com/badge.png"
    projectId: "PR..."
    points: 50
    hidden: false
    quizId: "QZ..."
    minScorePercentage: 80
    requireCompletion: true
  }) {
    id
    name
    quiz {
      name
    }
    minScorePercentage
  }
}
```

#### Step 2: Complete Quiz with High Score

Follow Scenario 2 to complete quiz with score ≥ 80%

#### Step 3: Verify Achievement Awarded

```graphql
query CheckAchievement {
  user(id: "US...") {
    achievements {
      id
      name
      achievedAt
      points
    }
  }
}
```

**Expected**:
- "Quiz Master" achievement appears in user's achievements
- `achievedAt` timestamp is set
- User gained 50 additional points from achievement

### Scenario 9: Per-Question Timeout

#### Step 1: Create Quiz with Question Timeout

```graphql
mutation CreateQuestionTimedQuiz {
  createQuiz(input: {
    projectId: "PR..."
    name: "Question Timed Quiz"
    description: "Each question has a time limit"
    questionTimeoutSeconds: 30
    revealCorrectAnswers: true
    allowRetakes: false
    completionPoints: 75
  }) {
    id
    questionTimeoutSeconds
    timeoutSeconds
  }
}
```

**Expected**:
- `questionTimeoutSeconds` is 30
- `timeoutSeconds` is `null` (mutually exclusive)

#### Step 2: Start and Complete Quiz

Follow similar flow to Scenario 2, but note that each question should be answered within 30 seconds.

### Scenario 10: Question Randomization

#### Step 1: Create Quiz with Randomization

```graphql
mutation CreateRandomizedQuiz {
  createQuiz(input: {
    projectId: "PR..."
    name: "Randomized Quiz"
    description: "Questions appear in random order"
    randomizeQuestions: true
    revealCorrectAnswers: true
    allowRetakes: true
    completionPoints: 60
  }) {
    id
    randomizeQuestions
  }
}
```

#### Step 2: Add Multiple Questions

Add 5+ questions following Scenario 1 patterns.

#### Step 3: Start Quiz Multiple Times

```graphql
mutation StartRandomized {
  startQuiz(quizId: "QZ...") {
    id
    questionOrder
    orderedQuestions {
      id
      questionText
    }
  }
}
```

**Expected**:
- Each time quiz is started, `questionOrder` array is in a different sequence
- `orderedQuestions` matches the randomized order

### Scenario 11: Pagination of Quizzes

#### Step 1: Query Quizzes with Pagination

```graphql
query GetQuizzesPaginated {
  quizzes(
    filter: { projectId: "PR..." }
    first: 10
  ) {
    edges {
      cursor
      node {
        id
        name
      }
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      startCursor
      endCursor
    }
    totalCount
  }
}
```

**Expected**:
- Returns up to 10 quizzes
- Proper pagination cursors
- Accurate totalCount

#### Step 2: Fetch Next Page

```graphql
query GetNextPage {
  quizzes(
    filter: { projectId: "PR..." }
    first: 10
    after: "QZ..."  # Use endCursor from previous query
  ) {
    edges {
      node {
        id
        name
      }
    }
    pageInfo {
      hasNextPage
    }
  }
}
```

### Scenario 12: M2M External Submission

```graphql
mutation ExternalQuizSubmission {
  createQuizSubmission(
    quizId: "QZ..."
    userId: "US..."
    completedAt: "2024-06-15T14:30:00Z"
    responses: [
      {
        questionId: "QQ..."
        selectedAnswerIds: ["QA..."]
      }
      {
        questionId: "QQ..."
        textResponse: "External response"
      }
    ]
  ) {
    id
    completedAt
    score
    pointsAwarded
  }
}
```

**Expected**:
- Submission created immediately as completed
- Score calculated from responses
- Points awarded to user

## Edge Cases to Test

### 1. Empty Quiz
- Create quiz with no questions
- Attempt to start it
- Expected: Should work, but submission completes immediately with 0 score

### 2. Incomplete Submission
- Start quiz
- Submit only some answers
- Finalize without answering all questions
- Expected: Score based only on answered questions

### 3. Update Question After Submissions Exist
- Create quiz, publish it
- User completes it
- Admin updates a question
- Expected: Old submissions remain valid, new submissions use updated question

### 4. Delete Quiz with Submissions
- Create quiz
- Users complete it
- Admin deletes quiz
- Expected: Submissions are cascade deleted (verify database schema)

### 5. Quiz Without Correct Answers
- Create quiz with predefined questions
- Set all `isCorrect` to false
- Complete quiz
- Expected: score is 0, maxScore is 0, scorePercentage is 0 or null

### 6. Reveal Correct Answers = False
- Create quiz with `revealCorrectAnswers: false`
- Complete it
- Query submission responses
- Expected: `isCorrect` field on answers should not be visible to user

## Performance Testing

### Load Test Submission Creation
1. Create quiz with 50 questions
2. Start quiz
3. Submit all 50 answers rapidly
4. Finalize
5. Expected: All responses saved, no database deadlocks

### Concurrent Quiz Taking
1. Multiple users (10+) start same quiz simultaneously
2. All submit answers
3. All finalize
4. Expected: All submissions independent, no race conditions

## Database Verification

After testing scenarios, verify database state:

```sql
-- Check quizzes
SELECT * FROM quizzes WHERE project_id = 'PR...';

-- Check questions
SELECT * FROM quiz_questions WHERE quiz_id = 'QZ...';

-- Check predefined answers
SELECT * FROM quiz_predefined_answers WHERE question_id = 'QQ...';

-- Check submissions
SELECT * FROM quiz_submissions WHERE quiz_id = 'QZ...';

-- Check responses
SELECT * FROM quiz_responses WHERE submission_id = 'QS...';

-- Check score journal entries
SELECT * FROM score_journal WHERE user_id = 'US...' AND source_type = 'QUIZ';

-- Check achievements
SELECT * FROM quiz_achievements WHERE quiz_id = 'QZ...';
```

## Known Limitations

1. **Free text and JSON questions are not auto-graded** - This is by design. These responses are informational only.

2. **Timeout enforcement is client-side** - The server checks expiry but doesn't prevent late submissions. Consider adding server-side deadline enforcement.

3. **No partial credit for multi-select** - Either all correct answers selected or question is marked wrong.

4. **Question order changes don't affect existing submissions** - Randomization happens at submission creation time.

## Next Steps

After manual testing verification:

1. Implement integration tests for critical flows
2. Add validation for edge cases
3. Performance test with realistic data volumes
4. Security audit for authorization checks
5. Consider implementing real-time quiz features (leaderboards, live results)
