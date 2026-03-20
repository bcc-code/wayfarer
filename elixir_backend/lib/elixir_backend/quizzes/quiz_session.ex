defmodule ElixirBackend.Quizzes.QuizSession do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}
  @valid_states ~w(DRAFT OPEN LOCKED FINISHED)

  schema "quiz_sessions" do
    field :name, :string
    field :state, :string, default: "DRAFT"
    field :open_at, :utc_datetime
    field :lock_at, :utc_datetime
    field :finish_at, :utc_datetime
    field :created_by, :string

    belongs_to :quiz, ElixirBackend.Quizzes.Quiz, type: :string

    timestamps(type: :utc_datetime)
  end

  @required_fields [:id, :name, :quiz_id]
  @optional_fields [:state, :open_at, :lock_at, :finish_at, :created_by]

  def changeset(session, attrs) do
    session
    |> cast(attrs, @required_fields ++ @optional_fields)
    |> validate_required(@required_fields)
    |> validate_inclusion(:state, @valid_states)
    |> foreign_key_constraint(:quiz_id)
  end

  def update_changeset(session, attrs) do
    session
    |> cast(attrs, @optional_fields ++ [:name])
    |> validate_inclusion(:state, @valid_states)
  end
end
