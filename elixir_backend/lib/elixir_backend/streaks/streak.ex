defmodule ElixirBackend.Streaks.Streak do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "streaks" do
    field :name, :string
    field :description, :string

    belongs_to :project, ElixirBackend.Projects.Project, type: :string
    has_many :relevant_days, ElixirBackend.Streaks.StreakRelevantDay

    timestamps(type: :utc_datetime)
  end

  def changeset(streak, attrs) do
    streak
    |> cast(attrs, [:id, :name, :description, :project_id])
    |> validate_required([:id, :name, :project_id])
    |> foreign_key_constraint(:project_id)
  end

  def update_changeset(streak, attrs) do
    streak
    |> cast(attrs, [:name, :description])
  end
end
