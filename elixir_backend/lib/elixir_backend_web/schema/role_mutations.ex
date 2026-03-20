defmodule ElixirBackendWeb.Schema.RoleMutations do
  @moduledoc "GraphQL mutation resolvers for roles."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Roles
  alias ElixirBackendWeb.Schema.Middleware.RequireRole

  object :role_mutations do
    field :assign_role, non_null(:user_role) do
      arg(:input, non_null(:assign_role_input))

      middleware(RequireRole, roles: ["admin", "superadmin", "church_admin"])

      resolve(fn _parent, %{input: input}, %{context: context} ->
        attrs = Map.put(input, :assigned_by, context[:current_user_id])
        Roles.assign_role(attrs)
      end)
    end

    field :revoke_role, non_null(:boolean) do
      arg(:input, non_null(:revoke_role_input))

      middleware(RequireRole, roles: ["admin", "superadmin", "church_admin"])

      resolve(fn _parent, %{input: input}, _resolution ->
        Roles.revoke_role(input)
      end)
    end
  end
end
