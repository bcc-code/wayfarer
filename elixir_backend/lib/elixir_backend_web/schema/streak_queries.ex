defmodule ElixirBackendWeb.Schema.StreakQueries do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Streaks
  alias ElixirBackend.Translations

  object :streak_queries do
    field :streak, non_null(:streak) do
      arg(:id, non_null(:id))

      resolve(fn _, %{id: id}, resolution ->
        Streaks.get_streak(id)
        |> Translations.translate_result(:streak, resolution)
      end)
    end

    field :streaks, non_null(:streak_connection) do
      arg(:filter, :streak_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn _, args, resolution ->
        filter = args[:filter] || %{}
        pagination = Map.take(args, [:first, :after, :last, :before])

        Streaks.list_streaks(filter, pagination)
        |> Translations.translate_connection(:streak, resolution)
      end)
    end
  end
end
