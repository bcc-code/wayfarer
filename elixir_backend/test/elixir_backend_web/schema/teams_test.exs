defmodule ElixirBackendWeb.Schema.TeamsTest do
  use ElixirBackendWeb.ConnCase

  alias ElixirBackend.Teams

  # ── Query: team(id) ──

  describe "team query" do
    test "returns team with members", %{conn: conn} do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team Alpha", description: "Desc"})
      user = create_user(%{name: "Member"})
      Teams.add_members(team.id, [user.id])

      query = """
      query($id: ID!) {
        team(id: $id) {
          id name description joinCode leaderboardExcluded
          members { id name isTeamLead joinedAt }
          parentProject { id }
        }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"id" => team.id})

      data = json_response(conn, 200)["data"]["team"]
      assert data["id"] == team.id
      assert data["name"] == "Team Alpha"
      assert data["joinCode"] != nil
      assert data["leaderboardExcluded"] == false
      assert length(data["members"]) == 1
      assert hd(data["members"])["name"] == "Member"
      assert data["parentProject"]["id"] == project.id
    end

    test "returns error for nonexistent team", %{conn: conn} do
      query = "query($id: ID!) { team(id: $id) { id } }"
      conn = graphql_query(conn, query, %{"id" => "TM00000000000000000000000000"})
      assert json_response(conn, 200)["errors"] != nil
    end
  end

  # ── Query: teamByJoinCode ──

  describe "teamByJoinCode query" do
    test "returns team by join code", %{conn: conn} do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})
      user = create_user()

      query = """
      query($code: String!, $projectId: ID!) {
        teamByJoinCode(code: $code, projectId: $projectId) { id name }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"code" => team.join_code, "projectId" => project.id})

      data = json_response(conn, 200)["data"]["teamByJoinCode"]
      assert data["id"] == team.id
    end

    test "returns null for non-existent code", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      query($code: String!, $projectId: ID!) {
        teamByJoinCode(code: $code, projectId: $projectId) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"code" => "BADCODE1", "projectId" => project.id})

      data = json_response(conn, 200)["data"]["teamByJoinCode"]
      assert data == nil
    end

    test "returns null when code exists but wrong project", %{conn: conn} do
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})
      {:ok, team} = Teams.create_team(p1.id, %{name: "Team"})
      user = create_user()

      query = """
      query($code: String!, $projectId: ID!) {
        teamByJoinCode(code: $code, projectId: $projectId) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"code" => team.join_code, "projectId" => p2.id})

      data = json_response(conn, 200)["data"]["teamByJoinCode"]
      assert data == nil
    end
  end

  # ── Query: teams(filter, pagination) ──

  describe "teams query" do
    test "returns paginated teams filtered by project", %{conn: conn} do
      project = create_project()
      Teams.create_team(project.id, %{name: "T1"})
      Teams.create_team(project.id, %{name: "T2"})
      user = create_user()

      query = """
      query($filter: TeamFilter, $first: Int) {
        teams(filter: $filter, first: $first) {
          edges { node { id name joinCode } }
          totalCount
        }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{
          "filter" => %{"projectId" => project.id},
          "first" => 10
        })

      data = json_response(conn, 200)["data"]["teams"]
      assert data["totalCount"] == 2
      assert length(data["edges"]) == 2
    end
  end

  # ── Mutation: createTeam ──

  describe "createTeam mutation" do
    test "admin can create team", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      mutation($projectId: ID!, $input: CreateTeamInput!) {
        createTeam(projectId: $projectId, input: $input) {
          id name description joinCode
        }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{
          "projectId" => project.id,
          "input" => %{"name" => "New Team", "description" => "A new team"}
        })

      data = json_response(conn, 200)["data"]["createTeam"]
      assert data["name"] == "New Team"
      assert String.starts_with?(data["id"], "TM")
      assert data["joinCode"] != nil
    end

    test "user cannot create team", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      mutation($projectId: ID!, $input: CreateTeamInput!) {
        createTeam(projectId: $projectId, input: $input) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{
          "projectId" => project.id,
          "input" => %{"name" => "Unauthorized"}
        })

      assert json_response(conn, 200)["errors"] != nil
    end
  end

  # ── Mutation: addTeamMembers ──

  describe "addTeamMembers mutation" do
    test "admin can add members", %{conn: conn} do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})
      admin = create_user(%{name: "Admin"})
      member = create_user(%{name: "Member"})

      query = """
      mutation($teamId: ID!, $userIds: [ID!]!) {
        addTeamMembers(teamId: $teamId, userIds: $userIds, force: true) {
          id members { id name }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{"teamId" => team.id, "userIds" => [member.id]})

      data = json_response(conn, 200)["data"]["addTeamMembers"]
      assert length(data["members"]) == 1
      assert hd(data["members"])["name"] == "Member"
    end
  end

  # ── Mutation: joinTeam ──

  describe "joinTeam mutation" do
    test "user can join team via code", %{conn: conn} do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})
      user = create_user()

      query = """
      mutation($code: ID!) {
        joinTeam(code: $code) { id name }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"code" => team.join_code})

      data = json_response(conn, 200)["data"]["joinTeam"]
      assert data["id"] == team.id
    end

    test "unauthenticated join returns error", %{conn: conn} do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})

      query = "mutation($code: ID!) { joinTeam(code: $code) { id } }"
      conn = graphql_query(conn, query, %{"code" => team.join_code})
      assert json_response(conn, 200)["errors"] != nil
    end
  end

  # ── Mutation: deleteTeam ──

  describe "deleteTeam mutation" do
    test "admin can delete team", %{conn: conn} do
      project = create_project()
      {:ok, team} = Teams.create_team(project.id, %{name: "Team"})
      user = create_user()

      query = "mutation($id: ID!) { deleteTeam(id: $id) }"

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"id" => team.id})

      assert json_response(conn, 200)["data"]["deleteTeam"] == true
    end
  end

  # ── SuperTeam queries ──

  describe "superteam query" do
    test "returns super team with teams", %{conn: conn} do
      project = create_project()
      {:ok, st} = Teams.create_super_team(project.id, %{name: "Super Team", description: "Desc"})
      {:ok, team} = Teams.create_team(project.id, %{name: "T1"})
      Teams.assign_teams_to_super_team(st.id, [team.id])

      query = """
      query($id: ID!) {
        superteam(id: $id) {
          id name description
          teams { id name }
          parentProject { id }
        }
      }
      """

      conn = graphql_query(conn, query, %{"id" => st.id})
      data = json_response(conn, 200)["data"]["superteam"]

      assert data["id"] == st.id
      assert data["name"] == "Super Team"
      assert length(data["teams"]) == 1
      assert data["parentProject"]["id"] == project.id
    end
  end

  describe "superteams query" do
    test "returns paginated super teams filtered by project", %{conn: conn} do
      project = create_project()
      Teams.create_super_team(project.id, %{name: "ST1"})
      Teams.create_super_team(project.id, %{name: "ST2"})

      query = """
      query($filter: SuperTeamFilter, $first: Int) {
        superteams(filter: $filter, first: $first) {
          edges { node { id name } }
          totalCount
        }
      }
      """

      conn =
        graphql_query(conn, query, %{
          "filter" => %{"projectId" => project.id},
          "first" => 10
        })

      data = json_response(conn, 200)["data"]["superteams"]
      assert data["totalCount"] == 2
    end
  end

  # ── SuperTeam mutations ──

  describe "createSuperTeam mutation" do
    test "admin can create super team", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      mutation($projectId: ID!, $input: CreateSuperTeamInput!) {
        createSuperTeam(projectId: $projectId, input: $input) {
          id name description color
        }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{
          "projectId" => project.id,
          "input" => %{"name" => "New ST", "description" => "Desc", "color" => "#FF0000"}
        })

      data = json_response(conn, 200)["data"]["createSuperTeam"]
      assert data["name"] == "New ST"
      assert String.starts_with?(data["id"], "ST")
      assert data["color"] == "#FF0000"
    end
  end

  describe "assignTeamsToSuperTeam mutation" do
    test "admin can assign teams to super team", %{conn: conn} do
      project = create_project()
      {:ok, st} = Teams.create_super_team(project.id, %{name: "ST"})
      {:ok, t1} = Teams.create_team(project.id, %{name: "T1"})
      {:ok, t2} = Teams.create_team(project.id, %{name: "T2"})
      user = create_user()

      query = """
      mutation($superTeamId: ID!, $teamIds: [ID!]!) {
        assignTeamsToSuperTeam(superTeamId: $superTeamId, teamIds: $teamIds) {
          id teams { id name }
        }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"superTeamId" => st.id, "teamIds" => [t1.id, t2.id]})

      data = json_response(conn, 200)["data"]["assignTeamsToSuperTeam"]
      assert data["id"] == st.id
      assert length(data["teams"]) == 2
    end
  end
end
