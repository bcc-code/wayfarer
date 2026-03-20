defmodule ElixirBackend.Teams.TeamMember do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "team_members" do
    field :is_team_lead, :boolean, default: false
    field :joined_at, :utc_datetime

    belongs_to :team, ElixirBackend.Teams.Team, type: :string
    belongs_to :user, ElixirBackend.Accounts.User, type: :string
  end

  def changeset(member, attrs) do
    member
    |> cast(attrs, [:team_id, :user_id, :is_team_lead, :joined_at])
    |> validate_required([:team_id, :user_id, :joined_at])
    |> unique_constraint([:team_id, :user_id])
    |> foreign_key_constraint(:team_id)
    |> foreign_key_constraint(:user_id)
  end
end
