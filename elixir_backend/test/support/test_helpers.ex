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
      description: "A test project",
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
      description: "A test event",
      start_date: ~U[2026-06-01 00:00:00Z],
      end_date: ~U[2026-06-30 23:59:59Z],
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

  def create_streak(project, attrs \\ %{}) do
    default_attrs = %{
      name: "Test Streak",
      description: "A test streak",
      project_id: project.id,
      relevant_days: [%{start: ~D[2026-03-01], end: ~D[2026-03-31]}]
    }

    {:ok, streak} = ElixirBackend.Streaks.create_streak(Map.merge(default_attrs, attrs))
    streak
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

  @doc "Default branding input for project creation tests."
  def default_branding_input do
    %{
      "logo" => nil,
      "banner" => nil,
      "rounding" => 8,
      "colors" => %{
        "light" => default_color_set_input(),
        "dark" => default_color_set_input()
      }
    }
  end

  defp default_color_set_input do
    %{
      "accent" => "#000000",
      "accentContrast" => "#FFFFFF",
      "onAccent" => "#FFFFFF",
      "backgroundDefault" => "#FFFFFF",
      "backgroundRaised" => "#F5F5F5",
      "backgroundIndent" => "#E0E0E0",
      "textDefault" => "#000000",
      "textMuted" => "#666666",
      "textHint" => "#999999",
      "shadowDefault" => "#00000020",
      "shadowBlank" => "#00000000",
      "borderDefault" => "#E0E0E0"
    }
  end
end
