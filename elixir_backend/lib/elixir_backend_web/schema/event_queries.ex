defmodule ElixirBackendWeb.Schema.EventQueries do
  @moduledoc "GraphQL query resolvers for events."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Events

  object :event_queries do
    field :event, non_null(:event) do
      arg(:id, non_null(:id))

      resolve(fn _parent, %{id: id}, _resolution ->
        Events.get_event(id)
      end)
    end

    field :events, non_null(:event_connection) do
      arg(:filter, :event_filter)
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

        Events.list_events(filter, pagination_opts)
      end)
    end

    field :my_events, non_null(list_of(non_null(:event))) do
      arg(:project, :id)

      resolve(fn _parent, args, %{context: context} ->
        case context do
          %{current_user_id: user_id} ->
            Events.my_events(user_id, args[:project])

          _ ->
            {:error, "authentication required"}
        end
      end)
    end
  end
end
