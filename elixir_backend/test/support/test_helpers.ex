defmodule ElixirBackend.TestHelpers do
  @moduledoc """
  Helper functions for creating test data.
  """

  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Projects.Project
  alias ElixirBackend.Accounts.User
  alias ElixirBackend.Events.Event

  def create_project(attrs \\ %{}) do
    defaults = %{
      id: ULID.new_project_id(),
      name: "Test Project",
      start_date: ~U[2026-01-01 00:00:00Z],
      end_date: ~U[2026-12-31 23:59:59Z]
    }

    %Project{}
    |> Project.changeset(Map.merge(defaults, attrs))
    |> Repo.insert!()
  end

  def create_user(attrs \\ %{}) do
    defaults = %{
      id: ULID.new_user_id(),
      name: "Test User"
    }

    %User{}
    |> User.changeset(Map.merge(defaults, attrs))
    |> Repo.insert!()
  end

  def create_event(project, attrs \\ %{}) do
    defaults = %{
      id: ULID.new_event_id(),
      name: "Test Event",
      project_id: project.id
    }

    %Event{}
    |> Event.changeset(Map.merge(defaults, attrs))
    |> Repo.insert!()
  end

  def build_auth_token(user_id, roles \\ ["user"]) do
    secret = Application.get_env(:elixir_backend, :jwt_secret, "dev-secret")
    signer = Joken.Signer.create("HS256", secret)

    claims = %{
      "sub" => user_id,
      "roles" => roles,
      "exp" => DateTime.utc_now() |> DateTime.add(3600) |> DateTime.to_unix()
    }

    {:ok, token, _claims} = Joken.encode_and_sign(claims, signer)
    token
  end

  def auth_conn(conn, user_id, roles \\ ["user"]) do
    token = build_auth_token(user_id, roles)
    Plug.Conn.put_req_header(conn, "authorization", "Bearer #{token}")
  end

  def graphql_query(conn, query, variables \\ %{}) do
    conn
    |> Plug.Conn.put_req_header("content-type", "application/json")
    |> Phoenix.ConnTest.dispatch(
      ElixirBackendWeb.Endpoint,
      :post,
      "/api/graphql",
      Jason.encode!(%{query: query, variables: variables})
    )
  end
end
