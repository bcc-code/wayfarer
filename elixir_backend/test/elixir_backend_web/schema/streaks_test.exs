defmodule ElixirBackendWeb.Schema.StreaksTest do
  use ElixirBackendWeb.ConnCase

  alias ElixirBackend.Streaks

  defp create_streak_fixture(project, attrs \\ %{}) do
    default_attrs = %{
      name: "Daily Streak",
      description: "Read every day",
      project_id: project.id,
      relevant_days: [
        %{start: ~D[2026-03-01], end: ~D[2026-03-31]}
      ]
    }

    {:ok, streak} = Streaks.create_streak(Map.merge(default_attrs, attrs))
    streak
  end

  # ── Query: streak ──

  describe "streak query" do
    test "returns a streak by id", %{conn: conn} do
      project = create_project()
      streak = create_streak_fixture(project)

      query = """
      query($id: ID!) {
        streak(id: $id) {
          id name description
          relevantDays { start end }
        }
      }
      """

      conn = graphql_query(conn, query, %{"id" => streak.id})
      data = json_response(conn, 200)["data"]["streak"]

      assert data["id"] == streak.id
      assert data["name"] == "Daily Streak"
      assert length(data["relevantDays"]) == 1
      assert data["relevantDays"] |> hd() |> Map.get("start") == "2026-03-01"
    end

    test "returns error for non-existent streak", %{conn: conn} do
      query = """
      query($id: ID!) {
        streak(id: $id) { id }
      }
      """

      conn = graphql_query(conn, query, %{"id" => "SK00000000000000000000000000"})
      json = json_response(conn, 200)
      assert json["errors"] != nil
    end
  end

  # ── Query: streaks ──

  describe "streaks query" do
    test "returns paginated streaks", %{conn: conn} do
      project = create_project()
      create_streak_fixture(project, %{name: "S1"})
      create_streak_fixture(project, %{name: "S2"})

      query = """
      query($filter: StreakFilter) {
        streaks(filter: $filter, first: 10) {
          totalCount
          edges { node { id name } }
          pageInfo { hasNextPage }
        }
      }
      """

      conn = graphql_query(conn, query, %{"filter" => %{"projectId" => project.id}})
      data = json_response(conn, 200)["data"]["streaks"]

      assert data["totalCount"] == 2
      assert length(data["edges"]) == 2
    end
  end

  # ── Mutation: createStreak ──

  describe "createStreak mutation" do
    test "admin can create a streak", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      project = create_project()

      query = """
      mutation($input: CreateStreakInput!) {
        createStreak(input: $input) {
          id name description
          relevantDays { start end }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "input" => %{
            "name" => "New Streak",
            "description" => "Test",
            "projectId" => project.id,
            "relevantDays" => [
              %{"start" => "2026-03-01", "end" => "2026-03-15"},
              %{"start" => "2026-03-20", "end" => "2026-03-31"}
            ]
          }
        })

      data = json_response(conn, 200)["data"]["createStreak"]
      assert data["name"] == "New Streak"
      assert String.starts_with?(data["id"], "SK")
      assert length(data["relevantDays"]) == 2
    end

    test "user cannot create a streak", %{conn: conn} do
      user = create_user()
      project = create_project()

      query = """
      mutation($input: CreateStreakInput!) {
        createStreak(input: $input) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{
          "input" => %{
            "name" => "Streak",
            "description" => "",
            "projectId" => project.id,
            "relevantDays" => []
          }
        })

      json = json_response(conn, 200)
      assert json["errors"] != nil
      assert hd(json["errors"])["message"] =~ "unauthorized"
    end
  end

  # ── Mutation: updateStreak ──

  describe "updateStreak mutation" do
    test "admin can update a streak", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      project = create_project()
      streak = create_streak_fixture(project)

      query = """
      mutation($id: ID!, $input: UpdateStreakInput!) {
        updateStreak(id: $id, input: $input) {
          id name
          relevantDays { start end }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "id" => streak.id,
          "input" => %{
            "name" => "Updated Streak",
            "relevantDays" => [%{"start" => "2026-04-01", "end" => "2026-04-30"}]
          }
        })

      data = json_response(conn, 200)["data"]["updateStreak"]
      assert data["name"] == "Updated Streak"
      assert length(data["relevantDays"]) == 1
    end
  end

  # ── Mutation: deleteStreak ──

  describe "deleteStreak mutation" do
    test "admin can delete a streak", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      project = create_project()
      streak = create_streak_fixture(project)

      query = """
      mutation($id: ID!) {
        deleteStreak(id: $id)
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{"id" => streak.id})

      assert json_response(conn, 200)["data"]["deleteStreak"] == true
    end
  end

  # ── Resolved fields: status, listenedDays ──

  describe "streak status and listenedDays" do
    test "returns status and listened days for authenticated user", %{conn: conn} do
      user = create_user()
      project = create_project()

      streak =
        create_streak_fixture(project, %{
          relevant_days: [%{start: ~D[2026-03-01], end: ~D[2026-03-05]}]
        })

      Streaks.record_activity(streak.id, user.id, ~D[2026-03-01])
      Streaks.record_activity(streak.id, user.id, ~D[2026-03-03])

      query = """
      query($id: ID!) {
        streak(id: $id) {
          status
          listenedDays(last: 10) { date active }
        }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"id" => streak.id})

      data = json_response(conn, 200)["data"]["streak"]
      assert is_integer(data["status"])

      listened = data["listenedDays"]
      assert length(listened) == 5
      active_dates = listened |> Enum.filter(& &1["active"]) |> Enum.map(& &1["date"])
      assert "2026-03-01" in active_dates
      assert "2026-03-03" in active_dates
    end
  end
end
