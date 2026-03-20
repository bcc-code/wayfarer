defmodule ElixirBackendWeb.Schema.ScoringTest do
  use ElixirBackendWeb.ConnCase, async: true

  import ElixirBackend.TestHelpers

  setup do
    project = create_project()
    user = create_user()
    %{project: project, user: user}
  end

  describe "scoreJournal query" do
    test "returns user score journal", %{conn: conn, project: project, user: user} do
      {:ok, _} =
        ElixirBackend.Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 100,
          reason: "Test"
        })

      conn = auth_conn(conn, user.id, ["user"])

      resp =
        graphql_query(
          conn,
          """
            query($projectId: ID!, $userId: ID!, $first: Int) {
              scoreJournal(projectId: $projectId, userId: $userId, first: $first) {
                totalCount
                edges {
                  node {
                    id
                    points
                    sourceType
                    reason
                    createdAt
                  }
                }
              }
            }
          """,
          %{"projectId" => project.id, "userId" => user.id, "first" => 10}
        )

      data = json_response(resp, 200)["data"]["scoreJournal"]
      assert data["totalCount"] >= 1
      node = hd(data["edges"])["node"]
      assert node["points"] == 100
      assert node["sourceType"] == "MANUAL"
    end
  end

  describe "adminScoreJournal query" do
    test "admin can query all entries", %{conn: conn, project: project, user: user} do
      {:ok, _} =
        ElixirBackend.Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 50
        })

      conn = auth_conn(conn, user.id, ["admin"])

      resp =
        graphql_query(
          conn,
          """
            query($filter: ScoreJournalFilter, $first: Int) {
              adminScoreJournal(filter: $filter, first: $first) {
                totalCount
                edges { node { id points } }
              }
            }
          """,
          %{"filter" => %{"projectId" => project.id}, "first" => 10}
        )

      data = json_response(resp, 200)["data"]["adminScoreJournal"]
      assert data["totalCount"] >= 1
    end

    test "non-admin cannot access", %{conn: conn, user: user} do
      conn = auth_conn(conn, user.id, ["user"])

      resp =
        graphql_query(
          conn,
          """
            query {
              adminScoreJournal(first: 10) {
                totalCount
              }
            }
          """
        )

      errors = json_response(resp, 200)["errors"]
      assert errors != nil
    end
  end

  describe "createScoreAdjustment mutation" do
    test "admin creates score adjustment", %{conn: conn, project: project, user: user} do
      conn = auth_conn(conn, user.id, ["admin"])

      resp =
        graphql_query(
          conn,
          """
            mutation($input: CreateScoreAdjustmentInput!) {
              createScoreAdjustment(input: $input) {
                id
                points
                sourceType
                reason
              }
            }
          """,
          %{
            "input" => %{
              "projectId" => project.id,
              "userId" => user.id,
              "points" => 42,
              "reason" => "Manual bonus"
            }
          }
        )

      data = json_response(resp, 200)["data"]["createScoreAdjustment"]
      assert data["points"] == 42
      assert data["sourceType"] == "MANUAL"
      assert data["reason"] == "Manual bonus"
    end
  end

  describe "deleteScoreJournalEntry mutation" do
    test "admin deletes entry", %{conn: conn, project: project, user: user} do
      {:ok, entry} =
        ElixirBackend.Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 10
        })

      conn = auth_conn(conn, user.id, ["admin"])

      resp =
        graphql_query(
          conn,
          """
            mutation($id: ID!) {
              deleteScoreJournalEntry(id: $id)
            }
          """,
          %{"id" => entry.id}
        )

      data = json_response(resp, 200)["data"]["deleteScoreJournalEntry"]
      assert data == true
    end
  end

  describe "project leaderboard" do
    test "returns leaderboard via project query", %{conn: conn, project: project, user: user} do
      {:ok, entry} =
        ElixirBackend.Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 200
        })

      ElixirBackend.Scoring.update_leaderboards(entry)

      conn = auth_conn(conn, user.id, ["user"])

      resp =
        graphql_query(
          conn,
          """
            query($id: ID!) {
              project(id: $id) {
                id
                leaderboard(entityType: PERSONS, first: 10) {
                  totalCount
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
          """,
          %{"id" => project.id}
        )

      data = json_response(resp, 200)["data"]["project"]["leaderboard"]
      assert data["totalCount"] == 1
      entry = hd(data["edges"])["node"]
      assert entry["score"] == 200
      assert entry["rank"] == 1
    end
  end
end
