defmodule ElixirBackendWeb.Schema.ExternalContentTest do
  use ElixirBackendWeb.ConnCase

  alias ElixirBackend.ExternalContent, as: EC

  defp create_content_fixture(attrs \\ %{}) do
    defaults = %{
      plan_id: "plan-#{System.unique_integer([:positive])}",
      task_id: "task-#{System.unique_integer([:positive])}",
      content_type: "ARTICLE",
      source: "ssf"
    }

    {:ok, content} = EC.upsert_content(Map.merge(defaults, attrs))
    content
  end

  # ── Query: externalContent ──

  describe "externalContent query" do
    test "admin can get external content by id", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      content = create_content_fixture()

      EC.upsert_translation(%{
        external_content_id: content.id,
        language_code: "en",
        title: "Test Article"
      })

      query = """
      query($id: ID!) {
        externalContent(id: $id) {
          id planId taskId contentType source
          translations { languageCode title }
          title
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{"id" => content.id})

      data = json_response(conn, 200)["data"]["externalContent"]
      assert data["id"] == content.id
      assert data["contentType"] == "ARTICLE"
      assert length(data["translations"]) == 1
      assert hd(data["translations"])["title"] == "Test Article"
    end

    test "user cannot access external content", %{conn: conn} do
      user = create_user()
      content = create_content_fixture()

      query = """
      query($id: ID!) {
        externalContent(id: $id) { id }
      }
      """

      conn =
        conn
        |> auth_conn(user.id, ["user"])
        |> graphql_query(query, %{"id" => content.id})

      json = json_response(conn, 200)
      assert json["errors"] != nil
      assert hd(json["errors"])["message"] =~ "unauthorized"
    end
  end

  # ── Query: externalContents ──

  describe "externalContents query" do
    test "admin can list and filter external contents", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      create_content_fixture(%{plan_id: "my-plan", task_id: "t1", content_type: "MEDIA"})
      create_content_fixture(%{plan_id: "my-plan", task_id: "t2", content_type: "ARTICLE"})
      create_content_fixture(%{plan_id: "other-plan", task_id: "t3"})

      query = """
      query($filter: ExternalContentFilter!) {
        externalContents(filter: $filter, first: 10) {
          totalCount
          edges { node { id planId contentType } }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{"filter" => %{"planId" => "my-plan"}})

      data = json_response(conn, 200)["data"]["externalContents"]
      assert data["totalCount"] == 2
      assert length(data["edges"]) == 2
    end

    test "admin can filter by content type", %{conn: conn} do
      admin = create_user(%{name: "Admin"})
      create_content_fixture(%{content_type: "SONG"})
      create_content_fixture(%{content_type: "MEDIA"})

      query = """
      query($filter: ExternalContentFilter!) {
        externalContents(filter: $filter, first: 10) {
          totalCount
          edges { node { contentType } }
        }
      }
      """

      conn =
        conn
        |> auth_conn(admin.id, ["admin"])
        |> graphql_query(query, %{"filter" => %{"contentType" => "SONG"}})

      data = json_response(conn, 200)["data"]["externalContents"]
      assert data["totalCount"] == 1
      assert hd(data["edges"])["node"]["contentType"] == "SONG"
    end
  end
end
