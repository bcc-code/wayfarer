defmodule ElixirBackendWeb.Schema.AchievementQueries do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Achievements
  alias ElixirBackend.Translations

  object :achievement_queries do
    field :achievement, non_null(:achievement) do
      arg(:id, non_null(:id))

      resolve(fn _, %{id: id}, resolution ->
        Achievements.get_achievement(id)
        |> Translations.translate_result(:achievement, resolution)
      end)
    end

    field :achievements, non_null(:achievement_connection) do
      arg(:filter, non_null(:achievement_filter))
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn _, args, resolution ->
        filter = args.filter
        pagination = Map.take(args, [:first, :after, :last, :before])

        Achievements.list_achievements(filter, pagination)
        |> Translations.translate_connection(:achievement, resolution)
      end)
    end
  end
end
