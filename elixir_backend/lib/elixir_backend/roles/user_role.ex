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

    scope = %{
      church_id: get_field(changeset, :church_id),
      project_id: get_field(changeset, :project_id),
      team_id: get_field(changeset, :team_id)
    }

    case check_scope(role, scope) do
      :ok -> changeset
      {:error, field, message} -> add_error(changeset, field, message)
    end
  end

  defp check_scope(role, scope) when role in @global_roles do
    if scope.church_id || scope.project_id || scope.team_id do
      {:error, :role, "global roles must not have a scope"}
    else
      :ok
    end
  end

  defp check_scope("CHURCH_ADMIN", %{church_id: nil}),
    do: {:error, :church_id, "CHURCH_ADMIN requires a church scope"}

  defp check_scope("PROJECT_ADMIN", %{project_id: nil}),
    do: {:error, :project_id, "PROJECT_ADMIN requires a project scope"}

  defp check_scope("TEAM_LEAD", %{team_id: nil}),
    do: {:error, :team_id, "TEAM_LEAD requires a team scope"}

  defp check_scope(_role, _scope), do: :ok
end
