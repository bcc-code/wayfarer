defmodule ElixirBackendWeb.Schema.UserQueries do
  @moduledoc "GraphQL query resolvers for users."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Accounts

  object :user_queries do
    field :me, :user do
      resolve(fn _parent, _args, %{context: context} ->
        case context do
          %{current_user_id: user_id} ->
            Accounts.me(user_id)

          _ ->
            {:error, "authentication required"}
        end
      end)
    end

    field :user, :user do
      arg(:id, non_null(:id))

      resolve(fn _parent, %{id: id}, %{context: context} ->
        Accounts.get_accessible_user(id,
          user_id: context[:current_user_id],
          roles: context[:roles] || []
        )
      end)
    end

    field :users, non_null(:user_connection) do
      arg(:filter, :user_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn _parent, args, %{context: context} ->
        filter = Map.get(args, :filter, %{}) || %{}

        pagination_opts =
          args
          |> Map.take([:first, :after, :last, :before])
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Accounts.list_users(filter, pagination_opts,
          user_id: context[:current_user_id],
          roles: context[:roles] || []
        )
      end)
    end
  end
end
