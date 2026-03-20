defmodule ElixirBackendWeb.Schema.TeamTypes do
  @moduledoc "Absinthe types for teams and super teams."
  use Absinthe.Schema.Notation

  import Absinthe.Resolution.Helpers, only: [dataloader: 1, on_load: 2]

  alias ElixirBackend.Teams

  # ── Team object ──

  object :team do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, :string
    field :join_code, non_null(:string)
    field :leaderboard_excluded, non_null(:boolean)

    field :members, non_null(list_of(non_null(:team_member))) do
      resolve(fn team, _, _ ->
        {:ok, Teams.get_team_members(team.id)}
      end)
    end

    field :parent_project, non_null(:project) do
      resolve(fn team, _, %{context: %{loader: loader}} ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :project, team)
        |> on_load(fn loader ->
          {:ok, Dataloader.get(loader, ElixirBackend.Repo, :project, team)}
        end)
      end)
    end

    field :super_team, :super_team, resolve: dataloader(ElixirBackend.Repo)
  end

  object :team_member do
    field :id, non_null(:id) do
      resolve(fn member, _, _ -> {:ok, member.user_id} end)
    end

    field :name, non_null(:string) do
      resolve(fn member, _, _ ->
        {:ok, member.user.name}
      end)
    end

    field :church, non_null(:church) do
      resolve(fn member, _, %{context: %{loader: loader}} ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :church, member.user)
        |> on_load(fn loader ->
          {:ok, Dataloader.get(loader, ElixirBackend.Repo, :church, member.user)}
        end)
      end)
    end

    field :is_team_lead, non_null(:boolean)

    field :joined_at, non_null(:string) do
      resolve(fn member, _, _ ->
        {:ok, DateTime.to_iso8601(member.joined_at)}
      end)
    end

    field :user, non_null(:user) do
      resolve(fn member, _, _ -> {:ok, member.user} end)
    end
  end

  # ── SuperTeam object ──

  object :super_team do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, :string
    field :color, :string
    field :image_url, :string

    field :teams, non_null(list_of(non_null(:team))), resolve: dataloader(ElixirBackend.Repo)

    field :parent_project, non_null(:project) do
      resolve(fn super_team, _, %{context: %{loader: loader}} ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :project, super_team)
        |> on_load(fn loader ->
          {:ok, Dataloader.get(loader, ElixirBackend.Repo, :project, super_team)}
        end)
      end)
    end
  end

  # ── Pagination ──

  object :team_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:team)
  end

  object :team_connection do
    field :edges, non_null(list_of(non_null(:team_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end

  object :super_team_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:super_team)
  end

  object :super_team_connection do
    field :edges, non_null(list_of(non_null(:super_team_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end

  # ── Input types ──

  input_object :create_team_input do
    field :name, non_null(:string)
    field :description, :string
  end

  input_object :update_team_input do
    field :name, :string
    field :description, :string
    field :leaderboard_excluded, :boolean
  end

  input_object :create_super_team_input do
    field :name, non_null(:string)
    field :description, :string
    field :image_url, :string
    field :color, :string
    field :team_ids, list_of(non_null(:id))
  end

  input_object :update_super_team_input do
    field :name, :string
    field :description, :string
    field :image_url, :string
    field :color, :string
  end

  input_object :team_filter do
    field :church_id, :id
    field :project_id, :id
    field :super_team_id, :id
    field :ids, list_of(non_null(:id))
    field :no_super_team, :boolean
  end

  input_object :super_team_filter do
    field :project_id, :id
    field :ids, list_of(non_null(:id))
  end
end
