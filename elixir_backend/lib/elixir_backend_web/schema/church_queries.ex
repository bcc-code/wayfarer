defmodule ElixirBackendWeb.Schema.ChurchQueries do
  @moduledoc "GraphQL query resolvers for churches."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Churches

  object :church_queries do
    field :church, non_null(:church) do
      arg(:id, non_null(:id))

      resolve(fn _parent, %{id: id}, _resolution ->
        Churches.get_church(id)
      end)
    end

    field :churches, non_null(:church_connection) do
      arg(:filter, :church_filter)
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

        Churches.list_churches(filter, pagination_opts)
      end)
    end
  end
end
