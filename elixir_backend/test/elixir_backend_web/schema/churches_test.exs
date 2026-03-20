defmodule ElixirBackendWeb.Schema.ChurchesTest do
  use ElixirBackendWeb.ConnCase

  # ── Query: church(id) ──

  describe "church query" do
    test "returns church by id", %{conn: conn} do
      church = create_church(%{name: "Oslo Church", country: "NO", category: "L"})

      query = """
      query($id: ID!) {
        church(id: $id) { id name country category }
      }
      """

      conn = graphql_query(conn, query, %{"id" => church.id})
      data = json_response(conn, 200)["data"]["church"]

      assert data["id"] == church.id
      assert data["name"] == "Oslo Church"
      assert data["country"] == "NO"
      assert data["category"] == "L"
    end

    test "returns error for nonexistent church", %{conn: conn} do
      query = "query($id: ID!) { church(id: $id) { id } }"
      conn = graphql_query(conn, query, %{"id" => "CH00000000000000000000000000"})
      assert json_response(conn, 200)["errors"] != nil
    end
  end

  # ── Query: churches(filter, pagination) ──

  describe "churches query" do
    test "returns paginated churches", %{conn: conn} do
      create_church(%{name: "Church A"})
      create_church(%{name: "Church B"})
      create_church(%{name: "Church C"})

      query = """
      query($first: Int) {
        churches(first: $first) {
          edges { cursor node { id name } }
          pageInfo { hasNextPage hasPreviousPage }
          totalCount
        }
      }
      """

      conn = graphql_query(conn, query, %{"first" => 2})
      data = json_response(conn, 200)["data"]["churches"]

      assert length(data["edges"]) == 2
      assert data["totalCount"] >= 3
      assert data["pageInfo"]["hasNextPage"] == true
    end

    test "filters by country", %{conn: conn} do
      create_church(%{name: "Norwegian", country: "NO"})
      create_church(%{name: "Swedish", country: "SE"})

      query = """
      query($filter: ChurchFilter) {
        churches(filter: $filter, first: 10) {
          edges { node { id country } }
          totalCount
        }
      }
      """

      conn = graphql_query(conn, query, %{"filter" => %{"country" => "NO"}})
      data = json_response(conn, 200)["data"]["churches"]

      assert Enum.all?(data["edges"], fn e -> e["node"]["country"] == "NO" end)
    end

    test "filters by category", %{conn: conn} do
      create_church(%{name: "Small", category: "S"})
      create_church(%{name: "Large", category: "L"})

      query = """
      query($filter: ChurchFilter) {
        churches(filter: $filter, first: 10) {
          edges { node { id category } }
        }
      }
      """

      conn = graphql_query(conn, query, %{"filter" => %{"category" => "S"}})
      data = json_response(conn, 200)["data"]["churches"]

      assert Enum.all?(data["edges"], fn e -> e["node"]["category"] == "S" end)
    end
  end

  # ── Mutation: updateChurch ──

  describe "updateChurch mutation" do
    test "superadmin can update church", %{conn: conn} do
      church = create_church(%{name: "Old Name"})
      user = create_user()

      query = """
      mutation($id: ID!, $input: UpdateChurchInput!) {
        updateChurch(id: $id, input: $input) { id name }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["superadmin"])
        |> graphql_query(query, %{
          "id" => church.id,
          "input" => %{"name" => "New Name"}
        })

      data = json_response(conn, 200)["data"]["updateChurch"]
      assert data["name"] == "New Name"
    end

    test "admin cannot update church", %{conn: conn} do
      church = create_church()
      user = create_user()

      query = """
      mutation($id: ID!, $input: UpdateChurchInput!) {
        updateChurch(id: $id, input: $input) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["admin"])
        |> graphql_query(query, %{
          "id" => church.id,
          "input" => %{"name" => "New"}
        })

      assert json_response(conn, 200)["errors"] != nil
    end

    test "user cannot update church", %{conn: conn} do
      church = create_church()
      user = create_user()

      query = """
      mutation($id: ID!, $input: UpdateChurchInput!) {
        updateChurch(id: $id, input: $input) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{
          "id" => church.id,
          "input" => %{"name" => "New"}
        })

      assert json_response(conn, 200)["errors"] != nil
    end
  end
end
