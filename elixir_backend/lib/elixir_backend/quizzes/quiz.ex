defmodule ElixirBackend.Quizzes.Quiz do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "quizzes" do
    field :name, :string
    field :description, :string
    field :image_url, :string
    field :timeout_seconds, :integer
    field :randomize_questions, :boolean, default: false
    field :reveal_correct_answers, :boolean, default: true
    field :allow_retakes, :boolean, default: false
    field :completion_points, :integer, default: 0
    field :end_time, :utc_datetime

    belongs_to :project, ElixirBackend.Projects.Project, type: :string
    field :challenge_id, :string

    has_many :questions, ElixirBackend.Quizzes.QuizQuestion

    timestamps(type: :utc_datetime)
  end

  @required_fields [:id, :name, :project_id]
  @optional_fields [
    :description,
    :image_url,
    :timeout_seconds,
    :randomize_questions,
    :reveal_correct_answers,
    :allow_retakes,
    :completion_points,
    :end_time,
    :challenge_id
  ]

  def changeset(quiz, attrs) do
    quiz
    |> cast(attrs, @required_fields ++ @optional_fields)
    |> validate_required(@required_fields)
    |> foreign_key_constraint(:project_id)
  end

  def update_changeset(quiz, attrs) do
    quiz
    |> cast(attrs, @optional_fields ++ [:name])
  end
end
