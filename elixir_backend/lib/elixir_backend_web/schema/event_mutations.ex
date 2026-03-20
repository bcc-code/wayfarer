defmodule ElixirBackendWeb.Schema.EventMutations do
  @moduledoc "GraphQL mutation resolvers for events."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Events
  alias ElixirBackendWeb.Schema.Middleware.RequireRole

  object :event_mutations do
    field :create_event, non_null(:event) do
      arg(:project_id, non_null(:id))
      arg(:input, non_null(:create_event_input))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{project_id: project_id, input: input}, _resolution ->
        Events.create_event(project_id, input)
      end)
    end

    field :update_event, non_null(:event) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_event_input))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id, input: input}, _resolution ->
        Events.update_event(id, input)
      end)
    end

    field :delete_event, non_null(:boolean) do
      arg(:id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id}, _resolution ->
        case Events.delete_event(id) do
          {:ok, _} -> {:ok, true}
          {:error, _} -> {:error, "failed to delete event"}
        end
      end)
    end

    field :move_event, non_null(:event) do
      arg(:id, non_null(:id))
      arg(:new_project_id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id, new_project_id: new_project_id}, _resolution ->
        Events.move_event(id, new_project_id)
      end)
    end

    field :join_event, non_null(:event) do
      arg(:event_id, non_null(:id))

      resolve(fn _parent, %{event_id: event_id}, %{context: context} ->
        case context do
          %{current_user_id: user_id} ->
            Events.join_event(user_id, event_id)

          _ ->
            {:error, "authentication required"}
        end
      end)
    end
  end
end
