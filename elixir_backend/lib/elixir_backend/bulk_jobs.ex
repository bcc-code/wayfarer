defmodule ElixirBackend.BulkJobs do
  @moduledoc """
  Context module for bulk job tracking.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.BulkJobs.BulkJob

  def get_job(id) do
    case Repo.get(BulkJob, id) do
      nil -> {:error, :not_found}
      job -> {:ok, job}
    end
  end

  def create_job(attrs) do
    id = ULID.new_bulk_job_id()
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    %BulkJob{}
    |> BulkJob.changeset(
      attrs
      |> Map.put(:id, id)
      |> Map.put(:created_at, now)
    )
    |> Repo.insert()
  end

  def list_jobs(filter \\ %{}, pagination_opts \\ %{}) do
    query = from(j in BulkJob)
    query = apply_filter(query, filter)
    total_count = Repo.aggregate(query, :count)

    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    items =
      query
      |> order_by([j], desc: j.created_at, desc: j.id)
      |> limit(^limit)
      |> Repo.all()

    edges = Enum.map(items, fn item -> %{cursor: item.id, node: item} end)

    {:ok,
     %{
       edges: edges,
       page_info: %{has_next_page: length(items) == limit, has_previous_page: false},
       total_count: total_count
     }}
  end

  def my_jobs(user_id, limit \\ 10) do
    from(j in BulkJob,
      where: j.created_by == ^user_id,
      order_by: [desc: j.created_at],
      limit: ^limit
    )
    |> Repo.all()
  end

  def mark_processing(id) do
    with {:ok, job} <- get_job(id) do
      now = DateTime.utc_now() |> DateTime.truncate(:second)

      job
      |> Ecto.Changeset.change(status: "PROCESSING", started_at: now)
      |> Repo.update()
    end
  end

  def mark_completed(id) do
    with {:ok, job} <- get_job(id) do
      now = DateTime.utc_now() |> DateTime.truncate(:second)

      job
      |> Ecto.Changeset.change(status: "COMPLETED", completed_at: now)
      |> Repo.update()
    end
  end

  def mark_failed(id, error_message) do
    with {:ok, job} <- get_job(id) do
      now = DateTime.utc_now() |> DateTime.truncate(:second)

      job
      |> Ecto.Changeset.change(
        status: "FAILED",
        error_message: error_message,
        completed_at: now
      )
      |> Repo.update()
    end
  end

  def update_progress(id, attrs) do
    with {:ok, job} <- get_job(id) do
      job
      |> BulkJob.changeset(attrs)
      |> Repo.update()
    end
  end

  # ── Private ──

  defp apply_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:status, s}, q when is_binary(s) ->
        where(q, [j], j.status == ^s)

      {:operation_type, t}, q when is_binary(t) ->
        where(q, [j], j.operation_type == ^t)

      {:project_id, pid}, q when is_binary(pid) ->
        where(q, [j], j.project_id == ^pid)

      {:created_by, uid}, q when is_binary(uid) ->
        where(q, [j], j.created_by == ^uid)

      _, q ->
        q
    end)
  end

  defp apply_filter(query, _), do: query
end
