defmodule ElixirBackendWeb.Schema.EventsTest do
  use ElixirBackendWeb.ConnCase

  # ── Query: event(id) ──

  describe "event query" do
    test "returns event by id with parent project", %{conn: conn} do
      project = create_project(%{name: "Parent Project"})
      event = create_event(project, %{name: "Test Event", description: "Desc"})

      query = """
      query($id: ID!) {
        event(id: $id) {
          id name description startDate endDate
          parentProject { id name }
        }
      }
      """

      conn = graphql_query(conn, query, %{"id" => event.id})
      data = json_response(conn, 200)["data"]["event"]

      assert data["id"] == event.id
      assert data["name"] == "Test Event"
      assert data["description"] == "Desc"
      assert data["parentProject"]["id"] == project.id
      assert data["parentProject"]["name"] == "Parent Project"
    end

    test "returns error for nonexistent event", %{conn: conn} do
      query = "query($id: ID!) { event(id: $id) { id } }"
      conn = graphql_query(conn, query, %{"id" => "EV00000000000000000000000000"})
      assert json_response(conn, 200)["errors"] != nil
    end
  end

  # ── Query: events(filter, pagination) ──

  describe "events query" do
    test "returns paginated events", %{conn: conn} do
      project = create_project()
      create_event(project, %{name: "E1"})
      create_event(project, %{name: "E2"})
      create_event(project, %{name: "E3"})

      query = """
      query($first: Int) {
        events(first: $first) {
          edges { cursor node { id name } }
          pageInfo { hasNextPage }
          totalCount
        }
      }
      """

      conn = graphql_query(conn, query, %{"first" => 2})
      data = json_response(conn, 200)["data"]["events"]

      assert length(data["edges"]) == 2
      assert data["totalCount"] >= 3
      assert data["pageInfo"]["hasNextPage"] == true
    end

    test "filters by projectId", %{conn: conn} do
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})
      create_event(p1, %{name: "E1"})
      create_event(p2, %{name: "E2"})

      query = """
      query($filter: EventFilter) {
        events(filter: $filter, first: 10) {
          edges { node { id name } }
          totalCount
        }
      }
      """

      conn = graphql_query(conn, query, %{"filter" => %{"projectId" => p1.id}})
      data = json_response(conn, 200)["data"]["events"]

      assert data["totalCount"] == 1
      assert hd(data["edges"])["node"]["name"] == "E1"
    end
  end

  # ── Query: myEvents ──

  describe "myEvents query" do
    test "returns events user has joined", %{conn: conn} do
      user = create_user()
      project = create_project()
      event = create_event(project, %{name: "Joined Event"})
      _other = create_event(project, %{name: "Not Joined"})
      create_user_event(user, event)

      query = """
      query { myEvents { id name } }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query)

      data = json_response(conn, 200)["data"]["myEvents"]
      assert length(data) == 1
      assert hd(data)["id"] == event.id
    end

    test "filters by project", %{conn: conn} do
      user = create_user()
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})
      e1 = create_event(p1, %{name: "E1"})
      e2 = create_event(p2, %{name: "E2"})
      create_user_event(user, e1)
      create_user_event(user, e2)

      query = """
      query($project: ID) { myEvents(project: $project) { id name } }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"project" => p1.id})

      data = json_response(conn, 200)["data"]["myEvents"]
      assert length(data) == 1
      assert hd(data)["name"] == "E1"
    end
  end

  # ── Mutation: createEvent ──

  describe "createEvent mutation" do
    test "admin can create event", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      mutation($projectId: ID!, $input: CreateEventInput!) {
        createEvent(projectId: $projectId, input: $input) {
          id name description startDate endDate
        }
      }
      """

      variables = %{
        "projectId" => project.id,
        "input" => %{
          "name" => "E2E Test Event",
          "description" => "Description",
          "startDate" => "2026-07-01T00:00:00Z",
          "endDate" => "2026-07-31T23:59:59Z"
        }
      }

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, variables)

      data = json_response(conn, 200)["data"]["createEvent"]
      assert data["name"] == "E2E Test Event"
      assert String.starts_with?(data["id"], "EV")
    end

    test "user cannot create event", %{conn: conn} do
      project = create_project()
      user = create_user()

      query = """
      mutation($projectId: ID!, $input: CreateEventInput!) {
        createEvent(projectId: $projectId, input: $input) { id }
      }
      """

      variables = %{
        "projectId" => project.id,
        "input" => %{
          "name" => "Unauthorized",
          "description" => "Desc",
          "startDate" => "2026-07-01T00:00:00Z",
          "endDate" => "2026-07-31T23:59:59Z"
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

  # ── Mutation: updateEvent ──

  describe "updateEvent mutation" do
    test "admin can update event", %{conn: conn} do
      project = create_project()
      event = create_event(project)
      user = create_user()

      query = """
      mutation($id: ID!, $input: UpdateEventInput!) {
        updateEvent(id: $id, input: $input) { id name description }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{
          "id" => event.id,
          "input" => %{"name" => "Updated Event Name", "description" => "Updated desc"}
        })

      data = json_response(conn, 200)["data"]["updateEvent"]
      assert data["name"] == "Updated Event Name"
      assert data["description"] == "Updated desc"
    end
  end

  # ── Mutation: deleteEvent ──

  describe "deleteEvent mutation" do
    test "admin can delete event", %{conn: conn} do
      project = create_project()
      event = create_event(project)
      user = create_user()

      query = "mutation($id: ID!) { deleteEvent(id: $id) }"

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"id" => event.id})

      assert json_response(conn, 200)["data"]["deleteEvent"] == true
    end
  end

  # ── Mutation: moveEvent ──

  describe "moveEvent mutation" do
    test "admin can move event to different project", %{conn: conn} do
      p1 = create_project(%{name: "Source"})
      p2 = create_project(%{name: "Target"})
      event = create_event(p1)
      user = create_user()

      query = """
      mutation($id: ID!, $newProjectId: ID!) {
        moveEvent(id: $id, newProjectId: $newProjectId) {
          id parentProject { id name }
        }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{"id" => event.id, "newProjectId" => p2.id})

      data = json_response(conn, 200)["data"]["moveEvent"]
      assert data["parentProject"]["id"] == p2.id
    end
  end

  # ── Mutation: joinEvent ──

  describe "joinEvent mutation" do
    test "user can join event", %{conn: conn} do
      project = create_project()
      event = create_event(project, %{name: "Join Me"})
      user = create_user()

      query = """
      mutation($eventId: ID!) {
        joinEvent(eventId: $eventId) { id name }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"eventId" => event.id})

      data = json_response(conn, 200)["data"]["joinEvent"]
      assert data["id"] == event.id
      assert data["name"] == "Join Me"
    end

    test "unauthenticated join returns error", %{conn: conn} do
      project = create_project()
      event = create_event(project)

      query = """
      mutation($eventId: ID!) {
        joinEvent(eventId: $eventId) { id }
      }
      """

      conn = graphql_query(conn, query, %{"eventId" => event.id})
      assert json_response(conn, 200)["errors"] != nil
    end
  end
end
