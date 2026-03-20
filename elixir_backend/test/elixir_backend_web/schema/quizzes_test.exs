defmodule ElixirBackendWeb.Schema.QuizzesTest do
  use ElixirBackendWeb.ConnCase, async: true

  import ElixirBackend.TestHelpers

  setup do
    project = create_project()
    user = create_user()
    quiz = create_quiz(project)
    %{project: project, user: user, quiz: quiz}
  end

  # ── Queries ──

  describe "quiz query" do
    test "returns quiz by id", %{conn: conn, user: user, quiz: quiz} do
      conn = auth_conn(conn, user.id, ["user"])

      resp =
        graphql_query(
          conn,
          """
            query($id: ID!) {
              quiz(id: $id) {
                id
                name
                description
                completionPoints
                randomizeQuestions
                revealCorrectAnswers
                allowRetakes
                questions { ... on FreeTextQuestion { id questionText } }
                sessions { id state }
              }
            }
          """,
          %{"id" => quiz.id}
        )

      data = json_response(resp, 200)["data"]["quiz"]
      assert data["id"] == quiz.id
      assert data["name"] == quiz.name
    end
  end

  describe "quizzes query" do
    test "returns paginated quizzes (admin)", %{conn: conn, user: user, project: project} do
      conn = auth_conn(conn, user.id, ["admin"])

      resp =
        graphql_query(
          conn,
          """
            query($filter: QuizFilter, $first: Int) {
              quizzes(filter: $filter, first: $first) {
                totalCount
                edges { node { id name } }
              }
            }
          """,
          %{"filter" => %{"projectId" => project.id}, "first" => 10}
        )

      data = json_response(resp, 200)["data"]["quizzes"]
      assert data["totalCount"] >= 1
    end
  end

  # ── Quiz CRUD Mutations ──

  describe "createQuiz mutation" do
    test "creates a quiz as admin", %{conn: conn, user: user, project: project} do
      conn = auth_conn(conn, user.id, ["admin"])

      resp =
        graphql_query(
          conn,
          """
            mutation($input: CreateQuizInput!) {
              createQuiz(input: $input) {
                id
                name
                completionPoints
              }
            }
          """,
          %{
            "input" => %{
              "name" => "GQL Quiz",
              "description" => "test",
              "projectId" => project.id,
              "challengeId" => "CL00000000000000000000000000",
              "randomizeQuestions" => false,
              "revealCorrectAnswers" => true,
              "allowRetakes" => false,
              "completionPoints" => 20
            }
          }
        )

      data = json_response(resp, 200)["data"]["createQuiz"]
      assert data["name"] == "GQL Quiz"
      assert data["completionPoints"] == 20
    end
  end

  describe "updateQuiz mutation" do
    test "updates quiz fields", %{conn: conn, user: user, quiz: quiz} do
      conn = auth_conn(conn, user.id, ["admin"])

      resp =
        graphql_query(
          conn,
          """
            mutation($id: ID!, $input: UpdateQuizInput!) {
              updateQuiz(id: $id, input: $input) {
                id
                name
              }
            }
          """,
          %{"id" => quiz.id, "input" => %{"name" => "Updated"}}
        )

      data = json_response(resp, 200)["data"]["updateQuiz"]
      assert data["name"] == "Updated"
    end
  end

  describe "deleteQuiz mutation" do
    test "deletes a quiz", %{conn: conn, user: user, quiz: quiz} do
      conn = auth_conn(conn, user.id, ["admin"])

      resp =
        graphql_query(
          conn,
          """
            mutation($id: ID!) {
              deleteQuiz(id: $id)
            }
          """,
          %{"id" => quiz.id}
        )

      data = json_response(resp, 200)["data"]["deleteQuiz"]
      assert data == true
    end
  end

  # ── Question Mutations ──

  describe "addQuizQuestion mutation" do
    test "adds a predefined question", %{conn: conn, user: user, quiz: quiz} do
      conn = auth_conn(conn, user.id, ["admin"])

      resp =
        graphql_query(
          conn,
          """
            mutation($quizId: ID!, $input: CreateQuizQuestionInput!) {
              addQuizQuestion(quizId: $quizId, input: $input) {
                ... on PredefinedQuestion {
                  id
                  questionText
                  predefinedAnswers { answerText isCorrect }
                }
              }
            }
          """,
          %{
            "quizId" => quiz.id,
            "input" => %{
              "questionType" => "PREDEFINED",
              "questionText" => "What color?",
              "questionOrder" => 1,
              "predefinedAnswers" => [
                %{"answerText" => "Red", "isCorrect" => true, "answerOrder" => 1},
                %{"answerText" => "Blue", "isCorrect" => false, "answerOrder" => 2}
              ]
            }
          }
        )

      data = json_response(resp, 200)["data"]["addQuizQuestion"]
      assert data["questionText"] == "What color?"
      assert length(data["predefinedAnswers"]) == 2
    end
  end

  # ── Submission Flow ──

  describe "submission flow" do
    test "submit answers and finalize", %{conn: conn, user: user, quiz: quiz} do
      # Add a question
      alias ElixirBackend.Quizzes

      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "What is 1+1?",
          question_order: 1
        })

      # Create submission
      {:ok, sub} = Quizzes.create_submission(quiz.id, user.id)
      conn = auth_conn(conn, user.id, ["user"])

      # Submit answer via GraphQL
      resp =
        graphql_query(
          conn,
          """
            mutation($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
              submitQuizAnswer(submissionId: $submissionId, input: $input) {
                ... on FreeTextResponse {
                  id
                  textResponse
                }
              }
            }
          """,
          %{
            "submissionId" => sub.id,
            "input" => %{
              "questionId" => q.id,
              "textResponse" => "2"
            }
          }
        )

      data = json_response(resp, 200)["data"]["submitQuizAnswer"]
      assert data["textResponse"] == "2"

      # Finalize via GraphQL
      resp =
        graphql_query(
          conn,
          """
            mutation($submissionId: ID!) {
              finalizeQuiz(submissionId: $submissionId) {
                id
                completedAt
                score
                maxScore
                pointsAwarded
              }
            }
          """,
          %{"submissionId" => sub.id}
        )

      data = json_response(resp, 200)["data"]["finalizeQuiz"]
      assert data["completedAt"] != nil
      assert data["pointsAwarded"] == quiz.completion_points
    end
  end

  describe "createQuizSubmission mutation (m2m)" do
    test "creates submission with responses in one call", %{conn: conn, user: user, quiz: quiz} do
      alias ElixirBackend.Quizzes

      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Q1",
          question_order: 1
        })

      conn = auth_conn(conn, user.id, ["m2m"])

      resp =
        graphql_query(
          conn,
          """
            mutation($quizId: ID!, $userId: ID!, $responses: [SubmitQuizAnswerInput!]!, $completedAt: DateTime) {
              createQuizSubmission(quizId: $quizId, userId: $userId, responses: $responses, completedAt: $completedAt) {
                id
                completedAt
              }
            }
          """,
          %{
            "quizId" => quiz.id,
            "userId" => user.id,
            "responses" => [%{"questionId" => q.id, "textResponse" => "answer"}],
            "completedAt" => DateTime.to_iso8601(DateTime.utc_now())
          }
        )

      data = json_response(resp, 200)["data"]["createQuizSubmission"]
      assert data["id"] != nil
      assert data["completedAt"] != nil
    end
  end

  # ── Quiz Submission Queries ──

  describe "quizSubmission query" do
    test "returns submission by id", %{conn: conn, user: user, quiz: quiz} do
      {:ok, sub} = ElixirBackend.Quizzes.create_submission(quiz.id, user.id)
      conn = auth_conn(conn, user.id, ["user"])

      resp =
        graphql_query(
          conn,
          """
            query($id: ID!) {
              quizSubmission(id: $id) {
                id
                startedAt
                responses { ... on FreeTextResponse { id } }
                scorePercentage
              }
            }
          """,
          %{"id" => sub.id}
        )

      data = json_response(resp, 200)["data"]["quizSubmission"]
      assert data["id"] == sub.id
    end
  end

  describe "quizSubmissions query" do
    test "returns paginated submissions for a quiz (admin)", %{conn: conn, user: user, quiz: quiz} do
      {:ok, _} = ElixirBackend.Quizzes.create_submission(quiz.id, user.id)
      conn = auth_conn(conn, user.id, ["admin"])

      resp =
        graphql_query(
          conn,
          """
            query($quizId: ID!, $first: Int) {
              quizSubmissions(quizId: $quizId, first: $first) {
                totalCount
                edges { node { id } }
              }
            }
          """,
          %{"quizId" => quiz.id, "first" => 10}
        )

      data = json_response(resp, 200)["data"]["quizSubmissions"]
      assert data["totalCount"] >= 1
    end
  end
end
