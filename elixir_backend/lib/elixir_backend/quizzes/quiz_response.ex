defmodule ElixirBackend.Quizzes.QuizResponse do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "quiz_responses" do
    field :selected_answer_ids, {:array, :string}
    field :text_response, :string
    field :number_response, :decimal
    field :json_response, :map
    field :submitted_order, {:array, :string}
    field :is_correct, :boolean
    field :points_earned, :integer
    field :answered_at, :utc_datetime
    field :time_spent_seconds, :integer
    field :bet_amount, :integer

    belongs_to :submission, ElixirBackend.Quizzes.QuizSubmission, type: :string
    belongs_to :question, ElixirBackend.Quizzes.QuizQuestion, type: :string
  end

  def changeset(response, attrs) do
    response
    |> cast(attrs, [
      :id,
      :selected_answer_ids,
      :text_response,
      :number_response,
      :json_response,
      :submitted_order,
      :is_correct,
      :points_earned,
      :answered_at,
      :time_spent_seconds,
      :bet_amount,
      :submission_id,
      :question_id
    ])
    |> validate_required([:id, :submission_id, :question_id])
    |> foreign_key_constraint(:submission_id)
    |> foreign_key_constraint(:question_id)
    |> unique_constraint([:submission_id, :question_id])
  end
end
