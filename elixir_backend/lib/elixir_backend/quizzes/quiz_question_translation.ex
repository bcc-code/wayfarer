defmodule ElixirBackend.Quizzes.QuizQuestionTranslation do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "quiz_question_translations" do
    field :language_code, :string, primary_key: true

    belongs_to :question, ElixirBackend.Quizzes.QuizQuestion,
      type: :string,
      primary_key: true

    field :question_text, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(translation, attrs) do
    translation
    |> cast(attrs, [:question_id, :language_code, :question_text])
    |> validate_required([:question_id, :language_code])
    |> foreign_key_constraint(:question_id)
  end
end
