defmodule ElixirBackend.Workers.QuizSessionTransitions do
  @moduledoc """
  Cron worker that transitions quiz sessions based on their scheduled times.

  Runs every minute and checks for sessions that need state changes:
  - DRAFT → OPEN when open_at has passed
  - OPEN → LOCKED when lock_at has passed
  - LOCKED → FINISHED when finish_at has passed
  """

  use Oban.Worker, queue: :maintenance, max_attempts: 1

  @impl Oban.Worker
  def perform(_job) do
    ElixirBackend.Quizzes.transition_due_sessions()
    :ok
  end
end
