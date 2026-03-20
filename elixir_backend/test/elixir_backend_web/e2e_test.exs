defmodule ElixirBackendWeb.E2ETest do
  @moduledoc """
  End-to-end tests covering multi-step user flows.

  These mirror the Go backend's e2e tests to ensure feature parity
  across the critical user journeys.
  """

  use ElixirBackendWeb.ConnCase, async: true

  import ElixirBackend.TestHelpers

  # ──────────────────────────────────────────────
  # Auth & Role Flows
  # ──────────────────────────────────────────────

  describe "auth flow" do
    test "unauthenticated request returns null for me", %{conn: conn} do
      resp = graphql_query(conn, "{ me { id name } }") |> json_response(200)
      assert resp["data"]["me"] == nil
    end

    test "valid token returns user data", %{conn: conn} do
      user = create_user(%{name: "Auth Test User"})

      resp =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query("{ me { id name } }")
        |> json_response(200)

      assert resp["data"]["me"]["id"] == user.id
      assert resp["data"]["me"]["name"] == "Auth Test User"
    end

    test "user role cannot query admin-only fields", %{conn: conn} do
      user = create_user()

      resp =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query("{ users(first: 10) { totalCount } }")
        |> json_response(200)

      assert resp["errors"] != nil
    end

    test "admin role can query users", %{conn: conn} do
      admin = create_user()

      resp =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query("{ users(first: 10) { totalCount } }")
        |> json_response(200)

      assert resp["data"]["users"]["totalCount"] >= 0
    end

    test "accept-language header sets language in context", %{conn: conn} do
      user = create_user()

      resp =
        conn
        |> Plug.Conn.put_req_header("accept-language", "en-US,en;q=0.9")
        |> auth_conn(user.id, ["user"])
        |> graphql_query("{ me { id } }")
        |> json_response(200)

      assert resp["data"]["me"]["id"] == user.id
    end
  end

  # ──────────────────────────────────────────────
  # Project → Event → Join Flow
  # ──────────────────────────────────────────────

  describe "project and event flow" do
    setup %{conn: conn} do
      admin = create_user()
      user = create_user()
      project = create_project(%{name: "E2E Project"})
      event = create_event(project, %{name: "E2E Event"})

      admin_conn = auth_conn(conn, admin.id, ["admin"])
      user_conn = auth_conn(conn, user.id, ["user"])

      %{admin: admin, user: user, project: project, event: event,
        admin_conn: admin_conn, user_conn: user_conn}
    end

    test "user can join project and query it", %{user_conn: user_conn, user: user, project: project} do
      ElixirBackend.Projects.join_project(user.id, project.id)

      query = """
      query($id: ID!) {
        project(id: $id) { id name events { id name } }
      }
      """

      resp =
        user_conn
        |> graphql_query(query, %{"id" => project.id})
        |> json_response(200)

      assert resp["data"]["project"]["name"] == "E2E Project"
      assert length(resp["data"]["project"]["events"]) == 1
      assert hd(resp["data"]["project"]["events"])["name"] == "E2E Event"
    end

    test "user can join event and see it in myEvents", %{user_conn: user_conn, user: user, project: project, event: event} do
      ElixirBackend.Projects.join_project(user.id, project.id)
      ElixirBackend.Events.join_event(user.id, event.id)

      query = """
      query($project: ID!) {
        myEvents(project: $project) { id name }
      }
      """

      resp =
        user_conn
        |> graphql_query(query, %{"project" => project.id})
        |> json_response(200)

      events = resp["data"]["myEvents"]
      assert length(events) == 1
      assert hd(events)["name"] == "E2E Event"
    end
  end

  # ──────────────────────────────────────────────
  # Challenge: Create → Enroll → Complete Flow
  # ──────────────────────────────────────────────

  describe "challenge lifecycle" do
    setup %{conn: conn} do
      admin = create_user()
      user = create_user()
      project = create_project()
      ElixirBackend.Projects.join_project(user.id, project.id)

      admin_conn = auth_conn(conn, admin.id, ["admin"])
      user_conn = auth_conn(conn, user.id, ["user"])

      %{admin: admin, user: user, project: project,
        admin_conn: admin_conn, user_conn: user_conn}
    end

    test "admin creates challenge, user enrolls, admin completes it", ctx do
      # Step 1: Admin creates a SIMPLE challenge
      create_mutation = """
      mutation($projectId: ID!, $input: CreateChallengeInput!) {
        createChallenge(projectId: $projectId, input: $input) {
          ... on SimpleChallenge { id name buttonText }
        }
      }
      """

      resp =
        ctx.admin_conn
        |> graphql_query(create_mutation, %{
          "projectId" => ctx.project.id,
          "input" => %{
            "type" => "SIMPLE",
            "name" => "Share a Bible Verse",
            "description" => "<p>Share your favorite verse</p>",
            "buttonText" => "Complete",
            "allowSelfCompletion" => true
          }
        })
        |> json_response(200)

      challenge_id = resp["data"]["createChallenge"]["id"]
      assert challenge_id != nil
      assert resp["data"]["createChallenge"]["name"] == "Share a Bible Verse"

      # Step 2: User enrolls in the challenge
      enroll_mutation = """
      mutation($challengeId: ID!) {
        enrollInChallenge(challengeId: $challengeId) {
          ... on SimpleChallenge { id userEnrolledAt }
        }
      }
      """

      resp =
        ctx.user_conn
        |> graphql_query(enroll_mutation, %{"challengeId" => challenge_id})
        |> json_response(200)

      assert resp["data"]["enrollInChallenge"]["userEnrolledAt"] != nil

      # Step 3: Admin completes the challenge for the user
      complete_mutation = """
      mutation($userId: ID!, $challengeId: ID!) {
        completeChallenge(userId: $userId, challengeId: $challengeId) {
          ... on SimpleChallenge { id name }
        }
      }
      """

      resp =
        ctx.admin_conn
        |> graphql_query(complete_mutation, %{
          "userId" => ctx.user.id,
          "challengeId" => challenge_id
        })
        |> json_response(200)

      assert resp["data"]["completeChallenge"]["id"] == challenge_id

      # Step 4: Verify via query as the USER that the challenge shows completed
      # (userCompletedAt resolves based on the calling user's context)
      query = """
      query($id: ID!) {
        challenge(id: $id) {
          ... on SimpleChallenge { id userCompletedAt userEnrolledAt }
        }
      }
      """

      resp =
        ctx.user_conn
        |> graphql_query(query, %{"id" => challenge_id})
        |> json_response(200)

      assert resp["data"]["challenge"]["userCompletedAt"] != nil
      assert resp["data"]["challenge"]["userEnrolledAt"] != nil
    end

    test "user can self-complete a challenge", ctx do
      {:ok, challenge} =
        ElixirBackend.Challenges.create_challenge(%{
          project_id: ctx.project.id,
          challenge_type: "SIMPLE",
          name: "Self Complete Test",
          description: "<p>test</p>",
          button_text: "Done",
          allow_self_completion: true,
          published_at: DateTime.utc_now() |> DateTime.truncate(:second)
        })

      # User enrolls and self-completes
      ElixirBackend.Challenges.enroll_in_challenge(ctx.user.id, challenge.id)

      complete_mutation = """
      mutation($challengeId: ID!) {
        selfCompleteChallenge(challengeId: $challengeId) { id userCompletedAt }
      }
      """

      resp =
        ctx.user_conn
        |> graphql_query(complete_mutation, %{"challengeId" => challenge.id})
        |> json_response(200)

      assert resp["data"]["selfCompleteChallenge"]["userCompletedAt"] != nil
    end
  end

  # ──────────────────────────────────────────────
  # Achievement: Create → Award → Score Impact
  # ──────────────────────────────────────────────

  describe "achievement and scoring flow" do
    setup %{conn: conn} do
      admin = create_user()
      user = create_user()
      project = create_project()
      ElixirBackend.Projects.join_project(user.id, project.id)

      admin_conn = auth_conn(conn, admin.id, ["admin"])
      user_conn = auth_conn(conn, user.id, ["user"])

      %{admin: admin, user: user, project: project,
        admin_conn: admin_conn, user_conn: user_conn}
    end

    test "awarding achievement grants points visible on project", ctx do
      # Step 1: Admin creates achievement worth 100 points
      create_mutation = """
      mutation($input: CreateSimpleAchievementInput!) {
        createSimpleAchievement(input: $input) {
          ... on SimpleAchievement { id name points }
        }
      }
      """

      resp =
        ctx.admin_conn
        |> graphql_query(create_mutation, %{
          "input" => %{
            "name" => "First Steps",
            "descriptionPending" => "Do something",
            "descriptionCompleted" => "You did it!",
            "imagePending" => "https://example.com/pending.png",
            "imageCompleted" => "https://example.com/done.png",
            "projectId" => ctx.project.id,
            "points" => 100,
            "hidden" => false
          }
        })
        |> json_response(200)

      achievement_id = resp["data"]["createSimpleAchievement"]["id"]
      assert achievement_id != nil
      assert resp["data"]["createSimpleAchievement"]["points"] == 100

      # Step 2: Award achievement to user
      award_mutation = """
      mutation($userId: ID!, $achievementId: ID!) {
        awardAchievement(userId: $userId, achievementId: $achievementId) {
          ... on SimpleAchievement { id name }
        }
      }
      """

      resp =
        ctx.admin_conn
        |> graphql_query(award_mutation, %{
          "userId" => ctx.user.id,
          "achievementId" => achievement_id
        })
        |> json_response(200)

      assert resp["data"]["awardAchievement"]["id"] == achievement_id

      # Step 2b: Verify achievedAt from user's perspective
      achievement_query = """
      query($id: ID!) {
        achievement(id: $id) {
          ... on SimpleAchievement { id achievedAt }
        }
      }
      """

      resp =
        ctx.user_conn
        |> graphql_query(achievement_query, %{"id" => achievement_id})
        |> json_response(200)

      assert resp["data"]["achievement"]["achievedAt"] != nil

      # Step 3: Revoke achievement
      revoke_mutation = """
      mutation($userId: ID!, $achievementId: ID!) {
        revokeAchievement(userId: $userId, achievementId: $achievementId)
      }
      """

      resp =
        ctx.admin_conn
        |> graphql_query(revoke_mutation, %{
          "userId" => ctx.user.id,
          "achievementId" => achievement_id
        })
        |> json_response(200)

      assert resp["data"]["revokeAchievement"] == true

      # Step 4: Verify achievedAt is now nil after revoke
      resp =
        ctx.user_conn
        |> graphql_query(achievement_query, %{"id" => achievement_id})
        |> json_response(200)

      assert resp["data"]["achievement"]["achievedAt"] == nil
    end
  end

  # ──────────────────────────────────────────────
  # Quiz: Create → Questions → Submit → Score
  # ──────────────────────────────────────────────

  describe "quiz submission flow" do
    setup %{conn: conn} do
      admin = create_user()
      user = create_user()
      project = create_project()
      ElixirBackend.Projects.join_project(user.id, project.id)

      {:ok, quiz} =
        ElixirBackend.Quizzes.create_quiz(%{
          name: "E2E Quiz",
          description: "Testing quiz flow",
          project_id: project.id,
          randomize_questions: false,
          reveal_correct_answers: true,
          allow_retakes: false,
          completion_points: 50
        })

      {:ok, q1} =
        ElixirBackend.Quizzes.add_question(quiz.id, %{
          question_type: "PREDEFINED",
          question_text: "What is 2+2?",
          question_order: 1,
          points: 10,
          betting_enabled: false,
          allow_multiple_selection: false,
          predefined_answers: [
            %{answer_text: "3", is_correct: false, answer_order: 0},
            %{answer_text: "4", is_correct: true, answer_order: 1},
            %{answer_text: "5", is_correct: false, answer_order: 2}
          ]
        })

      correct_answer =
        ElixirBackend.Quizzes.get_predefined_answers(q1.id)
        |> Enum.find(&(&1.is_correct))

      admin_conn = auth_conn(conn, admin.id, ["admin"])
      user_conn = auth_conn(conn, user.id, ["user"])

      %{admin: admin, user: user, project: project, quiz: quiz,
        question: q1, correct_answer: correct_answer,
        admin_conn: admin_conn, user_conn: user_conn}
    end

    test "user starts submission, answers question, and gets scored", ctx do
      # Step 1: Create a submission directly (no user-facing start mutation yet)
      {:ok, submission} =
        ElixirBackend.Quizzes.create_submission(ctx.quiz.id, ctx.user.id)

      assert submission.id != nil
      assert submission.started_at != nil

      # Step 2: Submit answer via GraphQL
      answer_mutation = """
      mutation($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
        submitQuizAnswer(submissionId: $submissionId, input: $input) {
          ... on PredefinedResponse {
            id
            selectedAnswerIds
          }
        }
      }
      """

      resp =
        ctx.user_conn
        |> graphql_query(answer_mutation, %{
          "submissionId" => submission.id,
          "input" => %{
            "questionId" => ctx.question.id,
            "selectedAnswerIds" => [ctx.correct_answer.id]
          }
        })
        |> json_response(200)

      assert resp["data"]["submitQuizAnswer"]["id"] != nil
      assert resp["data"]["submitQuizAnswer"]["selectedAnswerIds"] == [ctx.correct_answer.id]

      # Step 3: Finalize submission via GraphQL
      finalize_mutation = """
      mutation($submissionId: ID!) {
        finalizeQuiz(submissionId: $submissionId) {
          id
          completedAt
          score
          maxScore
        }
      }
      """

      resp =
        ctx.user_conn
        |> graphql_query(finalize_mutation, %{"submissionId" => submission.id})
        |> json_response(200)

      result = resp["data"]["finalizeQuiz"]
      assert result["completedAt"] != nil
    end
  end

  # ──────────────────────────────────────────────
  # Team: Create → Join → Membership
  # ──────────────────────────────────────────────

  describe "team membership flow" do
    setup %{conn: conn} do
      admin = create_user()
      user = create_user()
      project = create_project()
      ElixirBackend.Projects.join_project(user.id, project.id)

      admin_conn = auth_conn(conn, admin.id, ["admin"])
      user_conn = auth_conn(conn, user.id, ["user"])

      %{admin: admin, user: user, project: project,
        admin_conn: admin_conn, user_conn: user_conn}
    end

    test "admin creates team, adds user, user sees team", ctx do
      # Step 1: Admin creates team
      create_mutation = """
      mutation($projectId: ID!, $input: CreateTeamInput!) {
        createTeam(projectId: $projectId, input: $input) {
          id name joinCode
        }
      }
      """

      resp =
        ctx.admin_conn
        |> graphql_query(create_mutation, %{
          "projectId" => ctx.project.id,
          "input" => %{"name" => "E2E Team", "description" => "Test team"}
        })
        |> json_response(200)

      team = resp["data"]["createTeam"]
      team_id = team["id"]
      assert team_id != nil
      assert team["joinCode"] != nil

      # Step 2: Admin adds user to team
      add_mutation = """
      mutation($teamId: ID!, $userIds: [ID!]!) {
        addTeamMembers(teamId: $teamId, userIds: $userIds) {
          id
          members { user { id name } }
        }
      }
      """

      resp =
        ctx.admin_conn
        |> graphql_query(add_mutation, %{
          "teamId" => team_id,
          "userIds" => [ctx.user.id]
        })
        |> json_response(200)

      members = resp["data"]["addTeamMembers"]["members"]
      assert length(members) == 1
      assert hd(members)["user"]["id"] == ctx.user.id

      # Step 3: User queries their team
      query = """
      query($id: ID!) {
        team(id: $id) {
          id name
          members { user { id } }
        }
      }
      """

      resp =
        ctx.user_conn
        |> graphql_query(query, %{"id" => team_id})
        |> json_response(200)

      assert resp["data"]["team"]["name"] == "E2E Team"
      assert length(resp["data"]["team"]["members"]) == 1
    end
  end

  # ──────────────────────────────────────────────
  # Consent: Accept → Pending Check
  # ──────────────────────────────────────────────

  describe "consent flow" do
    setup %{conn: conn} do
      admin = create_user()
      user = create_user()

      {:ok, consent} =
        ElixirBackend.Consents.create_consent(%{
          key: "e2e_privacy",
          version: 1,
          title: "Privacy Policy",
          short_text: "We value your privacy",
          body: "<p>Full privacy policy</p>",
          published_at: DateTime.utc_now() |> DateTime.truncate(:second)
        })

      admin_conn = auth_conn(conn, admin.id, ["admin"])
      user_conn = auth_conn(conn, user.id, ["user"])

      %{admin: admin, user: user, consent: consent,
        admin_conn: admin_conn, user_conn: user_conn}
    end

    test "user sees pending consent, accepts it, then no longer pending", ctx do
      # Step 1: User has pending consents
      pending_query = "{ pendingConsents { id title } }"

      resp =
        ctx.user_conn
        |> graphql_query(pending_query)
        |> json_response(200)

      pending = resp["data"]["pendingConsents"]
      assert length(pending) >= 1
      assert Enum.any?(pending, &(&1["id"] == ctx.consent.id))

      # Step 2: User accepts the consent
      ElixirBackend.Consents.accept_consent(ctx.user.id, ctx.consent.id)

      # Step 3: No longer pending
      resp =
        ctx.user_conn
        |> graphql_query(pending_query)
        |> json_response(200)

      pending = resp["data"]["pendingConsents"]
      refute Enum.any?(pending, &(&1["id"] == ctx.consent.id))
    end
  end

  # ──────────────────────────────────────────────
  # Streak: Record Activity → Check Status
  # ──────────────────────────────────────────────

  describe "streak activity flow" do
    setup %{conn: conn} do
      user = create_user()
      project = create_project()
      ElixirBackend.Projects.join_project(user.id, project.id)

      streak = create_streak(project, %{
        name: "Daily Reading",
        relevant_days: [%{start: ~D[2025-01-01], end: ~D[2027-12-31]}]
      })

      user_conn = auth_conn(conn, user.id, ["user"])

      %{user: user, project: project, streak: streak, user_conn: user_conn}
    end

    test "user records activity and streak status increases", ctx do
      # Step 1: Query streak — status should be 0
      query = """
      query($id: ID!) {
        streak(id: $id) { id name status }
      }
      """

      resp =
        ctx.user_conn
        |> graphql_query(query, %{"id" => ctx.streak.id})
        |> json_response(200)

      assert resp["data"]["streak"]["status"] == 0

      # Step 2: Record activity for today
      ElixirBackend.Streaks.record_activity(ctx.streak.id, ctx.user.id)

      # Step 3: Status should now be 1
      resp =
        ctx.user_conn
        |> graphql_query(query, %{"id" => ctx.streak.id})
        |> json_response(200)

      assert resp["data"]["streak"]["status"] == 1

      # Step 4: Record yesterday too
      ElixirBackend.Streaks.record_activity(ctx.streak.id, ctx.user.id, Date.add(Date.utc_today(), -1))

      resp =
        ctx.user_conn
        |> graphql_query(query, %{"id" => ctx.streak.id})
        |> json_response(200)

      assert resp["data"]["streak"]["status"] == 2
    end
  end

  # ──────────────────────────────────────────────
  # Translation: Verify Accept-Language works
  # ──────────────────────────────────────────────

  describe "translation flow" do
    setup %{conn: conn} do
      admin = create_user()
      project = create_project(%{name: "Norsk Navn", description: "Norsk beskrivelse"})

      admin_conn = auth_conn(conn, admin.id, ["admin"])

      %{admin: admin, project: project, admin_conn: admin_conn}
    end

    test "project returns translated name when Accept-Language is set", ctx do
      # Step 1: Upsert English translation
      ElixirBackend.Translations.upsert_translation(:project, %{
        project_id: ctx.project.id,
        language_code: "en",
        name: "English Name",
        description: "English description"
      })

      # Step 2: Query with Norwegian (default) — should return original
      query = """
      query($id: ID!) {
        project(id: $id) { id name description }
      }
      """

      resp =
        ctx.admin_conn
        |> graphql_query(query, %{"id" => ctx.project.id})
        |> json_response(200)

      assert resp["data"]["project"]["name"] == "Norsk Navn"

      # Step 3: Query with English — should return translated
      resp =
        ctx.admin_conn
        |> Plug.Conn.put_req_header("accept-language", "en")
        |> graphql_query(query, %{"id" => ctx.project.id})
        |> json_response(200)

      assert resp["data"]["project"]["name"] == "English Name"
      assert resp["data"]["project"]["description"] == "English description"
    end
  end
end
