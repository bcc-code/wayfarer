defmodule ElixirBackend.RolesTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Roles
  alias ElixirBackend.Teams

  import ElixirBackend.TestHelpers

  describe "assign_role/1" do
    test "assigns global admin role" do
      user = create_user()

      assert {:ok, role} =
               Roles.assign_role(%{user_id: user.id, role: "ADMIN"})

      assert String.starts_with?(role.id, "UR")
      assert role.role == "ADMIN"
      assert role.user_id == user.id
      assert role.church_id == nil
      assert role.project_id == nil
      assert role.team_id == nil
    end

    test "assigns church admin role with scope" do
      user = create_user()
      church = create_church()

      assert {:ok, role} =
               Roles.assign_role(%{
                 user_id: user.id,
                 role: "CHURCH_ADMIN",
                 scope_type: "CHURCH",
                 scope_id: church.id
               })

      assert role.role == "CHURCH_ADMIN"
      assert role.church_id == church.id
    end

    test "assigns project admin role with scope" do
      user = create_user()
      project = create_project()

      assert {:ok, role} =
               Roles.assign_role(%{
                 user_id: user.id,
                 role: "PROJECT_ADMIN",
                 scope_type: "PROJECT",
                 scope_id: project.id
               })

      assert role.role == "PROJECT_ADMIN"
      assert role.project_id == project.id
    end

    test "assigns team lead role with scope" do
      user = create_user()
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})

      assert {:ok, role} =
               Roles.assign_role(%{
                 user_id: user.id,
                 role: "TEAM_LEAD",
                 scope_type: "TEAM",
                 scope_id: team.id
               })

      assert role.role == "TEAM_LEAD"
      assert role.team_id == team.id
    end

    test "rejects global role with scope" do
      user = create_user()
      church = create_church()

      assert {:error, changeset} =
               Roles.assign_role(%{
                 user_id: user.id,
                 role: "ADMIN",
                 scope_type: "CHURCH",
                 scope_id: church.id
               })

      assert errors_on(changeset)[:role] != nil
    end

    test "rejects church admin without scope" do
      user = create_user()

      assert {:error, changeset} =
               Roles.assign_role(%{user_id: user.id, role: "CHURCH_ADMIN"})

      assert errors_on(changeset)[:church_id] != nil
    end

    test "prevents duplicate assignment" do
      user = create_user()
      assert {:ok, _} = Roles.assign_role(%{user_id: user.id, role: "ADMIN"})
      assert {:error, _} = Roles.assign_role(%{user_id: user.id, role: "ADMIN"})
    end
  end

  describe "revoke_role/1" do
    test "revokes a global role" do
      user = create_user()
      Roles.assign_role(%{user_id: user.id, role: "ADMIN"})

      assert {:ok, true} = Roles.revoke_role(%{user_id: user.id, role: "ADMIN"})

      {:ok, roles} = Roles.list_user_roles(user.id)
      assert Enum.empty?(roles)
    end

    test "revokes a scoped role" do
      user = create_user()
      project = create_project()

      Roles.assign_role(%{
        user_id: user.id,
        role: "PROJECT_ADMIN",
        scope_type: "PROJECT",
        scope_id: project.id
      })

      assert {:ok, true} =
               Roles.revoke_role(%{
                 user_id: user.id,
                 role: "PROJECT_ADMIN",
                 scope_type: "PROJECT",
                 scope_id: project.id
               })

      {:ok, roles} = Roles.list_user_roles(user.id)
      assert Enum.empty?(roles)
    end

    test "returns false for non-existent role" do
      user = create_user()
      assert {:ok, false} = Roles.revoke_role(%{user_id: user.id, role: "ADMIN"})
    end
  end

  describe "list_user_roles/1" do
    test "returns all roles for user" do
      user = create_user()
      project = create_project()

      Roles.assign_role(%{user_id: user.id, role: "ADMIN"})

      Roles.assign_role(%{
        user_id: user.id,
        role: "PROJECT_ADMIN",
        scope_type: "PROJECT",
        scope_id: project.id
      })

      {:ok, roles} = Roles.list_user_roles(user.id)
      assert length(roles) == 2
    end

    test "returns empty list for user with no roles" do
      user = create_user()
      {:ok, roles} = Roles.list_user_roles(user.id)
      assert Enum.empty?(roles)
    end
  end

  describe "users_with_role/2" do
    test "returns users with a specific role" do
      u1 = create_user(%{name: "Admin1"})
      u2 = create_user(%{name: "Admin2"})
      _u3 = create_user(%{name: "Regular"})

      Roles.assign_role(%{user_id: u1.id, role: "ADMIN"})
      Roles.assign_role(%{user_id: u2.id, role: "ADMIN"})

      {:ok, users} = Roles.users_with_role("ADMIN")
      assert length(users) == 2
    end

    test "filters by scope" do
      user = create_user()
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})

      Roles.assign_role(%{
        user_id: user.id,
        role: "PROJECT_ADMIN",
        scope_type: "PROJECT",
        scope_id: p1.id
      })

      {:ok, users_p1} =
        Roles.users_with_role("PROJECT_ADMIN", scope_type: "PROJECT", scope_id: p1.id)

      {:ok, users_p2} =
        Roles.users_with_role("PROJECT_ADMIN", scope_type: "PROJECT", scope_id: p2.id)

      assert length(users_p1) == 1
      assert length(users_p2) == 0
    end
  end
end
