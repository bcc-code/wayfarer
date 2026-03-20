defmodule ElixirBackendWeb.Schema.RoleQueries do
  @moduledoc "GraphQL query resolvers for roles."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Roles

  object :role_queries do
    field :user_roles, non_null(list_of(non_null(:user_role))) do
      arg(:user_id, non_null(:id))

      resolve(fn _parent, %{user_id: user_id}, _resolution ->
        Roles.list_user_roles(user_id)
      end)
    end

    field :users_with_role, non_null(list_of(non_null(:user))) do
      arg(:role, non_null(:role_type))
      arg(:scope_type, :scope_type)
      arg(:scope_id, :id)

      resolve(fn _parent, args, _resolution ->
        Roles.users_with_role(args.role,
          scope_type: args[:scope_type],
          scope_id: args[:scope_id]
        )
      end)
    end
  end
end
