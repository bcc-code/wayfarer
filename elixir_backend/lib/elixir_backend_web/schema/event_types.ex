defmodule ElixirBackendWeb.Schema.EventTypes do
  @moduledoc "Absinthe types for events: object, inputs, pagination."
  use Absinthe.Schema.Notation

  import Absinthe.Resolution.Helpers, only: [on_load: 2]

  alias ElixirBackend.Translations

  # ── Event object ──

  object :event do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:string)
    field :start_date, non_null(:datetime)
    field :end_date, non_null(:datetime)

    field :leaderboard, non_null(:leaderboard_connection) do
      arg(:entity_type, non_null(:leaderboard_entity_type))
      arg(:filter, :leaderboard_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn event, args, _ ->
        filter = args[:filter] || %{}
        pagination = Map.take(args, [:first, :after, :last, :before])

        ElixirBackend.Scoring.get_event_leaderboard(
          event.id,
          args.entity_type,
          filter,
          pagination
        )
      end)
    end

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn event, _, _ ->
        {:ok, Translations.translation_status(:event, event.id)}
      end)
    end

    field :parent_project, non_null(:project) do
      resolve(fn event, _, %{context: %{loader: loader}} = resolution ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :project, event)
        |> on_load(fn loader ->
          project = Dataloader.get(loader, ElixirBackend.Repo, :project, event)

          {:ok, project}
          |> Translations.translate_result(:project, resolution)
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
