defmodule ElixirBackend.Repo.Migrations.CreateBulkJobs do
  use Ecto.Migration

  def change do
    create table(:bulk_jobs, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :operation_type, :string, size: 100, null: false
      add :status, :string, size: 50, null: false, default: "PENDING"
      add :created_by, :string, size: 28
      add :project_id, references(:projects, type: :string, on_delete: :nilify_all)
      add :input_params, :jsonb, default: "{}"
      add :total_count, :integer, null: false, default: 0
      add :processed_count, :integer, null: false, default: 0
      add :success_count, :integer, null: false, default: 0
      add :failure_count, :integer, null: false, default: 0
      add :error_message, :text
      add :error_details, :jsonb
      add :logs, :jsonb, default: "[]"
      add :message_id, :string, size: 100
      add :created_at, :utc_datetime, null: false, default: fragment("now()")
      add :started_at, :utc_datetime
      add :completed_at, :utc_datetime
    end

    create index(:bulk_jobs, [:status])
    create index(:bulk_jobs, [:created_by])
    create index(:bulk_jobs, [:project_id])
    create index(:bulk_jobs, [:status, :created_at])
  end
end
