defmodule ElixirBackend.Teams.Team do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "teams" do
    field :name, :string
    field :description, :string
    field :join_code, :string
    field :leaderboard_excluded, :boolean, default: false

    belongs_to :project, ElixirBackend.Projects.Project, type: :string
    belongs_to :super_team, ElixirBackend.Teams.SuperTeam, type: :string
    has_many :team_members, ElixirBackend.Teams.TeamMember

    timestamps(type: :utc_datetime)
  end

  def changeset(team, attrs) do
    team
    |> cast(attrs, [
      :id,
      :name,
      :description,
      :join_code,
      :leaderboard_excluded,
      :project_id,
      :super_team_id
    ])
    |> validate_required([:id, :name, :join_code, :project_id])
    |> unique_constraint(:join_code)
    |> foreign_key_constraint(:project_id)
    |> foreign_key_constraint(:super_team_id)
  end

  def update_changeset(team, attrs) do
    team
    |> cast(attrs, [:name, :description, :leaderboard_excluded])
  end
end
