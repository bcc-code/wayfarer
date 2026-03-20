defmodule ElixirBackend.Workers.BulkJobCleanup do
  @moduledoc """
  Cron worker that deletes completed and failed bulk jobs older than 7 days.
  """

  use Oban.Worker, queue: :maintenance, max_attempts: 1

  require Logger

  @impl Oban.Worker
  def perform(_job) do
    deleted = ElixirBackend.BulkJobs.cleanup_old_jobs()

    if deleted > 0 do
      Logger.info("Cleaned up #{deleted} old bulk jobs")
    end

    :ok
  end
end
