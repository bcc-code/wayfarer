defmodule ElixirBackendWeb.Schema.StreakTypes do
  use Absinthe.Schema.Notation
  @moduledoc false
  import Absinthe.Resolution.Helpers, only: [dataloader: 1]

  alias ElixirBackend.Streaks
  alias ElixirBackend.Translations

  object :streak do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:string)
    field :project, non_null(:project), resolve: dataloader(ElixirBackend.Repo)

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn streak, _, _ ->
        {:ok, Translations.translation_status(:streak, streak.id)}
      end)
    end

    field :relevant_days, non_null(list_of(non_null(:date_range))) do
      resolve(fn streak, _, _ ->
        days = Streaks.get_relevant_days(streak.id)
        {:ok, Enum.map(days, &%{start: &1.start_date, end: &1.end_date})}
      end)
    end

    field :status, non_null(:integer) do
      resolve(fn streak, _, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:ok, 0}
          user_id -> {:ok, Streaks.get_streak_status(streak.id, user_id)}
        end
      end)
    end

    field :listened_days, non_null(list_of(non_null(:streak_day))) do
      arg(:last, non_null(:integer))

      resolve(fn streak, %{last: last}, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:ok, []}
          user_id -> {:ok, Streaks.get_listened_days(streak.id, user_id, last)}
        end
      end)
    end
  end

  object :streak_day do
    field :date, non_null(:date)
    field :active, non_null(:boolean)
  end

  object :date_range do
    field :start, non_null(:date)
    field :end, non_null(:date)
  end

  input_object :date_range_input do
    field :start, non_null(:date)
    field :end, non_null(:date)
  end

  input_object :create_streak_input do
    field :name, non_null(:string)
    field :description, non_null(:string)
    field :project_id, non_null(:id)
    field :relevant_days, non_null(list_of(non_null(:date_range_input)))
  end

  input_object :update_streak_input do
    field :name, :string
    field :description, :string
    field :relevant_days, list_of(non_null(:date_range_input))
  end

  input_object :streak_filter do
    field :project_id, :id
    field :ids, list_of(non_null(:id))
  end

  # Pagination types
  object :streak_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:streak)
  end

  object :streak_connection do
    field :edges, non_null(list_of(non_null(:streak_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end
end
