defmodule ElixirBackendWeb.Schema.BulkJobQueries do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.BulkJobs

  object :bulk_job_queries do
    field :bulk_job, non_null(:bulk_job) do
      arg(:id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, %{id: id}, _ ->
        BulkJobs.get_job(id)
      end)
    end

    field :my_bulk_jobs, non_null(list_of(non_null(:bulk_job))) do
      arg(:limit, :integer)

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, args, %{context: context} ->
        limit = args[:limit] || 10

        case context[:current_user_id] do
          nil -> {:ok, []}
          user_id -> {:ok, BulkJobs.my_jobs(user_id, limit)}
        end
      end)
    end

    field :bulk_jobs, non_null(:bulk_job_connection) do
      arg(:filter, :bulk_job_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, args, _ ->
        filter = args[:filter] || %{}
        pagination = Map.take(args, [:first, :after, :last, :before])
        BulkJobs.list_jobs(filter, pagination)
      end)
    end
  end
end
