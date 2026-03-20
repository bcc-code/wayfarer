defmodule ElixirBackend.Achievements.ContentAchievement do
  use Ecto.Schema

  @primary_key false

  schema "content_achievements" do
    belongs_to :achievement, ElixirBackend.Achievements.Achievement,
      type: :string,
      foreign_key: :achievement_id,
      primary_key: true

    has_many :items, ElixirBackend.Achievements.ContentAchievementItem,
      foreign_key: :achievement_id,
      references: :achievement_id
  end
end
