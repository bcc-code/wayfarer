defmodule ElixirBackendWeb.Schema.TeamQueries do
  @moduledoc "GraphQL query resolvers for teams and super teams."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Teams

  object :team_queries do
    field :team, non_null(:team) do
      arg(:id, non_null(:id))

      resolve(fn _parent, %{id: id}, _resolution ->
        Teams.get_team(id)
      end)
    end

    field :team_by_join_code, :team do
      arg(:code, non_null(:string))
      arg(:project_id, non_null(:id))

      resolve(fn _parent, %{code: code, project_id: project_id}, _resolution ->
        Teams.get_team_by_join_code(code, project_id)
      end)
    end

    field :teams, non_null(:team_connection) do
      arg(:filter, :team_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn _parent, args, _resolution ->
        filter = Map.get(args, :filter, %{}) || %{}

        pagination_opts =
          args
          |> Map.take([:first, :after, :last, :before])
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Teams.list_teams(filter, pagination_opts)
      end)
    end

    field :superteam, non_null(:super_team) do
      arg(:id, non_null(:id))

      resolve(fn _parent, %{id: id}, _resolution ->
        Teams.get_super_team(id)
      end)
    end

    field :superteams, non_null(:super_team_connection) do
      arg(:filter, :super_team_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn _parent, args, _resolution ->
        filter = Map.get(args, :filter, %{}) || %{}

        pagination_opts =
          args
          |> Map.take([:first, :after, :last, :before])
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Teams.list_super_teams(filter, pagination_opts)
      end)
    end
  end
end
