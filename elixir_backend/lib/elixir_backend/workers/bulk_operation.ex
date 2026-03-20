defmodule ElixirBackend.Workers.BulkOperation do
  @moduledoc """
  Worker that processes bulk operations asynchronously.

  Supports the following operation types:
  - BULK_COMPLETE_CHALLENGE — complete a challenge for multiple users
  - BULK_AWARD_ACHIEVEMENT — award an achievement to multiple users

  Each job is linked to a BulkJob record for status tracking and progress reporting.
  """

  use Oban.Worker, queue: :bulk, max_attempts: 3

  require Logger

  alias ElixirBackend.BulkJobs

  @batch_size 100

  @impl Oban.Worker
  def perform(%Oban.Job{args: %{"bulk_job_id" => bulk_job_id}}) do
    with {:ok, job} <- BulkJobs.get_job(bulk_job_id),
         {:ok, _} <- BulkJobs.mark_processing(bulk_job_id) do
      result = process_operation(job.operation_type, job)

      case result do
        :ok ->
          BulkJobs.mark_completed(bulk_job_id)
          :ok

        {:error, reason} ->
          BulkJobs.mark_failed(bulk_job_id, to_string(reason))
          {:error, reason}
      end
    else
      {:error, :not_found} ->
        Logger.error("BulkOperation: job #{bulk_job_id} not found")
        {:error, :not_found}

      {:error, reason} ->
        Logger.error("BulkOperation: failed to start job #{bulk_job_id}: #{inspect(reason)}")
        {:error, reason}
    end
  end

  defp process_operation("BULK_COMPLETE_CHALLENGE", job) do
    challenge_id = job.input_params["challenge_id"]
    user_ids = job.input_params["user_ids"] || []

    process_in_batches(job.id, user_ids, fn user_id ->
      ElixirBackend.Challenges.complete_challenge(user_id, challenge_id)
    end)
  end

  defp process_operation("BULK_AWARD_ACHIEVEMENT", job) do
    achievement_id = job.input_params["achievement_id"]
    user_ids = job.input_params["user_ids"] || []

    process_in_batches(job.id, user_ids, fn user_id ->
      ElixirBackend.Achievements.award_achievement(user_id, achievement_id)
    end)
  end

  defp process_operation(operation_type, _job) do
    Logger.warning("BulkOperation: unknown operation type #{operation_type}")
    {:error, "unknown operation type: #{operation_type}"}
  end

  defp process_in_batches(bulk_job_id, items, process_fn) do
    total = length(items)
    BulkJobs.update_progress(bulk_job_id, %{total_count: total})

    _totals =
      items
      |> Enum.chunk_every(@batch_size)
      |> Enum.reduce({0, 0, 0}, fn batch, acc ->
        process_batch(bulk_job_id, batch, process_fn, acc)
      end)

    :ok
  end

  defp process_batch(bulk_job_id, batch, process_fn, {processed, successes, failures}) do
    {batch_ok, batch_fail} = count_results(batch, process_fn)
    totals = {processed + length(batch), successes + batch_ok, failures + batch_fail}

    BulkJobs.update_progress(bulk_job_id, %{
      processed_count: elem(totals, 0),
      success_count: elem(totals, 1),
      failure_count: elem(totals, 2)
    })

    totals
  end

  defp count_results(batch, process_fn) do
    Enum.reduce(batch, {0, 0}, fn item, {ok, fail} ->
      case process_fn.(item) do
        {:ok, _} -> {ok + 1, fail}
        :ok -> {ok + 1, fail}
        _ -> {ok, fail + 1}
      end
    end)
  end
end
