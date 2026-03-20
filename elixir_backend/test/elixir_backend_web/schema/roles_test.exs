defmodule ElixirBackendWeb.Schema.RolesTest do
  use ElixirBackendWeb.ConnCase

  alias ElixirBackend.Roles
  alias ElixirBackend.Teams

  # ── Mutation: assignRole ──

  describe "assignRole mutation" do
    test "superadmin can assign global admin role", %{conn: conn} do
      superadmin = create_user(%{name: "Superadmin"})
      target = create_user(%{name: "Target"})

      query = """
      mutation($input: AssignRoleInput!) {
        assignRole(input: $input) {
          id role scope { type id }
        }
      }
      """

      conn =
        conn
        |> auth_conn(superadmin.id, ["superadmin"])
        |> graphql_query(query, %{
          "input" => %{"userId" => target.id, "role" => "ADMIN"}
        })

      data = json_response(conn, 200)["data"]["assignRole"]
      assert data["role"] == "ADMIN"
      assert data["scope"] == nil
      assert String.starts_with?(data["id"], "UR")
    end

    test "admin can assign church admin role with scope", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      target = create_user(%{name: "Target"})
      church = create_church()

      query = """
      mutation($input: AssignRoleInput!) {
        assignRole(input: $input) {
          id role scope { type id }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "input" => %{
            "userId" => target.id,
            "role" => "CHURCH_ADMIN",
            "scopeType" => "CHURCH",
            "scopeId" => church.id
          }
        })

      data = json_response(conn, 200)["data"]["assignRole"]
      assert data["role"] == "CHURCH_ADMIN"
      assert data["scope"]["type"] == "CHURCH"
      assert data["scope"]["id"] == church.id
    end

    test "admin can assign project admin role", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      target = create_user(%{name: "Target"})
      project = create_project()

      query = """
      mutation($input: AssignRoleInput!) {
        assignRole(input: $input) {
          role scope { type id }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "input" => %{
            "userId" => target.id,
            "role" => "PROJECT_ADMIN",
            "scopeType" => "PROJECT",
            "scopeId" => project.id
          }
        })

      data = json_response(conn, 200)["data"]["assignRole"]
      assert data["role"] == "PROJECT_ADMIN"
      assert data["scope"]["type"] == "PROJECT"
    end

    test "admin can assign team lead role", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      target = create_user(%{name: "Target"})
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})

      query = """
      mutation($input: AssignRoleInput!) {
        assignRole(input: $input) {
          role scope { type id }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "input" => %{
            "userId" => target.id,
            "role" => "TEAM_LEAD",
            "scopeType" => "TEAM",
            "scopeId" => team.id
          }
        })

      data = json_response(conn, 200)["data"]["assignRole"]
      assert data["role"] == "TEAM_LEAD"
      assert data["scope"]["type"] == "TEAM"
      assert data["scope"]["id"] == team.id
    end

    test "user cannot assign roles", %{conn: conn} do
      user = create_user()
      target = create_user(%{name: "Target"})

      query = """
      mutation($input: AssignRoleInput!) {
        assignRole(input: $input) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{
          "input" => %{"userId" => target.id, "role" => "ADMIN"}
        })

      json = json_response(conn, 200)
      assert json["errors"] != nil
      assert hd(json["errors"])["message"] =~ "unauthorized"
    end
  end

  # ── Mutation: revokeRole ──

  describe "revokeRole mutation" do
    test "admin can revoke role", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      target = create_user(%{name: "Target"})
      project = create_project()

      Roles.assign_role(%{
        user_id: target.id,
        role: "PROJECT_ADMIN",
        scope_type: "PROJECT",
        scope_id: project.id
      })

      query = """
      mutation($input: RevokeRoleInput!) {
        revokeRole(input: $input)
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{
          "input" => %{
            "userId" => target.id,
            "role" => "PROJECT_ADMIN",
            "scopeType" => "PROJECT",
            "scopeId" => project.id
          }
        })

      assert json_response(conn, 200)["data"]["revokeRole"] == true
    end
  end

  # ── Query: userRoles ──

  describe "userRoles query" do
    test "returns roles for a user", %{conn: conn} do
      user = create_user()
      Roles.assign_role(%{user_id: user.id, role: "ADMIN"})

      query = """
      query($userId: ID!) {
        userRoles(userId: $userId) { id role scope { type id } }
      }
      """

      conn = graphql_query(conn, query, %{"userId" => user.id})
      data = json_response(conn, 200)["data"]["userRoles"]

      assert length(data) == 1
      assert hd(data)["role"] == "ADMIN"
    end
  end

  # ── Query: usersWithRole ──

  describe "usersWithRole query" do
    test "returns users with admin role", %{conn: conn} do
      user = create_user(%{name: "AdminUser"})
      Roles.assign_role(%{user_id: user.id, role: "ADMIN"})

      query = """
      query($role: RoleType!) {
        usersWithRole(role: $role) { id name }
      }
      """

      conn = graphql_query(conn, query, %{"role" => "ADMIN"})
      data = json_response(conn, 200)["data"]["usersWithRole"]

      assert data != []
      assert Enum.any?(data, fn u -> u["name"] == "AdminUser" end)
    end

    test "filters by scope", %{conn: conn} do
      user = create_user()
      project = create_project()

      Roles.assign_role(%{
        user_id: user.id,
        role: "PROJECT_ADMIN",
        scope_type: "PROJECT",
        scope_id: project.id
      })

      query = """
      query($role: RoleType!, $scopeType: ScopeType, $scopeId: ID) {
        usersWithRole(role: $role, scopeType: $scopeType, scopeId: $scopeId) { id }
      }
      """

      conn =
        graphql_query(conn, query, %{
          "role" => "PROJECT_ADMIN",
          "scopeType" => "PROJECT",
          "scopeId" => project.id
        })

      data = json_response(conn, 200)["data"]["usersWithRole"]
      assert data != []
    end
  end
end
