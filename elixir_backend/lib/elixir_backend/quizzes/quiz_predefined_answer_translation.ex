defmodule ElixirBackend.Quizzes.QuizPredefinedAnswerTranslation do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "quiz_answer_translations" do
    field :language_code, :string, primary_key: true

    belongs_to :answer, ElixirBackend.Quizzes.QuizPredefinedAnswer,
      type: :string,
      primary_key: true

    field :answer_text, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(translation, attrs) do
    translation
    |> cast(attrs, [:answer_id, :language_code, :answer_text])
    |> validate_required([:answer_id, :language_code])
    |> foreign_key_constraint(:answer_id)
  end
end
