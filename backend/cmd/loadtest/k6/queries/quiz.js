import { graphqlRequest, checkGraphQLResponse, parseResponse } from '../lib/graphql.js';

// Query to get quiz details including all questions and their answers
const GET_QUIZ_QUERY = `
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
`;

// Mutation to start a quiz (creates a submission)
const START_QUIZ_MUTATION = `
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
`;

// Mutation to start a quiz session (the current, session-based flow:
// Quiz.userActiveSession -> startQuizSession -> submitQuizAnswer -> finalizeQuiz)
const START_QUIZ_SESSION_MUTATION = `
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
        }
    }
}
`;

// Mutation to submit a free-text answer
const SUBMIT_TEXT_ANSWER_MUTATION = `
mutation SubmitQuizAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
    submitQuizAnswer(submissionId: $submissionId, input: $input) {
        id
        answeredAt
        ... on FreeTextResponse {
            textResponse
        }
    }
}
`;

// Mutation to submit a number answer
const SUBMIT_NUMBER_ANSWER_MUTATION = `
mutation SubmitQuizAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
    submitQuizAnswer(submissionId: $submissionId, input: $input) {
        id
        answeredAt
        ... on NumberResponse {
            numberResponse
        }
    }
}
`;

// Mutation to submit an answer
const SUBMIT_ANSWER_MUTATION = `
mutation SubmitQuizAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
    submitQuizAnswer(submissionId: $submissionId, input: $input) {
        id
        answeredAt
        ... on PredefinedResponse {
            selectedAnswerIds
            isCorrect
        }
    }
}
`;

// Mutation to finalize the quiz
const FINALIZE_QUIZ_MUTATION = `
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
`;

/**
 * Get quiz details
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} quizId - Quiz ID to fetch
 * @returns {object|null} Quiz data or null on error
 */
export function getQuiz(baseUrl, token, quizId) {
    const response = graphqlRequest(baseUrl, GET_QUIZ_QUERY, { id: quizId }, token, 'GetQuiz');
    if (checkGraphQLResponse(response, 'GetQuiz')) {
        const data = parseResponse(response);
        return data ? data.quiz : null;
    }
    return null;
}

/**
 * Start a quiz submission
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} quizId - Quiz ID to start
 * @returns {object|null} Submission data or null on error
 */
export function startQuiz(baseUrl, token, quizId) {
    const response = graphqlRequest(baseUrl, START_QUIZ_MUTATION, { quizId }, token, 'StartQuiz');
    if (checkGraphQLResponse(response, 'StartQuiz')) {
        const data = parseResponse(response);
        return data ? data.startQuiz : null;
    }
    return null;
}

/**
 * Start a quiz session, creating (or returning) the user's submission
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} sessionId - Quiz session ID (from Quiz.userActiveSession)
 * @returns {object|null} Submission data or null on error
 */
export function startQuizSession(baseUrl, token, sessionId) {
    const response = graphqlRequest(baseUrl, START_QUIZ_SESSION_MUTATION, { sessionId }, token, 'StartQuizSession');
    if (checkGraphQLResponse(response, 'StartQuizSession')) {
        const data = parseResponse(response);
        return data ? data.startQuizSession : null;
    }
    return null;
}

/**
 * Submit a free-text answer for a question
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} submissionId - Submission ID
 * @param {string} questionId - Question ID
 * @param {string} text - Free-text answer
 * @param {number} timeSpentSeconds - Time spent on the question
 * @returns {object|null} Response data or null on error
 */
export function submitTextAnswer(baseUrl, token, submissionId, questionId, text, timeSpentSeconds) {
    const input = {
        questionId,
        textResponse: text,
        timeSpentSeconds,
    };
    const response = graphqlRequest(
        baseUrl,
        SUBMIT_TEXT_ANSWER_MUTATION,
        { submissionId, input },
        token,
        'SubmitQuizAnswer'
    );
    if (checkGraphQLResponse(response, 'SubmitQuizAnswer')) {
        const data = parseResponse(response);
        return data ? data.submitQuizAnswer : null;
    }
    return null;
}

/**
 * Submit a number answer for a question
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} submissionId - Submission ID
 * @param {string} questionId - Question ID
 * @param {number} value - Numeric answer
 * @param {number} timeSpentSeconds - Time spent on the question
 * @returns {object|null} Response data or null on error
 */
export function submitNumberAnswer(baseUrl, token, submissionId, questionId, value, timeSpentSeconds) {
    const input = {
        questionId,
        numberResponse: value,
        timeSpentSeconds,
    };
    const response = graphqlRequest(
        baseUrl,
        SUBMIT_NUMBER_ANSWER_MUTATION,
        { submissionId, input },
        token,
        'SubmitQuizAnswer'
    );
    if (checkGraphQLResponse(response, 'SubmitQuizAnswer')) {
        const data = parseResponse(response);
        return data ? data.submitQuizAnswer : null;
    }
    return null;
}

/**
 * Submit an answer for a question
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} submissionId - Submission ID
 * @param {string} questionId - Question ID
 * @param {string[]} selectedAnswerIds - Array of selected answer IDs
 * @param {number} timeSpentSeconds - Time spent on the question
 * @returns {object|null} Response data or null on error
 */
export function submitAnswer(baseUrl, token, submissionId, questionId, selectedAnswerIds, timeSpentSeconds) {
    const input = {
        questionId,
        selectedAnswerIds,
        timeSpentSeconds,
    };
    const response = graphqlRequest(
        baseUrl,
        SUBMIT_ANSWER_MUTATION,
        { submissionId, input },
        token,
        'SubmitQuizAnswer'
    );
    if (checkGraphQLResponse(response, 'SubmitQuizAnswer')) {
        const data = parseResponse(response);
        return data ? data.submitQuizAnswer : null;
    }
    return null;
}

/**
 * Finalize a quiz submission
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} submissionId - Submission ID to finalize
 * @returns {object|null} Finalized submission data or null on error
 */
export function finalizeQuiz(baseUrl, token, submissionId) {
    const response = graphqlRequest(baseUrl, FINALIZE_QUIZ_MUTATION, { submissionId }, token, 'FinalizeQuiz');
    if (checkGraphQLResponse(response, 'FinalizeQuiz')) {
        const data = parseResponse(response);
        return data ? data.finalizeQuiz : null;
    }
    return null;
}

/**
 * Complete quiz flow: get quiz details, start, answer all questions, finalize
 * This function simulates a user taking a complete quiz
 *
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} quizId - Quiz ID to complete
 * @param {boolean} correctAnswers - Whether to select correct answers (first answer) or random
 * @returns {object} Result with success status and submission data
 */
export function completeQuizFlow(baseUrl, token, quizId, correctAnswers = true) {
    // Step 1: Get quiz details
    const quiz = getQuiz(baseUrl, token, quizId);
    if (!quiz) {
        return { success: false, error: 'Failed to get quiz details' };
    }

    // Check if user can start (hasn't already completed)
    if (!quiz.userCanStart) {
        return { success: true, skipped: true, reason: 'User already completed quiz' };
    }

    // Step 2: Start the quiz
    const submission = startQuiz(baseUrl, token, quizId);
    if (!submission) {
        return { success: false, error: 'Failed to start quiz' };
    }

    // Step 3: Answer all questions
    const questions = submission.orderedQuestions;
    for (let i = 0; i < questions.length; i++) {
        const question = questions[i];

        // Get predefined answers if available
        const answers = question.predefinedAnswers;
        if (!answers || answers.length === 0) {
            continue; // Skip non-predefined questions for now
        }

        // Select answer: first one (correct) or random
        let selectedAnswerId;
        if (correctAnswers) {
            // First answer is correct in our test data
            selectedAnswerId = answers[0].id;
        } else {
            // Random answer
            const randomIndex = Math.floor(Math.random() * answers.length);
            selectedAnswerId = answers[randomIndex].id;
        }

        // Simulate thinking time (0.5-2 seconds)
        const timeSpent = Math.floor(Math.random() * 1500) + 500;

        const answerResult = submitAnswer(
            baseUrl,
            token,
            submission.id,
            question.id,
            [selectedAnswerId],
            Math.floor(timeSpent / 1000)
        );

        if (!answerResult) {
            return {
                success: false,
                error: `Failed to submit answer for question ${i + 1}`,
                submissionId: submission.id,
            };
        }
    }

    // Step 4: Finalize the quiz
    const finalResult = finalizeQuiz(baseUrl, token, submission.id);
    if (!finalResult) {
        return { success: false, error: 'Failed to finalize quiz', submissionId: submission.id };
    }

    return {
        success: true,
        submission: finalResult,
        score: finalResult.score,
        maxScore: finalResult.maxScore,
        pointsAwarded: finalResult.pointsAwarded,
    };
}
