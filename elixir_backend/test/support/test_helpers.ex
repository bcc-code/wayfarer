defmodule ElixirBackend.TestHelpers do
  @moduledoc """
  Helper functions for creating test data.
  """

  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Projects.Project
  alias ElixirBackend.Accounts.User
  alias ElixirBackend.Accounts.UserProject
  alias ElixirBackend.Accounts.UserEvent
  alias ElixirBackend.Churches.Church
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

  def create_church(attrs \\ %{}) do
    defaults = %{
      id: ULID.new_church_id(),
      name: "Test Church",
      country: "NO",
      category: "L"
    }

    %Church{}
    |> Church.changeset(Map.merge(defaults, attrs))
    |> Repo.insert!()
  end

  def create_user(attrs \\ %{}) do
    # Auto-create church if not provided
    church_id =
      if attrs[:church_id] do
        attrs[:church_id]
      else
        create_church().id
      end

    unique = System.unique_integer([:positive])

    defaults = %{
      id: ULID.new_user_id(),
      name: "Test User",
      members_id: "member-#{unique}",
      email: "test-#{unique}@example.com",
      gender: "MALE",
      church_id: church_id,
      birthdate: ~D[2000-01-15]
    }

    %User{}
    |> User.create_changeset(Map.merge(defaults, attrs))
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

  def create_user_project(user, project) do
    %UserProject{}
    |> UserProject.changeset(%{
      user_id: user.id,
      project_id: project.id,
      joined_at: DateTime.utc_now() |> DateTime.truncate(:second)
    })
    |> Repo.insert!()
  end

  def create_user_event(user, event) do
    %UserEvent{}
    |> UserEvent.changeset(%{
      user_id: user.id,
      event_id: event.id,
      joined_at: DateTime.utc_now() |> DateTime.truncate(:second)
    })
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
