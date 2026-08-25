package main

// Read-path documents are copied verbatim from the validated k6 corpus
// (backend/cmd/loadtest/k6/queries/*.js) so the comparator replays exactly
// what the loadtest exercised. Admin/setup mutations mirror the shapes used by
// the e2e suite (backend/e2e/quiz_betting_test.go, ordering_questions_test.go).
var queries = map[string]string{
	"GetMe": `
query GetMe {
  me {
    id
    name
    email
    image
    membersId
    language
    church {
      id
      name
      country
      category
    }
    gender
    birthdate
    age
    createdAt
    roles {
      id
      role
      scope {
        id
        type
        church {
          id
        }
        team {
          id
        }
        project {
          id
        }
      }
    }
  }
}
`,

	"CurrentProject": `
query CurrentProject {
  myCurrentProject {
    id
    branding {
      logoImage {
        url
        width
        height
        blurhash
      }
      bannerImage {
        url
        width
        height
        blurhash
      }
      rounding
      colors {
        light {
          accent
          accentContrast
          onAccent
          backgroundDefault
          backgroundRaised
          backgroundIndent
          textDefault
          textMuted
          textHint
          shadowDefault
          shadowBlank
          borderDefault
        }
        dark {
          accent
          accentContrast
          onAccent
          backgroundDefault
          backgroundRaised
          backgroundIndent
          textDefault
          textMuted
          textHint
          shadowDefault
          shadowBlank
          borderDefault
        }
      }
    }
    activeChallengesCount
    leaderboard(entityType: PERSONS, first: 1) {
      totalCount
    }
  }
}
`,

	"GetFirebaseToken": `
query GetFirebaseToken {
  firebaseToken {
    token
    expiresIn
  }
}
`,

	"ProfilePage": `
query ProfilePage($ageFilter: LeaderboardFilter) {
  me {
    id
    name
    consentStatus {
      pendingConsents {
        __typename
        id
        key
        version
        title
        body {
          html
        }
        shortText
        url
        managementType
        managedBy
      }
    }
  }
  myCurrentProject {
    id
    name
    infoMessage {
      markdown
      html
    }
    infoMessageStart
    infoMessageEnd
    achievements {
      __typename
      id
      name
      descriptionPending
      descriptionCompleted
      hidden
      achievedAt
      celebratedAt
      points
    }
    myPoints
    leaderboard(entityType: PERSONS, filter: $ageFilter) {
      me {
        rank
      }
    }
    myTeam {
      superTeam {
        id
        name
        color
        imageObject {
          url
          blurhash
        }
      }
    }
  }
}
`,

	"ActiveChallengesPage": `
query ActiveChallengesPage {
  myCurrentProject {
    myTeam {
      joinCode
    }
    activeChallenges {
      __typename
      id
      name
      description
      buttonText
      notificationText
      publishedAt
      visibleAt
      startedAt
      endTime
      requiresTeamMembership
      requiresSuperTeamMembership
      userCompletedAt
      userEnrolledAt
      ... on SimpleChallenge {
        allowSelfCompletion
      }
      ... on ExternalChallenge {
        url
      }
      ... on QuizChallenge {
        quiz {
          timeoutSeconds
          randomizeQuestions
          revealCorrectAnswers
          allowRetakes
          completionPoints
          endTime
          userCanStart
          userActiveSubmission {
            id
            startedAt
            completedAt
            expiresAt
            isExpired
            autoSubmitted
            score
            maxScore
            scorePercentage
            pointsAwarded
          }
          userActiveSession {
            id
            state
            openAt
            lockAt
            finishAt
            userHasAccess
          }
          userSubmissions {
            id
            startedAt
            completedAt
            expiresAt
            isExpired
            autoSubmitted
            score
            maxScore
            scorePercentage
            pointsAwarded
          }
        }
      }
    }
  }
}
`,

	"CompletedChallengesPage": `
query CompletedChallengesPage {
  myCurrentProject {
    completedChallenges {
      __typename
      id
      name
      description
      buttonText
      notificationText
      publishedAt
      visibleAt
      startedAt
      endTime
      requiresTeamMembership
      requiresSuperTeamMembership
      userCompletedAt
      userEnrolledAt
      ... on SimpleChallenge {
        allowSelfCompletion
      }
      ... on ExternalChallenge {
        url
      }
      ... on QuizChallenge {
        quiz {
          timeoutSeconds
          randomizeQuestions
          revealCorrectAnswers
          allowRetakes
          completionPoints
          endTime
          userCanStart
          userActiveSubmission {
            id
            score
            maxScore
            pointsAwarded
          }
          userSubmissions {
            id
            score
            maxScore
            pointsAwarded
          }
        }
      }
    }
  }
}
`,

	"ChallengePage": `
query ChallengePage($challengeId: ID!) {
  myCurrentProject {
    myPoints
  }
  challenge(id: $challengeId) {
    __typename
    id
    name
    description
    requiresTeamMembership
    requiresSuperTeamMembership
    userEnrolledAt
    userCompletedAt
    ... on SimpleChallenge {
      allowSelfCompletion
    }
    ... on PluginChallenge {
      pluginChallengeId
    }
    ... on ExternalChallenge {
      url
    }
    ... on QuizChallenge {
      quiz {
        id
        name
        description
        timeoutSeconds
        randomizeQuestions
        revealCorrectAnswers
        allowRetakes
        completionPoints
        endTime
        userCanStart
        userActiveSubmission {
          id
        }
        userActiveSession {
          id
          state
        }
        userSubmissions {
          id
          startedAt
          completedAt
          expiresAt
          isExpired
          score
          maxScore
          scorePercentage
          pointsAwarded
          responses {
            __typename
            id
            answeredAt
            timeSpentSeconds
            betAmount
            pointsEarned
            question {
              id
            }
            ... on FreeTextResponse {
              textResponse
            }
            ... on NumberResponse {
              numberResponse
            }
            ... on PredefinedResponse {
              isCorrect
              selectedAnswers {
                id
                answerText
                answerOrder
                isCorrect
              }
            }
            ... on OrderingResponse {
              isCorrect
              submittedOrder
            }
          }
        }
      }
    }
  }
}
`,

	"EnrollInChallenge": `
mutation EnrollInChallenge($challengeId: ID!) {
  enrollInChallenge(challengeId: $challengeId) {
    id
    userEnrolledAt
  }
}
`,

	"GetQuiz": `
query GetQuiz($id: ID!) {
    quiz(id: $id) {
        id
        name
        description
        timeoutSeconds
        randomizeQuestions
        revealCorrectAnswers
        allowRetakes
        completionPoints
        userCanStart
        userActiveSubmission {
            id
            completedAt
        }
        questions {
            id
            questionText
            questionOrder
            ... on PredefinedQuestion {
                allowMultipleSelection
                predefinedAnswers {
                    id
                    answerText
                    answerOrder
                }
            }
        }
    }
}
`,

	"StartQuiz": `
mutation StartQuiz($quizId: ID!) {
    startQuiz(quizId: $quizId) {
        id
        startedAt
        expiresAt
        questionOrder
        orderedQuestions {
            id
            questionText
            ... on PredefinedQuestion {
                predefinedAnswers {
                    id
                    answerText
                    answerOrder
                }
            }
        }
    }
}
`,

	"StartQuizSession": `
mutation StartQuizSession($sessionId: ID!) {
    startQuizSession(sessionId: $sessionId) {
        id
        startedAt
        expiresAt
        completedAt
        orderedQuestions {
            __typename
            id
            questionText
            questionOrder
            ... on PredefinedQuestion {
                allowMultipleSelection
                predefinedAnswers {
                    id
                    answerOrder
                }
            }
            ... on NumberQuestion {
                minValue
                maxValue
                stepValue
            }
            ... on OrderingQuestion {
                orderingItems {
                    id
                    itemText
                }
            }
        }
    }
}
`,

	"SubmitQuizAnswer": `
mutation SubmitQuizAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
    submitQuizAnswer(submissionId: $submissionId, input: $input) {
        id
        answeredAt
        betAmount
        ... on FreeTextResponse {
            textResponse
        }
        ... on NumberResponse {
            numberResponse
        }
        ... on PredefinedResponse {
            selectedAnswerIds
            isCorrect
        }
        ... on OrderingResponse {
            isCorrect
            submittedOrder
        }
    }
}
`,

	"FinalizeQuiz": `
mutation FinalizeQuiz($submissionId: ID!) {
    finalizeQuiz(submissionId: $submissionId) {
        id
        completedAt
        score
        maxScore
        scorePercentage
        pointsAwarded
    }
}
`,

	"StandingsGlobalPage": `
query StandingsGlobalPage($entityType: LeaderboardEntityType!, $filter: LeaderboardFilter, $first: Int) {
  myCurrentProject {
    id
    leaderboard(entityType: $entityType, filter: $filter, first: $first) {
      edges {
        node {
          id
          name
          description
          score
          rank
          tags
        }
      }
      me {
        id
        name
        description
        score
        rank
        tags
      }
    }
  }
}
`,

	"StandingsLocalPage": `
query StandingsLocalPage($filter: LeaderboardFilter, $first: Int) {
  me {
    church {
      id
      name
    }
  }
  myCurrentProject {
    id
    personLeaderboard: leaderboard(entityType: PERSONS, filter: $filter, first: $first) {
      totalCount
      edges {
        node {
          id
          name
          score
          rank
          tags
        }
      }
      me {
        id
        name
        score
        rank
        tags
      }
    }
    unitLeaderboard: leaderboard(entityType: TEAMS, filter: $filter, first: $first) {
      totalCount
      edges {
        node {
          id
          name
          score
          rank
          tags
        }
      }
      me {
        id
        name
        score
        rank
        tags
      }
    }
  }
}
`,

	"StandingsUnitPage": `
query StandingsUnitPage {
  myCurrentProject {
    id
    myTeam {
      id
      name
      memberLeaderboard {
        id
        name
        score
        rank
        tags
      }
      superTeam {
        id
        name
        color
        imageObject {
          url
          blurhash
        }
      }
    }
  }
}
`,

	"StandingsPage": `
query StandingsPage {
  myCurrentProject {
    myTeam {
      id
    }
  }
}
`,

	"LeaderboardPageForward": `
query LeaderboardPageForward($entityType: LeaderboardEntityType!, $first: Int, $after: String) {
  myCurrentProject {
    id
    leaderboard(entityType: $entityType, first: $first, after: $after) {
      totalCount
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          name
          score
          rank
        }
      }
    }
  }
}
`,

	// ---- admin / setup mutations (shapes from the e2e suite) ----

	"CreateChallenge": `
mutation CreateChallenge($projectId: ID!, $eventId: ID!, $input: CreateChallengeInput!) {
  createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
    id
  }
}
`,

	"PublishChallenge": `
mutation PublishChallenge($id: ID!, $publishedAt: DateTime!) {
  publishChallenge(id: $id, publishedAt: $publishedAt) {
    id
  }
}
`,

	"SetChallengeVisibility": `
mutation SetChallengeVisibility($id: ID!, $visibleAt: DateTime!) {
  setChallengeVisibility(id: $id, visibleAt: $visibleAt) {
    id
  }
}
`,

	"CreateQuiz": `
mutation CreateQuiz($input: CreateQuizInput!) {
  createQuiz(input: $input) {
    id
  }
}
`,

	"AddQuizQuestion": `
mutation AddQuizQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
  addQuizQuestion(quizId: $quizId, input: $input) {
    ... on PredefinedQuestion {
      id
      questionText
      bettingEnabled
      bettingMinAbsolute
      bettingMaxAbsolute
      predefinedAnswers {
        id
        answerText
        answerOrder
        isCorrect
      }
    }
    ... on OrderingQuestion {
      id
      questionText
      points
      orderingItems {
        id
        itemText
      }
    }
  }
}
`,

	"CreateQuizSession": `
mutation CreateQuizSession($input: CreateQuizSessionInput!) {
  createQuizSession(input: $input) {
    id
  }
}
`,

	"GrantQuizSessionAccess": `
mutation GrantQuizSessionAccess($input: GrantQuizSessionAccessInput!) {
  grantQuizSessionAccess(input: $input)
}
`,

	"OpenQuizSession": `
mutation OpenQuizSession($id: ID!) {
  openQuizSession(id: $id) {
    id
  }
}
`,

	"CreateScoreAdjustment": `
mutation CreateScoreAdjustment($input: CreateScoreAdjustmentInput!) {
  createScoreAdjustment(input: $input) {
    id
    points
  }
}
`,

	"UpdateProjectInfoMessage": `
mutation UpdateProjectInfoMessage($id: ID!, $input: UpdateProjectInput!) {
  updateProject(id: $id, input: $input) {
    id
    infoMessage {
      markdown
    }
  }
}
`,

	"ProjectInfoProbe": `
query ProjectInfoProbe {
  myCurrentProject {
    id
    infoMessage {
      markdown
    }
  }
}
`,

	"LeakProbe": `
query LeakProbe {
  myCurrentProject {
    id
    leaderboard(entityType: PERSONS, first: 10) {
      me {
        id
        name
        score
        rank
      }
      edges {
        node {
          id
        }
      }
    }
  }
}
`,

	"VarCollisionProbe": `
query VarCollisionProbe($first: Int!) {
  myCurrentProject {
    id
    leaderboard(entityType: PERSONS, first: $first) {
      edges {
        node {
          id
        }
      }
    }
  }
}
`,
}
