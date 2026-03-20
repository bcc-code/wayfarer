defmodule ElixirBackend.Accounts.UserProject do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "user_projects" do
    field :user_id, :string
    field :project_id, :string
    field :joined_at, :utc_datetime
  end

  def changeset(user_project, attrs) do
    user_project
    |> cast(attrs, [:user_id, :project_id, :joined_at])
    |> validate_required([:user_id, :project_id, :joined_at])
    |> unique_constraint([:user_id, :project_id])
    |> foreign_key_constraint(:user_id)
    |> foreign_key_constraint(:project_id)
  end
end
