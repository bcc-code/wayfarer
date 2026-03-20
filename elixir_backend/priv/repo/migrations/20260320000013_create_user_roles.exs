defmodule ElixirBackend.Repo.Migrations.CreateUserRoles do
  use Ecto.Migration

  def change do
    create table(:user_roles, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :role, :string, size: 50, null: false
      add :church_id, references(:churches, type: :string, on_delete: :nilify_all)
      add :project_id, references(:projects, type: :string, on_delete: :nilify_all)
      add :team_id, references(:teams, type: :string, on_delete: :nilify_all)
      add :assigned_by, references(:users, type: :string, on_delete: :restrict)
      add :assigned_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:user_roles, [:user_id, :role, :church_id, :project_id, :team_id],
      name: :user_roles_unique_assignment
    )

    create index(:user_roles, [:user_id])
    create index(:user_roles, [:role])
    create index(:user_roles, [:church_id])
    create index(:user_roles, [:project_id])
    create index(:user_roles, [:team_id])
  end
end
