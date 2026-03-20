defmodule ElixirBackend.Teams.SuperTeam do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "super_teams" do
    field :name, :string
    field :description, :string
    field :image_url, :string
    field :color, :string

    belongs_to :project, ElixirBackend.Projects.Project, type: :string
    has_many :teams, ElixirBackend.Teams.Team

    timestamps(type: :utc_datetime)
  end

  def changeset(super_team, attrs) do
    super_team
    |> cast(attrs, [:id, :name, :description, :image_url, :color, :project_id])
    |> validate_required([:id, :name, :project_id])
    |> foreign_key_constraint(:project_id)
  end

  def update_changeset(super_team, attrs) do
    super_team
    |> cast(attrs, [:name, :description, :image_url, :color])
  end
end
