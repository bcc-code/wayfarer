defmodule ElixirBackendWeb.Schema.StreakQueries do
  use Absinthe.Schema.Notation

  alias ElixirBackend.Streaks

  object :streak_queries do
    field :streak, non_null(:streak) do
      arg(:id, non_null(:id))

      resolve(fn _, %{id: id}, _ ->
        Streaks.get_streak(id)
      end)
    end

    field :streaks, non_null(:streak_connection) do
      arg(:filter, :streak_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn _, args, _ ->
        filter = args[:filter] || %{}
        pagination = Map.take(args, [:first, :after, :last, :before])
        Streaks.list_streaks(filter, pagination)
      end)
    end
  end
end
