defmodule ElixirBackendWeb.Schema.AchievementQueries do
  use Absinthe.Schema.Notation

  alias ElixirBackend.Achievements

  object :achievement_queries do
    field :achievement, non_null(:achievement) do
      arg(:id, non_null(:id))

      resolve(fn _, %{id: id}, _ ->
        Achievements.get_achievement(id)
      end)
    end

    field :achievements, non_null(:achievement_connection) do
      arg(:filter, non_null(:achievement_filter))
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn _, args, _ ->
        filter = args.filter
        pagination = Map.take(args, [:first, :after, :last, :before])
        Achievements.list_achievements(filter, pagination)
      end)
    end
  end
end
