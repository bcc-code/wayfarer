defmodule ElixirBackendWeb.Schema.UserMutations do
  @moduledoc "GraphQL mutation resolvers for users."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Accounts
  alias ElixirBackendWeb.Schema.Middleware.RequireRole

  object :user_mutations do
    field :assign_user_to_project, non_null(:user) do
      arg(:user_id, non_null(:id))
      arg(:project_id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin", "m2m"])

      resolve(fn _parent, %{user_id: user_id, project_id: project_id}, _resolution ->
        Accounts.assign_user_to_project(user_id, project_id)
      end)
    end

    field :remove_user_from_project, non_null(:user) do
      arg(:user_id, non_null(:id))
      arg(:project_id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin", "m2m"])

      resolve(fn _parent, %{user_id: user_id, project_id: project_id}, _resolution ->
        Accounts.remove_user_from_project(user_id, project_id)
      end)
    end

    field :assign_user_to_event, non_null(:user) do
      arg(:user_id, non_null(:id))
      arg(:event_id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin", "m2m"])

      resolve(fn _parent, %{user_id: user_id, event_id: event_id}, _resolution ->
        Accounts.assign_user_to_event(user_id, event_id)
      end)
    end

    field :lock_user_church, non_null(:user) do
      arg(:user_id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{user_id: user_id}, _resolution ->
        Accounts.lock_user_church(user_id)
      end)
    end

    field :unlock_user_church, non_null(:user) do
      arg(:user_id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{user_id: user_id}, _resolution ->
        Accounts.unlock_user_church(user_id)
      end)
    end
  end
end
