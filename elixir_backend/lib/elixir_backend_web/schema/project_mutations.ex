defmodule ElixirBackendWeb.Schema.ProjectMutations do
  @moduledoc "GraphQL mutation resolvers for projects."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Projects
  alias ElixirBackendWeb.Schema.Middleware.RequireRole

  object :project_mutations do
    field :create_project, non_null(:project) do
      arg(:input, non_null(:create_project_input))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{input: input}, _resolution ->
        attrs = flatten_branding(input)
        Projects.create_project(attrs)
      end)
    end

    field :update_project, non_null(:project) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_project_input))

      middleware(RequireRole, roles: ["admin", "superadmin", "project_admin"])

      resolve(fn _parent, %{id: id, input: input}, _resolution ->
        attrs = flatten_branding(input)
        Projects.update_project(id, attrs)
      end)
    end

    field :delete_project, non_null(:boolean) do
      arg(:id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id}, _resolution ->
        case Projects.delete_project(id) do
          {:ok, _} -> {:ok, true}
          {:error, _} -> {:error, "failed to delete project"}
        end
      end)
    end

    field :archive_project, non_null(:boolean) do
      arg(:id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id}, _resolution ->
        case Projects.archive_project(id) do
          {:ok, _} -> {:ok, true}
          {:error, _} -> {:error, "failed to archive project"}
        end
      end)
    end

    field :join_project, non_null(:project) do
      arg(:project_id, non_null(:id))

      resolve(fn _parent, %{project_id: project_id}, %{context: context} ->
        case context do
          %{current_user_id: user_id} ->
            Projects.join_project(user_id, project_id)

          _ ->
            {:error, "authentication required"}
        end
      end)
    end
  end

  defp flatten_branding(input) do
    case Map.pop(input, :branding) do
      {nil, attrs} ->
        attrs

      {branding, attrs} ->
        attrs
        |> Map.put(:logo_url, branding[:logo])
        |> Map.put(:banner_url, branding[:banner])
        |> Map.put(:rounding, branding[:rounding])
        |> flatten_color_set(branding[:colors][:light], "light")
        |> flatten_color_set(branding[:colors][:dark], "dark")
    end
  end

  defp flatten_color_set(attrs, nil, _prefix), do: attrs

  defp flatten_color_set(attrs, color_set, prefix) do
    Enum.reduce(color_set, attrs, fn {key, value}, acc ->
      Map.put(acc, :"color_#{prefix}_#{key}", value)
    end)
  end
end
