defmodule ElixirBackend.Quizzes.QuizSubmission do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "quiz_submissions" do
    field :started_at, :utc_datetime
    field :completed_at, :utc_datetime
    field :expires_at, :utc_datetime
    field :auto_submitted, :boolean, default: false
    field :question_order, {:array, :string}
    field :score, :integer
    field :max_score, :integer
    field :points_awarded, :integer
    field :session_id, :string

    belongs_to :quiz, ElixirBackend.Quizzes.Quiz, type: :string
    belongs_to :user, ElixirBackend.Accounts.User, type: :string

    has_many :responses, ElixirBackend.Quizzes.QuizResponse, foreign_key: :submission_id
  end

  def changeset(submission, attrs) do
    submission
    |> cast(attrs, [
      :id,
      :started_at,
      :completed_at,
      :expires_at,
      :auto_submitted,
      :question_order,
      :score,
      :max_score,
      :points_awarded,
      :session_id,
      :quiz_id,
      :user_id
    ])
    |> validate_required([:id, :quiz_id, :user_id])
    |> foreign_key_constraint(:quiz_id)
    |> foreign_key_constraint(:user_id)
  end
end
