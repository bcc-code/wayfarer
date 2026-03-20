defmodule ElixirBackendWeb.Schema.ProjectQueries do
  @moduledoc "GraphQL query resolvers for projects."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Projects
  alias ElixirBackend.Translations

  object :project_queries do
    field :project, non_null(:project) do
      arg(:id, non_null(:id))

      resolve(fn _parent, %{id: id}, resolution ->
        Projects.get_project(id)
        |> Translations.translate_result(:project, resolution)
      end)
    end

    field :projects, non_null(:project_connection) do
      arg(:filter, :project_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn _parent, args, resolution ->
        filter = Map.get(args, :filter, %{}) || %{}

        pagination_opts =
          args
          |> Map.take([:first, :after, :last, :before])
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Projects.list_projects(filter, pagination_opts)
        |> Translations.translate_connection(:project, resolution)
      end)
    end

    field :my_projects, non_null(list_of(non_null(:project))) do
      resolve(fn _parent, _args, %{context: context} = resolution ->
        case context do
          %{current_user_id: user_id} ->
            Projects.my_projects(user_id)
            |> Translations.translate_list(:project, resolution)

          _ ->
            {:error, "authentication required"}
        end
      end)
    end
  end
end
