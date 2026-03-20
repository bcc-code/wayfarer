defmodule ElixirBackendWeb.Schema.AchievementsTest do
  use ElixirBackendWeb.ConnCase

  alias ElixirBackend.Achievements
  alias ElixirBackend.ExternalContent, as: EC

  defp create_external_content(attrs \\ %{}) do
    defaults = %{
      plan_id: "plan-#{System.unique_integer([:positive])}",
      task_id: "task-#{System.unique_integer([:positive])}",
      content_type: "ARTICLE",
      source: "ssf"
    }

    {:ok, c} = EC.upsert_content(Map.merge(defaults, attrs))
    c
  end

  # ── Mutation: createSimpleAchievement ──

  describe "createSimpleAchievement mutation" do
    test "admin can create simple achievement", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      project = create_project()

      query = """
      mutation($input: CreateSimpleAchievementInput!) {
        createSimpleAchievement(input: $input) {
          id name points hidden descriptionPending descriptionCompleted
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "input" => %{
            "name" => "First Blood",
            "descriptionPending" => "Not yet",
            "descriptionCompleted" => "Done!",
            "imagePending" => "pending.png",
            "imageCompleted" => "done.png",
            "projectId" => project.id,
            "points" => 50,
            "hidden" => false
          }
        })

      data = json_response(conn, 200)["data"]["createSimpleAchievement"]
      assert data["name"] == "First Blood"
      assert data["points"] == 50
      assert String.starts_with?(data["id"], "AC")
    end

    test "user cannot create achievement", %{conn: conn} do
      user = create_user()
      project = create_project()

      query = """
      mutation($input: CreateSimpleAchievementInput!) {
        createSimpleAchievement(input: $input) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{
          "input" => %{
            "name" => "X",
            "descriptionPending" => "X",
            "descriptionCompleted" => "X",
            "imagePending" => "x",
            "imageCompleted" => "x",
            "projectId" => project.id,
            "points" => 10,
            "hidden" => false
          }
        })

      json = json_response(conn, 200)
      assert json["errors"] != nil
      assert hd(json["errors"])["message"] =~ "unauthorized"
    end
  end

  # ── Mutation: createContentAchievement ──

  describe "createContentAchievement mutation" do
    test "admin can create content achievement with items", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      project = create_project()
      ec = create_external_content()

      query = """
      mutation($input: CreateContentAchievementInput!) {
        createContentAchievement(input: $input) {
          id name totalItems
          items { id }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "input" => %{
            "name" => "Reader",
            "descriptionPending" => "Read all",
            "descriptionCompleted" => "All read",
            "imagePending" => "p.png",
            "imageCompleted" => "c.png",
            "projectId" => project.id,
            "points" => 20,
            "hidden" => false,
            "items" => [%{"externalContentId" => ec.id}]
          }
        })

      data = json_response(conn, 200)["data"]["createContentAchievement"]
      assert data["name"] == "Reader"
      assert data["totalItems"] == 1
      assert length(data["items"]) == 1
    end
  end

  # ── Mutation: awardAchievement ──

  describe "awardAchievement mutation" do
    test "admin can award achievement to user", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      user = create_user(%{name: "Target"})
      project = create_project()

      {:ok, a} =
        Achievements.create_simple_achievement(%{
          name: "Award Me",
          description_pending: "P",
          description_completed: "C",
          image_pending: "p.png",
          image_completed: "c.png",
          project_id: project.id,
          points: 10,
          hidden: false
        })

      query = """
      mutation($userId: ID!, $achievementId: ID!) {
        awardAchievement(userId: $userId, achievementId: $achievementId) {
          id name
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "userId" => user.id,
          "achievementId" => a.id
        })

      data = json_response(conn, 200)["data"]["awardAchievement"]
      assert data["id"] == a.id
    end

    test "awarding with future awardableFrom fails", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      user = create_user(%{name: "Target"})
      project = create_project()
      future = DateTime.utc_now() |> DateTime.add(86_400) |> DateTime.truncate(:second)

      {:ok, a} =
        Achievements.create_simple_achievement(%{
          name: "Future",
          description_pending: "P",
          description_completed: "C",
          image_pending: "p.png",
          image_completed: "c.png",
          project_id: project.id,
          points: 10,
          hidden: false,
          awardable_from: future
        })

      query = """
      mutation($userId: ID!, $achievementId: ID!) {
        awardAchievement(userId: $userId, achievementId: $achievementId) { id }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "userId" => user.id,
          "achievementId" => a.id
        })

      json = json_response(conn, 200)
      assert json["errors"] != nil
    end
  end

  # ── Mutation: revokeAchievement ──

  describe "revokeAchievement mutation" do
    test "admin can revoke achievement", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      user = create_user(%{name: "Target"})
      project = create_project()

      {:ok, a} =
        Achievements.create_simple_achievement(%{
          name: "Revoke Me",
          description_pending: "P",
          description_completed: "C",
          image_pending: "p.png",
          image_completed: "c.png",
          project_id: project.id,
          points: 10,
          hidden: false
        })

      Achievements.award_achievement(user.id, a.id)

      query = """
      mutation($userId: ID!, $achievementId: ID!) {
        revokeAchievement(userId: $userId, achievementId: $achievementId)
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "userId" => user.id,
          "achievementId" => a.id
        })

      assert json_response(conn, 200)["data"]["revokeAchievement"] == true
    end
  end

  # ── Query: achievement ──

  describe "achievement query" do
    test "returns achievement by id", %{conn: conn} do
      project = create_project()

      {:ok, a} =
        Achievements.create_simple_achievement(%{
          name: "Query Me",
          description_pending: "P",
          description_completed: "C",
          image_pending: "p.png",
          image_completed: "c.png",
          project_id: project.id,
          points: 10,
          hidden: false
        })

      query = """
      query($id: ID!) {
        achievement(id: $id) {
          ... on SimpleAchievement {
            id name points hidden
          }
        }
      }
      """

      conn = graphql_query(conn, query, %{"id" => a.id})
      data = json_response(conn, 200)["data"]["achievement"]
      assert data["id"] == a.id
      assert data["name"] == "Query Me"
    end
  end

  # ── Query: achievements ──

  describe "achievements query" do
    test "returns paginated achievements with filter", %{conn: conn} do
      project = create_project()

      Achievements.create_simple_achievement(%{
        name: "A1",
        description_pending: "P",
        description_completed: "C",
        image_pending: "p.png",
        image_completed: "c.png",
        project_id: project.id,
        points: 10,
        hidden: false
      })

      Achievements.create_simple_achievement(%{
        name: "A2",
        description_pending: "P",
        description_completed: "C",
        image_pending: "p.png",
        image_completed: "c.png",
        project_id: project.id,
        points: 20,
        hidden: false
      })

      query = """
      query($filter: AchievementFilter!) {
        achievements(filter: $filter, first: 10) {
          totalCount
          edges {
            node {
              ... on SimpleAchievement { id name points }
            }
          }
        }
      }
      """

      conn = graphql_query(conn, query, %{"filter" => %{"projectId" => project.id}})
      data = json_response(conn, 200)["data"]["achievements"]
      assert data["totalCount"] == 2
    end
  end

  # ── Mutation: updateAchievement ──

  describe "updateAchievement mutation" do
    test "admin can update achievement", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      project = create_project()

      {:ok, a} =
        Achievements.create_simple_achievement(%{
          name: "Original",
          description_pending: "P",
          description_completed: "C",
          image_pending: "p.png",
          image_completed: "c.png",
          project_id: project.id,
          points: 10,
          hidden: false
        })

      query = """
      mutation($id: ID!, $input: UpdateAchievementInput!) {
        updateAchievement(id: $id, input: $input) {
          ... on SimpleAchievement { id name points }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "id" => a.id,
          "input" => %{"name" => "Updated", "points" => 99}
        })

      data = json_response(conn, 200)["data"]["updateAchievement"]
      assert data["name"] == "Updated"
      assert data["points"] == 99
    end
  end

  # ── Mutation: deleteAchievement ──

  describe "deleteAchievement mutation" do
    test "admin can delete achievement", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      project = create_project()

      {:ok, a} =
        Achievements.create_simple_achievement(%{
          name: "Delete Me",
          description_pending: "P",
          description_completed: "C",
          image_pending: "p.png",
          image_completed: "c.png",
          project_id: project.id,
          points: 10,
          hidden: false
        })

      query = """
      mutation($id: ID!) {
        deleteAchievement(id: $id)
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{"id" => a.id})

      assert json_response(conn, 200)["data"]["deleteAchievement"] == true
    end
  end

  # ── Mutation: reorderAchievements ──

  describe "reorderAchievements mutation" do
    test "admin can reorder achievements", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      project = create_project()

      {:ok, a1} =
        Achievements.create_simple_achievement(%{
          name: "First",
          description_pending: "P",
          description_completed: "C",
          image_pending: "p.png",
          image_completed: "c.png",
          project_id: project.id,
          points: 10,
          hidden: false
        })

      {:ok, a2} =
        Achievements.create_simple_achievement(%{
          name: "Second",
          description_pending: "P",
          description_completed: "C",
          image_pending: "p.png",
          image_completed: "c.png",
          project_id: project.id,
          points: 20,
          hidden: false
        })

      query = """
      mutation($projectId: ID!, $ids: [ID!]!) {
        reorderAchievements(projectId: $projectId, achievementIds: $ids) {
          ... on SimpleAchievement { id name }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "projectId" => project.id,
          "ids" => [a2.id, a1.id]
        })

      data = json_response(conn, 200)["data"]["reorderAchievements"]
      assert length(data) == 2
      assert hd(data)["id"] == a2.id
    end
  end

  # ── Mutation: markContentItemCompleted ──

  describe "markContentItemCompleted mutation" do
    test "m2m can mark content completed", %{conn: conn} do
      m2m = create_user(%{name: "M2M"})
      user = create_user(%{name: "Target"})
      project = create_project()
      ec = create_external_content()

      Achievements.create_content_achievement(%{
        name: "Content",
        description_pending: "P",
        description_completed: "C",
        image_pending: "p.png",
        image_completed: "c.png",
        project_id: project.id,
        points: 10,
        hidden: false,
        items: [%{external_content_id: ec.id}]
      })

      query = """
      mutation($userId: ID!, $ecId: ID!) {
        markContentItemCompleted(userId: $userId, externalContentId: $ecId) {
          id name
        }
      }
      """

      conn =
        conn
        |> auth_conn(m2m.id, ["m2m"])
        |> graphql_query(query, %{
          "userId" => user.id,
          "ecId" => ec.id
        })

      data = json_response(conn, 200)["data"]["markContentItemCompleted"]
      assert length(data) == 1
      assert hd(data)["name"] == "Content"
    end
  end
end
