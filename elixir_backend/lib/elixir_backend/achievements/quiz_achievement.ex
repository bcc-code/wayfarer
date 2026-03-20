defmodule ElixirBackend.Achievements.QuizAchievement do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "quiz_achievements" do
    field :min_score_percentage, :integer
    field :require_completion, :boolean, default: true
    field :quiz_id, :string

    belongs_to :achievement, ElixirBackend.Achievements.Achievement,
      type: :string,
      foreign_key: :achievement_id,
      primary_key: true
  end

  def changeset(qa, attrs) do
    qa
    |> cast(attrs, [:achievement_id, :quiz_id, :min_score_percentage, :require_completion])
    |> validate_required([:achievement_id])
    |> foreign_key_constraint(:achievement_id)
  end
end
