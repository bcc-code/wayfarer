defmodule ElixirBackend.Quizzes.QuizPredefinedAnswer do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "quiz_predefined_answers" do
    field :answer_text, :string
    field :is_correct, :boolean, default: false
    field :answer_order, :integer

    belongs_to :question, ElixirBackend.Quizzes.QuizQuestion, type: :string

    timestamps(type: :utc_datetime)
  end

  def changeset(answer, attrs) do
    answer
    |> cast(attrs, [:id, :answer_text, :is_correct, :answer_order, :question_id])
    |> validate_required([:id, :answer_text, :question_id])
    |> foreign_key_constraint(:question_id)
  end
end
