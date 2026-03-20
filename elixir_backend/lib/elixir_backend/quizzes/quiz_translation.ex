defmodule ElixirBackend.Quizzes.QuizTranslation do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "quiz_translations" do
    field :language_code, :string, primary_key: true

    belongs_to :quiz, ElixirBackend.Quizzes.Quiz,
      type: :string,
      primary_key: true

    field :name, :string
    field :description, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(translation, attrs) do
    translation
    |> cast(attrs, [:quiz_id, :language_code, :name, :description])
    |> validate_required([:quiz_id, :language_code])
    |> foreign_key_constraint(:quiz_id)
  end
end
