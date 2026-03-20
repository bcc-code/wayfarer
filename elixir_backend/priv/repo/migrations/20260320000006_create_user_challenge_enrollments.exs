defmodule ElixirBackend.Repo.Migrations.CreateUserChallengeEnrollments do
  use Ecto.Migration

  def change do
    create table(:user_challenge_enrollments, primary_key: false) do
      add :user_id, references(:users, type: :string, on_delete: :delete_all), primary_key: true

      add :challenge_id, references(:challenges, type: :string, on_delete: :delete_all),
        primary_key: true

      add :enrolled_at, :utc_datetime, default: fragment("NOW()")
    end

    create index(:user_challenge_enrollments, [:user_id])
    create index(:user_challenge_enrollments, [:challenge_id])
  end
end
