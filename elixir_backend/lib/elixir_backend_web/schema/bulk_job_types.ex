defmodule ElixirBackendWeb.Schema.BulkJobTypes do
  use Absinthe.Schema.Notation
  @moduledoc false

  enum :bulk_job_status do
    value(:pending, as: "PENDING")
    value(:processing, as: "PROCESSING")
    value(:completed, as: "COMPLETED")
    value(:failed, as: "FAILED")
  end

  object :bulk_job do
    field :id, non_null(:id)
    field :operation_type, non_null(:string)
    field :status, non_null(:bulk_job_status)
    field :total_count, non_null(:integer)
    field :processed_count, non_null(:integer)
    field :success_count, non_null(:integer)
    field :failure_count, non_null(:integer)
    field :error_message, :string
    field :created_at, non_null(:datetime)
    field :started_at, :datetime
    field :completed_at, :datetime
  end

  input_object :bulk_job_filter do
    field :status, :bulk_job_status
    field :operation_type, :string
    field :project_id, :id
    field :created_by, :id
  end

  object :bulk_job_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:bulk_job)
  end

  object :bulk_job_connection do
    field :edges, non_null(list_of(non_null(:bulk_job_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end
end
