defmodule ElixirBackend.Scoring.ScoreJournal do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}
  @timestamps_opts [inserted_at: :created_at, updated_at: false]

  schema "score_journal" do
    field :points, :integer
    field :source_type, :string
    field :source_id, :string
    field :reason, :string
    field :awarded_by, :string
    field :challenge_id, :string

    belongs_to :project, ElixirBackend.Projects.Project, type: :string
    belongs_to :user, ElixirBackend.Accounts.User, type: :string
    belongs_to :event, ElixirBackend.Events.Event, type: :string

    field :created_at, :utc_datetime
  end

  @valid_source_types ~w(ACHIEVEMENT CHALLENGE EVENT MANUAL BET)

  def changeset(entry, attrs) do
    entry
    |> cast(attrs, [
      :id,
      :project_id,
      :user_id,
      :event_id,
      :challenge_id,
      :points,
      :source_type,
      :source_id,
      :reason,
      :awarded_by,
      :created_at
    ])
    |> validate_required([:id, :project_id, :user_id, :points, :source_type])
    |> validate_inclusion(:source_type, @valid_source_types)
    |> foreign_key_constraint(:project_id)
    |> foreign_key_constraint(:user_id)
    |> foreign_key_constraint(:event_id)
  end
end
