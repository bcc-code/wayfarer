defmodule ElixirBackendWeb.Schema.ChallengesTest do
  use ElixirBackendWeb.ConnCase

  alias ElixirBackend.Challenges

  # ── Query: challenge(id) ──

  describe "challenge query" do
    test "returns a SIMPLE challenge by id", %{conn: conn} do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)

      query = """
      query($id: ID!) {
        challenge(id: $id) {
          id
          name
          description
          __typename
          ... on SimpleChallenge {
            allowSelfCompletion
            buttonText
          }
        }
      }
      """

      conn = graphql_query(conn, query, %{"id" => challenge.id})
      data = json_response(conn, 200)["data"]["challenge"]

      assert data["id"] == challenge.id
      assert data["name"] == "Simple Challenge"
      assert data["__typename"] == "SimpleChallenge"
      assert data["allowSelfCompletion"] == true
      assert data["buttonText"] == "Complete"
    end

    test "returns correct __typename for EXTERNAL challenge", %{conn: conn} do
      project = create_project()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "EXTERNAL",
          name: "External",
          button_text: "Go",
          url: "https://example.com"
        })

      query = """
      query($id: ID!) {
        challenge(id: $id) {
          __typename
          ... on ExternalChallenge {
            url
          }
        }
      }
      """

      conn = graphql_query(conn, query, %{"id" => challenge.id})
      data = json_response(conn, 200)["data"]["challenge"]

      assert data["__typename"] == "ExternalChallenge"
      assert data["url"] == "https://example.com"
    end

    test "returns correct __typename for PLUGIN challenge", %{conn: conn} do
      project = create_project()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "PLUGIN",
          name: "Plugin",
          plugin_challenge_id: "plugin-abc"
        })

      query = """
      query($id: ID!) {
        challenge(id: $id) {
          __typename
          ... on PluginChallenge {
            pluginChallengeId
          }
        }
      }
      """

      conn = graphql_query(conn, query, %{"id" => challenge.id})
      data = json_response(conn, 200)["data"]["challenge"]

      assert data["__typename"] == "PluginChallenge"
      assert data["pluginChallengeId"] == "plugin-abc"
    end

    test "returns correct __typename for QUIZ challenge", %{conn: conn} do
      project = create_project()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "QUIZ",
          name: "Quiz",
          button_text: "Start"
        })

      query = """
      query($id: ID!) {
        challenge(id: $id) {
          __typename
        }
      }
      """

      conn = graphql_query(conn, query, %{"id" => challenge.id})
      data = json_response(conn, 200)["data"]["challenge"]

      assert data["__typename"] == "QuizChallenge"
    end

    test "returns error for nonexistent challenge", %{conn: conn} do
      query = """
      query($id: ID!) {
        challenge(id: $id) {
          id
        }
      }
      """

      conn = graphql_query(conn, query, %{"id" => "CL00000000000000000000000000"})
      json = json_response(conn, 200)

      assert json["errors"] != nil
    end
  end

  # ── Query: challenges(filter, pagination) ──

  describe "challenges query" do
    test "returns paginated challenges", %{conn: conn} do
      project = create_project()

      for i <- 1..5 do
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Challenge #{i}",
          button_text: "Do it"
        })
      end

      query = """
      query($filter: ChallengeFilter, $first: Int) {
        challenges(filter: $filter, first: $first) {
          edges {
            cursor
            node { id name }
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
      """

      conn =
        conn
        |> auth_conn(create_user().id, ["admin"])
        |> graphql_query(query, %{
          "filter" => %{"projectId" => project.id},
          "first" => 3
        })

      data = json_response(conn, 200)["data"]["challenges"]
      assert length(data["edges"]) == 3
      assert data["totalCount"] == 5
      assert data["pageInfo"]["hasNextPage"] == true
      assert data["pageInfo"]["startCursor"] != nil
      assert data["pageInfo"]["endCursor"] != nil
    end

    test "filters by projectId and eventId", %{conn: conn} do
      project = create_project()
      event = create_event(project)

      Challenges.create_challenge(%{
        project_id: project.id,
        event_id: event.id,
        challenge_type: "SIMPLE",
        name: "Event Challenge",
        button_text: "Do"
      })

      create_simple_challenge(project, %{name: "No Event"})

      query = """
      query($filter: ChallengeFilter, $first: Int) {
        challenges(filter: $filter, first: $first) {
          edges { node { id name } }
          totalCount
        }
      }
      """

      conn =
        conn
        |> auth_conn(create_user().id, ["admin"])
        |> graphql_query(query, %{
          "filter" => %{"projectId" => project.id, "eventId" => event.id},
          "first" => 10
        })

      data = json_response(conn, 200)["data"]["challenges"]
      assert data["totalCount"] == 1
      assert hd(data["edges"])["node"]["name"] == "Event Challenge"
    end
  end

  # ── Visibility via GraphQL ──

  describe "challenge visibility via GraphQL" do
    test "unenrolled user cannot see challenge with future visible_at", %{conn: conn} do
      project = create_project()
      user = create_user()
      future = DateTime.add(DateTime.utc_now(), 3600)

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Hidden",
          button_text: "Do",
          visible_at: future
        })

      query = "query($id: ID!) { challenge(id: $id) { id } }"

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"id" => challenge.id})

      json = json_response(conn, 200)
      assert json["errors"] != nil
    end

    test "enrolled user CAN see challenge with future visible_at", %{conn: conn} do
      project = create_project()
      user = create_user()
      future = DateTime.add(DateTime.utc_now(), 3600)

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Hidden",
          button_text: "Do",
          visible_at: future
        })

      Challenges.enroll_in_challenge(user.id, challenge.id)

      query = "query($id: ID!) { challenge(id: $id) { id name } }"

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"id" => challenge.id})

      data = json_response(conn, 200)["data"]["challenge"]
      assert data["id"] == challenge.id
    end

    test "admin can see challenge with future visible_at", %{conn: conn} do
      project = create_project()
      admin = create_user(%{name: "Admin"})
      future = DateTime.add(DateTime.utc_now(), 3600)

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Hidden",
          button_text: "Do",
          visible_at: future,
          published_at: future
        })

      query = "query($id: ID!) { challenge(id: $id) { id name } }"

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{"id" => challenge.id})

      data = json_response(conn, 200)["data"]["challenge"]
      assert data["id"] == challenge.id
    end
  end

  # ── Mutation: createChallenge ──

  describe "createChallenge mutation" do
    test "admin can create SIMPLE challenge", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      mutation($projectId: ID!, $input: CreateChallengeInput!) {
        createChallenge(projectId: $projectId, input: $input) {
          id name __typename
          ... on SimpleChallenge { allowSelfCompletion buttonText }
        }
      }
      """

      variables = %{
        "projectId" => project.id,
        "input" => %{
          "type" => "SIMPLE",
          "name" => "New Challenge",
          "description" => "<p>Desc</p>",
          "buttonText" => "Complete"
        }
      }

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, variables)

      data = json_response(conn, 200)["data"]["createChallenge"]
      assert data["name"] == "New Challenge"
      assert data["__typename"] == "SimpleChallenge"
      assert data["allowSelfCompletion"] == true
      assert data["buttonText"] == "Complete"
      assert String.starts_with?(data["id"], "CL")
    end

    test "admin can create EXTERNAL challenge", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      mutation($projectId: ID!, $input: CreateChallengeInput!) {
        createChallenge(projectId: $projectId, input: $input) {
          __typename
          ... on ExternalChallenge { url }
        }
      }
      """

      variables = %{
        "projectId" => project.id,
        "input" => %{
          "type" => "EXTERNAL",
          "name" => "External",
          "buttonText" => "Open",
          "url" => "https://example.com"
        }
      }

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, variables)

      data = json_response(conn, 200)["data"]["createChallenge"]
      assert data["__typename"] == "ExternalChallenge"
      assert data["url"] == "https://example.com"
    end

    test "admin can create challenge without event", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      mutation($projectId: ID!, $input: CreateChallengeInput!) {
        createChallenge(projectId: $projectId, input: $input) { id }
      }
      """

      variables = %{
        "projectId" => project.id,
        "input" => %{
          "type" => "SIMPLE",
          "name" => "No Event",
          "buttonText" => "Go"
        }
      }

      conn =
        conn
        |> auth_conn(user.id, ["superadmin"])
        |> graphql_query(query, variables)

      data = json_response(conn, 200)["data"]["createChallenge"]
      assert data["id"] != nil
    end

    test "user cannot create challenge", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      mutation($projectId: ID!, $input: CreateChallengeInput!) {
        createChallenge(projectId: $projectId, input: $input) { id }
      }
      """

      variables = %{
        "projectId" => project.id,
        "input" => %{
          "type" => "SIMPLE",
          "name" => "New Challenge",
          "buttonText" => "Complete"
        }
      }

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, variables)

      json = json_response(conn, 200)
      assert [%{"message" => "unauthorized" <> _}] = json["errors"]
    end

    test "unauthenticated request rejected", %{conn: conn} do
      project = create_project()

      query = """
      mutation($projectId: ID!, $input: CreateChallengeInput!) {
        createChallenge(projectId: $projectId, input: $input) { id }
      }
      """

      variables = %{
        "projectId" => project.id,
        "input" => %{"type" => "SIMPLE", "name" => "X", "buttonText" => "Y"}
      }

      conn = graphql_query(conn, query, variables)
      json = json_response(conn, 200)
      assert [%{"message" => "unauthorized" <> _}] = json["errors"]
    end
  end

  # ── Mutation: updateChallenge ──

  describe "updateChallenge mutation" do
    test "admin can update challenge", %{conn: conn} do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      query = """
      mutation($id: ID!, $input: UpdateChallengeInput!) {
        updateChallenge(id: $id, input: $input) { id name }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{
          "id" => challenge.id,
          "input" => %{"name" => "Updated Name"}
        })

      data = json_response(conn, 200)["data"]["updateChallenge"]
      assert data["name"] == "Updated Name"
    end
  end

  # ── Mutation: deleteChallenge ──

  describe "deleteChallenge mutation" do
    test "admin can delete challenge", %{conn: conn} do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      query = "mutation($id: ID!) { deleteChallenge(id: $id) }"

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"id" => challenge.id})

      assert json_response(conn, 200)["data"]["deleteChallenge"] == true
    end
  end

  # ── Mutation: publishChallenge ──

  describe "publishChallenge mutation" do
    test "admin can publish challenge", %{conn: conn} do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      query = """
      mutation($id: ID!, $publishedAt: DateTime!) {
        publishChallenge(id: $id, publishedAt: $publishedAt) { id publishedAt }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{
          "id" => challenge.id,
          "publishedAt" => "2026-06-15T12:00:00Z"
        })

      data = json_response(conn, 200)["data"]["publishChallenge"]
      assert data["publishedAt"] == "2026-06-15T12:00:00Z"
    end
  end

  # ── Mutation: setChallengeVisibility ──

  describe "setChallengeVisibility mutation" do
    test "admin can set challenge visibility", %{conn: conn} do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      query = """
      mutation($id: ID!, $visibleAt: DateTime!, $startedAt: DateTime) {
        setChallengeVisibility(id: $id, visibleAt: $visibleAt, startedAt: $startedAt) {
          id visibleAt startedAt
        }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{
          "id" => challenge.id,
          "visibleAt" => "2026-06-01T00:00:00Z",
          "startedAt" => "2026-06-01T10:00:00Z"
        })

      data = json_response(conn, 200)["data"]["setChallengeVisibility"]
      assert data["visibleAt"] == "2026-06-01T00:00:00Z"
      assert data["startedAt"] == "2026-06-01T10:00:00Z"
    end
  end

  # ── Mutation: assignChallengeToEvent ──

  describe "assignChallengeToEvent mutation" do
    test "admin can assign challenge to event", %{conn: conn} do
      project = create_project()
      event = create_event(project)
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      query = """
      mutation($challengeId: ID!, $eventId: ID!) {
        assignChallengeToEvent(challengeId: $challengeId, eventId: $eventId) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{
          "challengeId" => challenge.id,
          "eventId" => event.id
        })

      data = json_response(conn, 200)["data"]["assignChallengeToEvent"]
      assert data["id"] == challenge.id

      # Verify assignment persisted
      {:ok, updated} = Challenges.get_challenge(challenge.id)
      assert updated.event_id == event.id
    end
  end

  # ── Mutation: enrollInChallenge ──

  describe "enrollInChallenge mutation" do
    test "user can enroll in challenge", %{conn: conn} do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      query = """
      mutation($challengeId: ID!) {
        enrollInChallenge(challengeId: $challengeId) { id name }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"challengeId" => challenge.id})

      data = json_response(conn, 200)["data"]["enrollInChallenge"]
      assert data["id"] == challenge.id

      # Verify enrollment persisted
      assert {:ok, %DateTime{}} = Challenges.get_user_enrolled_at(user.id, challenge.id)
    end

    test "unauthenticated enrollment rejected", %{conn: conn} do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)

      query = """
      mutation($challengeId: ID!) {
        enrollInChallenge(challengeId: $challengeId) { id }
      }
      """

      conn = graphql_query(conn, query, %{"challengeId" => challenge.id})
      json = json_response(conn, 200)
      assert [%{"message" => "authentication required"}] = json["errors"]
    end
  end

  # ── Mutation: unenrollFromChallenge ──

  describe "unenrollFromChallenge mutation" do
    test "user can unenroll from challenge", %{conn: conn} do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)
      Challenges.enroll_in_challenge(user.id, challenge.id)

      query = """
      mutation($challengeId: ID!) {
        unenrollFromChallenge(challengeId: $challengeId)
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"challengeId" => challenge.id})

      assert json_response(conn, 200)["data"]["unenrollFromChallenge"] == true
      assert {:ok, nil} = Challenges.get_user_enrolled_at(user.id, challenge.id)
    end
  end

  # ── Mutation: completeChallenge (admin) ──

  describe "completeChallenge mutation" do
    test "admin can complete challenge for user", %{conn: conn} do
      project = create_project()
      admin = create_user(%{name: "Admin"})
      user = create_user(%{name: "User"})
      {:ok, challenge} = create_simple_challenge(project)

      query = """
      mutation($userId: ID!, $challengeId: ID!) {
        completeChallenge(userId: $userId, challengeId: $challengeId) { id }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "userId" => user.id,
          "challengeId" => challenge.id
        })

      data = json_response(conn, 200)["data"]["completeChallenge"]
      assert data["id"] == challenge.id

      # Verify completion persisted
      assert {:ok, %DateTime{}} = Challenges.get_user_completed_at(user.id, challenge.id)
    end

    test "user cannot complete challenge for others", %{conn: conn} do
      project = create_project()
      user = create_user()
      other = create_user(%{name: "Other"})
      {:ok, challenge} = create_simple_challenge(project)

      query = """
      mutation($userId: ID!, $challengeId: ID!) {
        completeChallenge(userId: $userId, challengeId: $challengeId) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{
          "userId" => other.id,
          "challengeId" => challenge.id
        })

      json = json_response(conn, 200)
      assert [%{"message" => "unauthorized" <> _}] = json["errors"]
    end
  end

  # ── Mutation: selfCompleteChallenge ──

  describe "selfCompleteChallenge mutation" do
    test "user can self-complete simple challenge", %{conn: conn} do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      query = """
      mutation($challengeId: ID!) {
        selfCompleteChallenge(challengeId: $challengeId) {
          id allowSelfCompletion
        }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"challengeId" => challenge.id})

      data = json_response(conn, 200)["data"]["selfCompleteChallenge"]
      assert data["id"] == challenge.id
      assert data["allowSelfCompletion"] == true
    end

    test "user cannot self-complete when not allowed", %{conn: conn} do
      project = create_project()
      user = create_user()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "No Self",
          button_text: "Do",
          allow_self_completion: false
        })

      query = """
      mutation($challengeId: ID!) {
        selfCompleteChallenge(challengeId: $challengeId) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"challengeId" => challenge.id})

      json = json_response(conn, 200)
      assert json["errors"] != nil
      assert hd(json["errors"])["message"] =~ "self-completion not allowed"
    end
  end

  # ── Helpers ──

  defp create_simple_challenge(project, extra_attrs \\ %{}) do
    attrs =
      Map.merge(
        %{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Simple Challenge",
          button_text: "Complete"
        },
        extra_attrs
      )

    Challenges.create_challenge(attrs)
  end
end
