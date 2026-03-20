defmodule ElixirBackend.Workers.UserDataSync do
  @moduledoc """
  Cron worker that syncs incomplete user data from external Members API.

  Finds users with missing gender or church data and syncs from the upstream API.
  Currently a stub — the Members API client needs to be implemented.
  """

  use Oban.Worker, queue: :maintenance, max_attempts: 3

  require Logger

  @impl Oban.Worker
  def perform(_job) do
    # Stub: implement when Members API client is available.
    # Will query users with missing data and sync from upstream.
    Logger.debug("UserDataSync: no-op (Members API client not yet implemented)")
    :ok
  end
end
