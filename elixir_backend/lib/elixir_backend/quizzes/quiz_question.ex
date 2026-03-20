defmodule ElixirBackend.Quizzes.QuizQuestion do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}
  @valid_types ~w(PREDEFINED FREE_TEXT NUMBER JSON ORDERING)

  schema "quiz_questions" do
    field :question_type, :string
    field :question_text, :string
    field :question_order, :integer
    field :allow_multiple_selection, :boolean, default: false
    field :min_value, :decimal
    field :max_value, :decimal
    field :step_value, :decimal
    field :points, :integer
    field :timeout_seconds, :integer
    field :betting_enabled, :boolean, default: false
    field :betting_min_percentage, :float
    field :betting_max_percentage, :float
    field :betting_min_absolute, :integer
    field :betting_max_absolute, :integer

    belongs_to :quiz, ElixirBackend.Quizzes.Quiz, type: :string

    has_many :predefined_answers, ElixirBackend.Quizzes.QuizPredefinedAnswer,
      foreign_key: :question_id

    has_many :ordering_items, ElixirBackend.Quizzes.QuizOrderingItem, foreign_key: :question_id

    timestamps(type: :utc_datetime)
  end

  @required_fields [:id, :question_type, :question_text, :question_order, :quiz_id]
  @optional_fields [
    :allow_multiple_selection,
    :min_value,
    :max_value,
    :step_value,
    :points,
    :timeout_seconds,
    :betting_enabled,
    :betting_min_percentage,
    :betting_max_percentage,
    :betting_min_absolute,
    :betting_max_absolute
  ]

  def changeset(question, attrs) do
    question
    |> cast(attrs, @required_fields ++ @optional_fields)
    |> validate_required(@required_fields)
    |> validate_inclusion(:question_type, @valid_types)
    |> foreign_key_constraint(:quiz_id)
  end

  def update_changeset(question, attrs) do
    question
    |> cast(attrs, @optional_fields ++ [:question_text, :question_order, :question_type])
    |> validate_inclusion(:question_type, @valid_types)
  end
end
