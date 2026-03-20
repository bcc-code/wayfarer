defmodule ElixirBackend.TeamsTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Teams

  import ElixirBackend.TestHelpers

  # ── Teams ──

  describe "create_team/2" do
    test "creates a team with auto-generated join code" do
      project = create_project()

      assert {:ok, team} =
               Teams.create_team(project.id, %{name: "Team Alpha", description: "A team"})

      assert String.starts_with?(team.id, "TM")
      assert team.name == "Team Alpha"
      assert team.join_code != nil
      assert String.length(team.join_code) == 8
      assert team.project_id == project.id
    end

    test "each team gets a unique join code" do
      project = create_project()
      {:ok, t1} = Teams.create_team(project.id, %{name: "T1"})
      {:ok, t2} = Teams.create_team(project.id, %{name: "T2"})
      assert t1.join_code != t2.join_code
    end
  end

  describe "get_team/1" do
    test "returns team by id" do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})
      assert {:ok, found} = Teams.get_team(team.id)
      assert found.id == team.id
    end

    test "returns error for nonexistent team" do
      assert {:error, :not_found} = Teams.get_team("TM00000000000000000000000000")
    end
  end

  describe "get_team_by_join_code/2" do
    test "returns team by join code and project" do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})

      assert {:ok, found} = Teams.get_team_by_join_code(team.join_code, project.id)
      assert found.id == team.id
    end

    test "returns nil for nonexistent code" do
      project = create_project()
      assert {:ok, nil} = Teams.get_team_by_join_code("BADCODE1", project.id)
    end

    test "returns nil when code exists but wrong project" do
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})
      {:ok, team} = Teams.create_team(p1.id, %{name: "Team"})

      assert {:ok, nil} = Teams.get_team_by_join_code(team.join_code, p2.id)
    end
  end

  describe "list_teams/2" do
    test "returns paginated teams" do
      project = create_project()
      Teams.create_team(project.id, %{name: "T1"})
      Teams.create_team(project.id, %{name: "T2"})
      Teams.create_team(project.id, %{name: "T3"})

      assert {:ok, result} = Teams.list_teams(%{project_id: project.id}, %{first: 2})
      assert length(result.edges) == 2
      assert result.total_count == 3
    end

    test "filters by project_id" do
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})
      Teams.create_team(p1.id, %{name: "T1"})
      Teams.create_team(p2.id, %{name: "T2"})

      assert {:ok, result} = Teams.list_teams(%{project_id: p1.id}, %{first: 10})
      assert result.total_count == 1
    end
  end

  describe "update_team/2" do
    test "updates team fields" do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Old"})

      assert {:ok, updated} = Teams.update_team(team.id, %{name: "New"})
      assert updated.name == "New"
    end
  end

  describe "delete_team/1" do
    test "deletes a team" do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "T"})
      assert {:ok, _} = Teams.delete_team(team.id)
      assert {:error, :not_found} = Teams.get_team(team.id)
    end
  end

  describe "join_team/2" do
    test "user joins team via join code" do
      project = create_project()
      user = create_user()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})

      assert {:ok, joined} = Teams.join_team(user.id, team.join_code)
      assert joined.id == team.id

      members = Teams.get_team_members(team.id)
      assert length(members) == 1
      assert hd(members).user_id == user.id
    end

    test "returns error for invalid code" do
      user = create_user()
      assert {:error, "invalid join code"} = Teams.join_team(user.id, "BADCODE1")
    end

    test "joining twice is idempotent" do
      project = create_project()
      user = create_user()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})

      Teams.join_team(user.id, team.join_code)
      Teams.join_team(user.id, team.join_code)

      members = Teams.get_team_members(team.id)
      assert length(members) == 1
    end
  end

  describe "add_members/2" do
    test "adds multiple members to a team" do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})
      u1 = create_user(%{name: "U1"})
      u2 = create_user(%{name: "U2"})

      assert {:ok, _} = Teams.add_members(team.id, [u1.id, u2.id])

      members = Teams.get_team_members(team.id)
      assert length(members) == 2
    end
  end

  describe "remove_members/2" do
    test "removes members from a team" do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})
      user = create_user()
      Teams.add_members(team.id, [user.id])

      assert {:ok, _} = Teams.remove_members(team.id, [user.id])

      members = Teams.get_team_members(team.id)
      assert members == []
    end
  end

  describe "regenerate_join_code/1" do
    test "generates a new join code" do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})
      old_code = team.join_code

      assert {:ok, updated} = Teams.regenerate_join_code(team.id)
      assert updated.join_code != old_code
    end
  end

  describe "assign_team_lead/2" do
    test "assigns team lead" do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})
      user = create_user()
      Teams.add_members(team.id, [user.id])

      assert {:ok, _} = Teams.assign_team_lead(team.id, user.id)

      members = Teams.get_team_members(team.id)
      lead = Enum.find(members, &(&1.user_id == user.id))
      assert lead.is_team_lead == true
    end

    test "returns error if user not a member" do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})
      user = create_user()

      assert {:error, "user is not a member of the team"} =
               Teams.assign_team_lead(team.id, user.id)
    end
  end

  # ── SuperTeams ──

  describe "create_super_team/2" do
    test "creates a super team" do
      project = create_project()

      assert {:ok, st} =
               Teams.create_super_team(project.id, %{
                 name: "Super Team",
                 description: "Desc",
                 color: "#FF0000"
               })

      assert String.starts_with?(st.id, "ST")
      assert st.name == "Super Team"
      assert st.project_id == project.id
    end

    test "creates super team with initial teams" do
      project = create_project()
      {:ok, t1} = Teams.create_team(project.id, %{name: "T1"})
      {:ok, t2} = Teams.create_team(project.id, %{name: "T2"})

      assert {:ok, st} =
               Teams.create_super_team(project.id, %{
                 name: "Super",
                 team_ids: [t1.id, t2.id]
               })

      # Verify teams got assigned
      {:ok, t1_updated} = Teams.get_team(t1.id)
      assert t1_updated.super_team_id == st.id
    end
  end

  describe "get_super_team/1" do
    test "returns super team by id" do
      project = create_project()
      {:ok, st} = Teams.create_super_team(project.id, %{name: "ST"})
      assert {:ok, found} = Teams.get_super_team(st.id)
      assert found.id == st.id
    end

    test "returns error for nonexistent super team" do
      assert {:error, :not_found} = Teams.get_super_team("ST00000000000000000000000000")
    end
  end

  describe "list_super_teams/2" do
    test "returns paginated super teams" do
      project = create_project()
      Teams.create_super_team(project.id, %{name: "ST1"})
      Teams.create_super_team(project.id, %{name: "ST2"})

      assert {:ok, result} = Teams.list_super_teams(%{project_id: project.id}, %{first: 10})
      assert result.total_count == 2
    end

    test "filters by project_id" do
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})
      Teams.create_super_team(p1.id, %{name: "ST1"})
      Teams.create_super_team(p2.id, %{name: "ST2"})

      assert {:ok, result} = Teams.list_super_teams(%{project_id: p1.id}, %{first: 10})
      assert result.total_count == 1
    end
  end

  describe "update_super_team/2" do
    test "updates super team fields" do
      project = create_project()
      {:ok, st} = Teams.create_super_team(project.id, %{name: "Old"})

      assert {:ok, updated} = Teams.update_super_team(st.id, %{name: "New"})
      assert updated.name == "New"
    end
  end

  describe "delete_super_team/1" do
    test "deletes a super team" do
      project = create_project()
      {:ok, st} = Teams.create_super_team(project.id, %{name: "ST"})
      assert {:ok, _} = Teams.delete_super_team(st.id)
      assert {:error, :not_found} = Teams.get_super_team(st.id)
    end

    test "nullifies team super_team_id on delete" do
      project = create_project()
      {:ok, st} = Teams.create_super_team(project.id, %{name: "ST"})
      {:ok, team} = Teams.create_team(project.id, %{name: "T"})
      Teams.assign_teams_to_super_team(st.id, [team.id])

      Teams.delete_super_team(st.id)
      {:ok, team_after} = Teams.get_team(team.id)
      assert team_after.super_team_id == nil
    end
  end

  describe "assign_teams_to_super_team/2" do
    test "assigns teams to a super team" do
      project = create_project()
      {:ok, st} = Teams.create_super_team(project.id, %{name: "ST"})
      {:ok, t1} = Teams.create_team(project.id, %{name: "T1"})
      {:ok, t2} = Teams.create_team(project.id, %{name: "T2"})

      assert {:ok, _} = Teams.assign_teams_to_super_team(st.id, [t1.id, t2.id])

      {:ok, t1_updated} = Teams.get_team(t1.id)
      {:ok, t2_updated} = Teams.get_team(t2.id)
      assert t1_updated.super_team_id == st.id
      assert t2_updated.super_team_id == st.id
    end
  end
end
