defmodule ElixirBackend.Roles.UserRole do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  @valid_roles ~w(SUPERADMIN ADMIN CHURCH_ADMIN PROJECT_ADMIN TEAM_LEAD USER M2M)
  @global_roles ~w(SUPERADMIN ADMIN USER M2M)

  schema "user_roles" do
    field :role, :string
    field :assigned_at, :utc_datetime

    belongs_to :user, ElixirBackend.Accounts.User, type: :string
    belongs_to :church, ElixirBackend.Churches.Church, type: :string
    belongs_to :project, ElixirBackend.Projects.Project, type: :string
    belongs_to :team, ElixirBackend.Teams.Team, type: :string

    belongs_to :assigned_by_user, ElixirBackend.Accounts.User,
      type: :string,
      foreign_key: :assigned_by
  end

  def changeset(user_role, attrs) do
    user_role
    |> cast(attrs, [
      :id,
      :user_id,
      :role,
      :church_id,
      :project_id,
      :team_id,
      :assigned_by,
      :assigned_at
    ])
    |> validate_required([:id, :user_id, :role, :assigned_at])
    |> validate_inclusion(:role, @valid_roles)
    |> validate_scope()
    |> foreign_key_constraint(:user_id)
    |> foreign_key_constraint(:church_id)
    |> foreign_key_constraint(:project_id)
    |> foreign_key_constraint(:team_id)
    |> unique_constraint([:user_id, :role, :church_id, :project_id, :team_id],
      name: :user_roles_unique_assignment
    )
  end

  defp validate_scope(changeset) do
    role = get_field(changeset, :role)
    church_id = get_field(changeset, :church_id)
    project_id = get_field(changeset, :project_id)
    team_id = get_field(changeset, :team_id)

    cond do
      role in @global_roles and (church_id || project_id || team_id) ->
        add_error(changeset, :role, "global roles must not have a scope")

      role == "CHURCH_ADMIN" and is_nil(church_id) ->
        add_error(changeset, :church_id, "CHURCH_ADMIN requires a church scope")

      role == "PROJECT_ADMIN" and is_nil(project_id) ->
        add_error(changeset, :project_id, "PROJECT_ADMIN requires a project scope")

      role == "TEAM_LEAD" and is_nil(team_id) ->
        add_error(changeset, :team_id, "TEAM_LEAD requires a team scope")

      true ->
        changeset
    end
  end
end
