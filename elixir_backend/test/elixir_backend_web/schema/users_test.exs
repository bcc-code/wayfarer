defmodule ElixirBackendWeb.Schema.UsersTest do
  use ElixirBackendWeb.ConnCase

  import ElixirBackend.TestHelpers

  # ── me query ──

  describe "me query" do
    test "returns current user with all fields", %{conn: conn} do
      church = create_church(%{name: "My Church"})

      user =
        create_user(%{
          name: "Me User",
          email: "me@example.com",
          gender: "FEMALE",
          birthdate: ~D[2000-01-15],
          church_id: church.id,
          display_name: "My Display Name",
          language: "nb"
        })

      conn = auth_conn(conn, user.id, ["user"])

      conn =
        graphql_query(conn, """
          query {
            me {
              id name email gender age birthdate
              displayName language churchId
              createdAt
            }
          }
        """)

      assert %{"data" => %{"me" => me}} = json_response(conn, 200)
      assert me["id"] == user.id
      assert me["name"] == "Me User"
      assert me["email"] == "me@example.com"
      assert me["gender"] == "FEMALE"
      assert me["age"] == 26
      assert me["birthdate"] == "2000-01-15"
      assert me["displayName"] == "My Display Name"
      assert me["language"] == "nb"
      assert me["churchId"] == church.id
      assert me["createdAt"] != nil
    end

    test "returns error when unauthenticated", %{conn: conn} do
      conn =
        graphql_query(conn, """
          query { me { id name } }
        """)

      assert %{"errors" => _} = json_response(conn, 200)
    end
  end

  # ── user query ──

  describe "user query" do
    test "admin can access any user", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      target = create_user(%{name: "Target User", birthdate: ~D[2000-06-15]})

      conn = auth_conn(conn, admin.id, ["admin"])

      conn =
        graphql_query(
          conn,
          """
            query($id: ID!) { user(id: $id) { id name age birthdate } }
          """,
          %{id: target.id}
        )

      assert %{"data" => %{"user" => u}} = json_response(conn, 200)
      assert u["id"] == target.id
      assert u["name"] == "Target User"
      assert is_integer(u["age"])
    end

    test "m2m can access any user", %{conn: conn} do
      target = create_user(%{name: "M2M Target"})

      conn = auth_conn(conn, "US_M2M_SERVICE", ["m2m"])

      conn =
        graphql_query(
          conn,
          """
            query($id: ID!) { user(id: $id) { id name } }
          """,
          %{id: target.id}
        )

      assert %{"data" => %{"user" => u}} = json_response(conn, 200)
      assert u["id"] == target.id
    end

    test "regular user can access themselves", %{conn: conn} do
      user = create_user(%{name: "Self"})
      conn = auth_conn(conn, user.id, ["user"])

      conn =
        graphql_query(
          conn,
          """
            query($id: ID!) { user(id: $id) { id name } }
          """,
          %{id: user.id}
        )

      assert %{"data" => %{"user" => u}} = json_response(conn, 200)
      assert u["id"] == user.id
    end

    test "regular user denied for other user — returns null", %{conn: conn} do
      user = create_user(%{name: "Regular"})
      other = create_user(%{name: "Other"})

      conn = auth_conn(conn, user.id, ["user"])

      conn =
        graphql_query(
          conn,
          """
            query($id: ID!) { user(id: $id) { id } }
          """,
          %{id: other.id}
        )

      resp = json_response(conn, 200)
      assert resp["data"]["user"] == nil || resp["errors"] != nil
    end

    test "non-existent user returns null/error", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      conn = auth_conn(conn, admin.id, ["admin"])

      conn =
        graphql_query(
          conn,
          """
            query($id: ID!) { user(id: $id) { id } }
          """,
          %{id: "US00000000000000000000000000"}
        )

      resp = json_response(conn, 200)
      assert resp["data"]["user"] == nil || resp["errors"] != nil
    end
  end

  # ── users query ──

  describe "users query" do
    test "admin can list users with pagination", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      _user2 = create_user(%{name: "User2"})
      _user3 = create_user(%{name: "User3"})

      conn = auth_conn(conn, admin.id, ["admin"])

      conn =
        graphql_query(conn, """
          query {
            users(first: 10) {
              totalCount
              edges { cursor node { id name } }
              pageInfo { hasNextPage hasPreviousPage startCursor endCursor }
            }
          }
        """)

      assert %{"data" => %{"users" => result}} = json_response(conn, 200)
      assert result["totalCount"] >= 3
      assert length(result["edges"]) >= 3
      assert result["pageInfo"]["startCursor"] != nil
      assert result["pageInfo"]["endCursor"] != nil
    end

    test "regular user denied from listing", %{conn: conn} do
      user = create_user(%{name: "Regular"})
      conn = auth_conn(conn, user.id, ["user"])

      conn =
        graphql_query(conn, """
          query { users(first: 10) { totalCount edges { node { id } } } }
        """)

      resp = json_response(conn, 200)
      assert resp["errors"] != nil
    end

    test "with gender filter", %{conn: conn} do
      admin = create_user(%{name: "Admin", gender: "MALE"})
      _female = create_user(%{name: "Female User", gender: "FEMALE"})

      conn = auth_conn(conn, admin.id, ["admin"])

      conn =
        graphql_query(
          conn,
          """
            query($filter: UserFilter) {
              users(first: 10, filter: $filter) {
                totalCount
                edges { node { id name gender } }
              }
            }
          """,
          %{filter: %{gender: "FEMALE"}}
        )

      assert %{"data" => %{"users" => result}} = json_response(conn, 200)
      assert result["totalCount"] >= 1
      assert Enum.all?(result["edges"], fn e -> e["node"]["gender"] == "FEMALE" end)
    end

    test "with text query filter", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      _target = create_user(%{name: "UniqueSearchableName"})

      conn = auth_conn(conn, admin.id, ["admin"])

      conn =
        graphql_query(
          conn,
          """
            query($filter: UserFilter) {
              users(first: 10, filter: $filter) {
                totalCount
                edges { node { id name } }
              }
            }
          """,
          %{filter: %{query: "UniqueSearchable"}}
        )

      assert %{"data" => %{"users" => result}} = json_response(conn, 200)
      assert result["totalCount"] == 1
      assert hd(result["edges"])["node"]["name"] == "UniqueSearchableName"
    end

    test "with church_id filter", %{conn: conn} do
      church = create_church(%{name: "Specific Church"})
      admin = create_user(%{name: "Admin", church_id: church.id})
      _other = create_user(%{name: "Other Church User"})

      conn = auth_conn(conn, admin.id, ["admin"])

      conn =
        graphql_query(
          conn,
          """
            query($filter: UserFilter) {
              users(first: 10, filter: $filter) { totalCount edges { node { id churchId } } }
            }
          """,
          %{filter: %{churchId: church.id}}
        )

      assert %{"data" => %{"users" => result}} = json_response(conn, 200)
      assert Enum.all?(result["edges"], fn e -> e["node"]["churchId"] == church.id end)
    end

    test "pagination cursor works across pages", %{conn: conn} do
      church = create_church()

      for i <- 1..5 do
        create_user(%{name: "PagUser #{i}", church_id: church.id})
      end

      admin = create_user(%{name: "PagAdmin", church_id: church.id})
      conn = auth_conn(conn, admin.id, ["admin"])

      conn1 =
        graphql_query(
          conn,
          """
            query($filter: UserFilter) {
              users(first: 3, filter: $filter) {
                edges { node { id } }
                pageInfo { endCursor hasNextPage }
              }
            }
          """,
          %{filter: %{churchId: church.id}}
        )

      result1 = json_response(conn1, 200)["data"]["users"]
      assert result1["pageInfo"]["hasNextPage"] == true
      end_cursor = result1["pageInfo"]["endCursor"]

      conn2 =
        graphql_query(
          conn,
          """
            query($filter: UserFilter, $after: String) {
              users(first: 3, after: $after, filter: $filter) {
                edges { node { id } }
                pageInfo { hasNextPage }
              }
            }
          """,
          %{filter: %{churchId: church.id}, after: end_cursor}
        )

      result2 = json_response(conn2, 200)["data"]["users"]

      ids1 = Enum.map(result1["edges"], & &1["node"]["id"])
      ids2 = Enum.map(result2["edges"], & &1["node"]["id"])
      assert MapSet.disjoint?(MapSet.new(ids1), MapSet.new(ids2))
    end
  end

  # ── Mutations ──

  describe "assignUserToProject mutation" do
    test "admin can assign user to project", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      target = create_user(%{name: "Target"})
      project = create_project()

      conn = auth_conn(conn, admin.id, ["admin"])

      conn =
        graphql_query(
          conn,
          """
            mutation($userId: ID!, $projectId: ID!) {
              assignUserToProject(userId: $userId, projectId: $projectId) { id name }
            }
          """,
          %{userId: target.id, projectId: project.id}
        )

      assert %{"data" => %{"assignUserToProject" => result}} = json_response(conn, 200)
      assert result["id"] == target.id
    end

    test "m2m can assign user to project", %{conn: conn} do
      target = create_user(%{name: "Target"})
      project = create_project()

      conn = auth_conn(conn, "US_M2M", ["m2m"])

      conn =
        graphql_query(
          conn,
          """
            mutation($userId: ID!, $projectId: ID!) {
              assignUserToProject(userId: $userId, projectId: $projectId) { id }
            }
          """,
          %{userId: target.id, projectId: project.id}
        )

      assert %{"data" => %{"assignUserToProject" => _}} = json_response(conn, 200)
    end

    test "regular user denied", %{conn: conn} do
      user = create_user(%{name: "Regular"})
      project = create_project()

      conn = auth_conn(conn, user.id, ["user"])

      conn =
        graphql_query(
          conn,
          """
            mutation($userId: ID!, $projectId: ID!) {
              assignUserToProject(userId: $userId, projectId: $projectId) { id }
            }
          """,
          %{userId: user.id, projectId: project.id}
        )

      assert %{"errors" => _} = json_response(conn, 200)
    end
  end

  describe "removeUserFromProject mutation" do
    test "removes user from project", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      target = create_user(%{name: "Target"})
      project = create_project()
      create_user_project(target, project)

      conn = auth_conn(conn, admin.id, ["admin"])

      conn =
        graphql_query(
          conn,
          """
            mutation($userId: ID!, $projectId: ID!) {
              removeUserFromProject(userId: $userId, projectId: $projectId) { id }
            }
          """,
          %{userId: target.id, projectId: project.id}
        )

      assert %{"data" => %{"removeUserFromProject" => result}} = json_response(conn, 200)
      assert result["id"] == target.id
    end
  end

  describe "assignUserToEvent mutation" do
    test "admin can assign user to event", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      target = create_user(%{name: "Target"})
      project = create_project()
      event = create_event(project)

      conn = auth_conn(conn, admin.id, ["admin"])

      conn =
        graphql_query(
          conn,
          """
            mutation($userId: ID!, $eventId: ID!) {
              assignUserToEvent(userId: $userId, eventId: $eventId) { id }
            }
          """,
          %{userId: target.id, eventId: event.id}
        )

      assert %{"data" => %{"assignUserToEvent" => result}} = json_response(conn, 200)
      assert result["id"] == target.id
    end

    test "regular user denied", %{conn: conn} do
      user = create_user(%{name: "Regular"})
      project = create_project()
      event = create_event(project)

      conn = auth_conn(conn, user.id, ["user"])

      conn =
        graphql_query(
          conn,
          """
            mutation($userId: ID!, $eventId: ID!) {
              assignUserToEvent(userId: $userId, eventId: $eventId) { id }
            }
          """,
          %{userId: user.id, eventId: event.id}
        )

      assert %{"errors" => _} = json_response(conn, 200)
    end
  end

  describe "lockUserChurch / unlockUserChurch mutations" do
    test "admin can lock and unlock user church", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      target = create_user(%{name: "Target"})

      conn = auth_conn(conn, admin.id, ["admin"])

      # Lock
      conn_lock =
        graphql_query(
          conn,
          """
            mutation($userId: ID!) {
              lockUserChurch(userId: $userId) { id churchLockedUntil }
            }
          """,
          %{userId: target.id}
        )

      assert %{"data" => %{"lockUserChurch" => result}} = json_response(conn_lock, 200)
      assert result["id"] == target.id
      assert result["churchLockedUntil"] != nil

      # Unlock
      conn_unlock =
        graphql_query(
          conn,
          """
            mutation($userId: ID!) {
              unlockUserChurch(userId: $userId) { id churchLockedUntil }
            }
          """,
          %{userId: target.id}
        )

      assert %{"data" => %{"unlockUserChurch" => result}} = json_response(conn_unlock, 200)
      assert result["churchLockedUntil"] == nil
    end

    test "regular user denied from lock", %{conn: conn} do
      user = create_user(%{name: "Regular"})

      conn = auth_conn(conn, user.id, ["user"])

      conn =
        graphql_query(
          conn,
          """
            mutation($userId: ID!) {
              lockUserChurch(userId: $userId) { id }
            }
          """,
          %{userId: user.id}
        )

      assert %{"errors" => _} = json_response(conn, 200)
    end

    test "regular user denied from unlock", %{conn: conn} do
      user = create_user(%{name: "Regular"})

      conn = auth_conn(conn, user.id, ["user"])

      conn =
        graphql_query(
          conn,
          """
            mutation($userId: ID!) {
              unlockUserChurch(userId: $userId) { id }
            }
          """,
          %{userId: user.id}
        )

      assert %{"errors" => _} = json_response(conn, 200)
    end
  end
end
