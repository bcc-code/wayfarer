defmodule ElixirBackendWeb.Schema.ProjectsTest do
  use ElixirBackendWeb.ConnCase

  # ── Query: project(id) ──

  describe "project query" do
    test "returns project by id with fields", %{conn: conn} do
      project =
        create_project(%{
          name: "Test Project",
          description: "A description"
        })

      query = """
      query($id: ID!) {
        project(id: $id) {
          id name description startDate endDate
          branding {
            rounding
            colors { light { accent } dark { accent } }
          }
        }
      }
      """

      conn = graphql_query(conn, query, %{"id" => project.id})
      data = json_response(conn, 200)["data"]["project"]

      assert data["id"] == project.id
      assert data["name"] == "Test Project"
      assert data["description"] == "A description"
      assert data["startDate"] != nil
      assert data["endDate"] != nil
      assert data["branding"]["rounding"] == 0
      assert data["branding"]["colors"]["light"]["accent"] != nil
    end

    test "returns error for nonexistent project", %{conn: conn} do
      query = "query($id: ID!) { project(id: $id) { id } }"
      conn = graphql_query(conn, query, %{"id" => "PR00000000000000000000000000"})
      assert json_response(conn, 200)["errors"] != nil
    end
  end

  # ── Query: projects(filter, pagination) ──

  describe "projects query" do
    test "returns paginated projects", %{conn: conn} do
      create_project(%{name: "P1"})
      create_project(%{name: "P2"})
      create_project(%{name: "P3"})

      query = """
      query($first: Int) {
        projects(first: $first) {
          edges { cursor node { id name } }
          pageInfo { hasNextPage hasPreviousPage }
          totalCount
        }
      }
      """

      conn = graphql_query(conn, query, %{"first" => 2})
      data = json_response(conn, 200)["data"]["projects"]

      assert length(data["edges"]) == 2
      assert data["totalCount"] >= 3
      assert data["pageInfo"]["hasNextPage"] == true
    end

    test "user can query projects list", %{conn: conn} do
      create_project()
      user = create_user()

      query = """
      query {
        projects(first: 10) { totalCount edges { node { id } } }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query)

      data = json_response(conn, 200)["data"]["projects"]
      assert data["totalCount"] > 0
    end
  end

  # ── Query: myProjects ──

  describe "myProjects query" do
    test "returns projects user has joined", %{conn: conn} do
      user = create_user()
      project = create_project(%{name: "Joined Project"})
      _other = create_project(%{name: "Not Joined"})
      create_user_project(user, project)

      query = """
      query { myProjects { id name } }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query)

      data = json_response(conn, 200)["data"]["myProjects"]
      assert length(data) == 1
      assert hd(data)["id"] == project.id
    end

    test "unauthenticated returns error", %{conn: conn} do
      query = "query { myProjects { id } }"
      conn = graphql_query(conn, query)
      assert json_response(conn, 200)["errors"] != nil
    end
  end

  # ── Mutation: createProject ──

  describe "createProject mutation" do
    test "admin can create project", %{conn: conn} do
      user = create_user()

      query = """
      mutation($input: CreateProjectInput!) {
        createProject(input: $input) {
          id name description startDate endDate
          branding { rounding colors { light { accent } } }
        }
      }
      """

      variables = %{
        "input" => %{
          "name" => "E2E Test Project",
          "description" => "Desc",
          "startDate" => "2026-01-01T00:00:00Z",
          "endDate" => "2026-12-31T23:59:59Z",
          "branding" => default_branding_input()
        }
      }

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, variables)

      data = json_response(conn, 200)["data"]["createProject"]
      assert data["name"] == "E2E Test Project"
      assert String.starts_with?(data["id"], "PR")
      assert data["branding"]["rounding"] == 8
    end

    test "superadmin can create project", %{conn: conn} do
      user = create_user()

      query = """
      mutation($input: CreateProjectInput!) {
        createProject(input: $input) { id name }
      }
      """

      variables = %{
        "input" => %{
          "name" => "Superadmin Project",
          "startDate" => "2026-03-01T00:00:00Z",
          "endDate" => "2026-09-30T23:59:59Z",
          "branding" => default_branding_input()
        }
      }

      conn =
        conn
        |> auth_conn(user.id, ["superadmin"])
        |> graphql_query(query, variables)

      data = json_response(conn, 200)["data"]["createProject"]
      assert data["name"] == "Superadmin Project"
      assert String.starts_with?(data["id"], "PR")
    end

    test "user cannot create project", %{conn: conn} do
      user = create_user()

      query = """
      mutation($input: CreateProjectInput!) {
        createProject(input: $input) { id }
      }
      """

      variables = %{
        "input" => %{
          "name" => "Unauthorized",
          "startDate" => "2026-01-01T00:00:00Z",
          "endDate" => "2026-12-31T23:59:59Z",
          "branding" => default_branding_input()
        }
      }

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, variables)

      json = json_response(conn, 200)
      assert json["errors"] != nil
      assert hd(json["errors"])["message"] =~ "unauthorized"
    end
  end

  # ── Mutation: updateProject ──

  describe "updateProject mutation" do
    test "admin can update project", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      mutation($id: ID!, $input: UpdateProjectInput!) {
        updateProject(id: $id, input: $input) { id name description }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{
          "id" => project.id,
          "input" => %{"name" => "Updated Name", "description" => "Updated desc"}
        })

      data = json_response(conn, 200)["data"]["updateProject"]
      assert data["name"] == "Updated Name"
      assert data["description"] == "Updated desc"
    end
  end

  # ── Mutation: deleteProject ──

  describe "deleteProject mutation" do
    test "admin can delete project", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = "mutation($id: ID!) { deleteProject(id: $id) }"

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"id" => project.id})

      assert json_response(conn, 200)["data"]["deleteProject"] == true
    end
  end

  # ── Mutation: archiveProject ──

  describe "archiveProject mutation" do
    test "admin can archive project", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = "mutation($id: ID!) { archiveProject(id: $id) }"

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"id" => project.id})

      assert json_response(conn, 200)["data"]["archiveProject"] == true
    end
  end

  # ── Mutation: joinProject ──

  describe "joinProject mutation" do
    test "user can join project", %{conn: conn} do
      user = create_user()
      project = create_project()

      query = """
      mutation($projectId: ID!) {
        joinProject(projectId: $projectId) { id name }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"projectId" => project.id})

      data = json_response(conn, 200)["data"]["joinProject"]
      assert data["id"] == project.id
    end

    test "unauthenticated join returns error", %{conn: conn} do
      project = create_project()

      query = """
      mutation($projectId: ID!) {
        joinProject(projectId: $projectId) { id }
      }
      """

      conn = graphql_query(conn, query, %{"projectId" => project.id})
      assert json_response(conn, 200)["errors"] != nil
    end
  end
end
