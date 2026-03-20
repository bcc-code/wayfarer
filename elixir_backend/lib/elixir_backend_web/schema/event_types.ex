defmodule ElixirBackendWeb.Schema.EventTypes do
  @moduledoc "Absinthe types for events: object, inputs, pagination."
  use Absinthe.Schema.Notation

  import Absinthe.Resolution.Helpers, only: [on_load: 2]

  # ── Event object ──

  object :event do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:string)
    field :start_date, non_null(:datetime)
    field :end_date, non_null(:datetime)

    field :parent_project, non_null(:project) do
      resolve(fn event, _, %{context: %{loader: loader}} ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :project, event)
        |> on_load(fn loader ->
          {:ok, Dataloader.get(loader, ElixirBackend.Repo, :project, event)}
        end)
      end)
    end
  end

  # ── Pagination ──

  object :event_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:event)
  end

  object :event_connection do
    field :edges, non_null(list_of(non_null(:event_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end

  # ── Input types ──

  input_object :create_event_input do
    field :name, non_null(:string)
    field :description, non_null(:string)
    field :start_date, non_null(:datetime)
    field :end_date, non_null(:datetime)
  end

  input_object :update_event_input do
    field :name, :string
    field :description, :string
    field :start_date, :datetime
    field :end_date, :datetime
  end

  input_object :event_filter do
    field :project_id, :id
    field :ids, list_of(non_null(:id))
    field :start_date_after, :datetime
    field :start_date_before, :datetime
    field :end_date_after, :datetime
    field :end_date_before, :datetime
  end
end
