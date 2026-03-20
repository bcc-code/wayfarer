defmodule ElixirBackend.Repo.Migrations.CreateUserChallengeCompletions do
  use Ecto.Migration

  def change do
    create table(:user_challenge_completions, primary_key: false) do
      add :user_id, references(:users, type: :string, on_delete: :delete_all), primary_key: true

      add :challenge_id, references(:challenges, type: :string, on_delete: :delete_all),
        primary_key: true

      add :completed_at, :utc_datetime, default: fragment("NOW()")
    end

    create index(:user_challenge_completions, [:user_id])
    create index(:user_challenge_completions, [:challenge_id])
  end
end
