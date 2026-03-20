defmodule ElixirBackend.BulkJobs.BulkJob do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}
  @valid_statuses ~w(PENDING PROCESSING COMPLETED FAILED)

  schema "bulk_jobs" do
    field :operation_type, :string
    field :status, :string, default: "PENDING"
    field :created_by, :string
    field :project_id, :string
    field :input_params, :map, default: %{}
    field :total_count, :integer, default: 0
    field :processed_count, :integer, default: 0
    field :success_count, :integer, default: 0
    field :failure_count, :integer, default: 0
    field :error_message, :string
    field :error_details, :map
    field :logs, {:array, :map}, default: []
    field :message_id, :string
    field :created_at, :utc_datetime
    field :started_at, :utc_datetime
    field :completed_at, :utc_datetime
  end

  def changeset(job, attrs) do
    job
    |> cast(attrs, [
      :id,
      :operation_type,
      :status,
      :created_by,
      :project_id,
      :input_params,
      :total_count,
      :processed_count,
      :success_count,
      :failure_count,
      :error_message,
      :error_details,
      :logs,
      :message_id,
      :created_at,
      :started_at,
      :completed_at
    ])
    |> validate_required([:id, :operation_type])
    |> validate_inclusion(:status, @valid_statuses)
  end
end
