defmodule ElixirBackend.Achievements.StreakAchievement do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "streak_achievements" do
    field :needed_streak, :integer

    belongs_to :achievement, ElixirBackend.Achievements.Achievement,
      type: :string,
      foreign_key: :achievement_id,
      primary_key: true

    belongs_to :streak, ElixirBackend.Streaks.Streak, type: :string
  end

  def changeset(sa, attrs) do
    sa
    |> cast(attrs, [:achievement_id, :streak_id, :needed_streak])
    |> validate_required([:achievement_id, :streak_id, :needed_streak])
    |> validate_number(:needed_streak, greater_than: 0)
    |> foreign_key_constraint(:achievement_id)
    |> foreign_key_constraint(:streak_id)
  end
end
