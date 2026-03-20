defmodule ElixirBackend.Cache do
  @moduledoc """
  Caching layer using Cachex. Mirrors the Go backend's cache key structure.

  Cache names:
  - `:challenges` — Challenge structs by ID
  - `:user_challenge_completions` — Completion timestamps by {user_id, challenge_id}
  - `:user_challenge_enrollments` — Enrollment timestamps by {user_id, challenge_id}
  """

  # Default TTL: 15 minutes (matching Go backend)
  @default_ttl :timer.minutes(15)

  # ── Cache names ──

  def challenge_cache, do: :challenges
  def completion_cache, do: :user_challenge_completions
  def enrollment_cache, do: :user_challenge_enrollments

  @doc "All cache specs for the supervision tree."
  def child_specs do
    [
      Supervisor.child_spec({Cachex, challenge_cache()}, id: challenge_cache()),
      Supervisor.child_spec({Cachex, completion_cache()}, id: completion_cache()),
      Supervisor.child_spec({Cachex, enrollment_cache()}, id: enrollment_cache())
    ]
  end

  # ── Challenge cache ──

  @doc "Cache key for a challenge by ID."
  def challenge_key(challenge_id), do: "challenge:#{challenge_id}"

  @doc "Get a challenge from cache."
  def get_challenge(challenge_id) do
    case Cachex.get(challenge_cache(), challenge_key(challenge_id)) do
      {:ok, nil} -> :miss
      {:ok, value} -> {:ok, value}
      _ -> :miss
    end
  end

  @doc "Store a challenge in cache with default TTL."
  def put_challenge(challenge) do
    Cachex.put(challenge_cache(), challenge_key(challenge.id), challenge, expire: @default_ttl)
  end

  @doc "Invalidate a cached challenge."
  def delete_challenge(challenge_id) do
    Cachex.del(challenge_cache(), challenge_key(challenge_id))
  end

  @doc """
  Get-or-compute a challenge. Calls the fallback function on cache miss
  and caches the result.
  """
  def fetch_challenge(challenge_id, fallback) do
    Cachex.fetch(challenge_cache(), challenge_key(challenge_id), fn _key ->
      case fallback.() do
        {:ok, challenge} -> {:commit, challenge, expire: @default_ttl}
        error -> {:ignore, error}
      end
    end)
    |> normalize_fetch_result()
  end

  # ── User challenge completion cache ──

  @doc "Cache key for a user's challenge completion timestamp."
  def completion_key(user_id, challenge_id), do: "#{user_id}:#{challenge_id}"

  @doc "Get a user's completion timestamp from cache."
  def get_completion(user_id, challenge_id) do
    case Cachex.get(completion_cache(), completion_key(user_id, challenge_id)) do
      {:ok, nil} -> :miss
      {:ok, :cached_nil} -> {:ok, nil}
      {:ok, value} -> {:ok, value}
      _ -> :miss
    end
  end

  @doc "Store a user's completion timestamp (or nil for uncompleted)."
  def put_completion(user_id, challenge_id, timestamp) do
    # Use :cached_nil sentinel to distinguish "cached nil" from "not in cache"
    value = if is_nil(timestamp), do: :cached_nil, else: timestamp

    Cachex.put(completion_cache(), completion_key(user_id, challenge_id), value,
      expire: @default_ttl
    )
  end

  @doc "Invalidate a cached completion timestamp."
  def delete_completion(user_id, challenge_id) do
    Cachex.del(completion_cache(), completion_key(user_id, challenge_id))
  end

  # ── User challenge enrollment cache ──

  @doc "Cache key for a user's challenge enrollment timestamp."
  def enrollment_key(user_id, challenge_id), do: "#{user_id}:#{challenge_id}"

  @doc "Get a user's enrollment timestamp from cache."
  def get_enrollment(user_id, challenge_id) do
    case Cachex.get(enrollment_cache(), enrollment_key(user_id, challenge_id)) do
      {:ok, nil} -> :miss
      {:ok, :cached_nil} -> {:ok, nil}
      {:ok, value} -> {:ok, value}
      _ -> :miss
    end
  end

  @doc "Store a user's enrollment timestamp (or nil for unenrolled)."
  def put_enrollment(user_id, challenge_id, timestamp) do
    value = if is_nil(timestamp), do: :cached_nil, else: timestamp

    Cachex.put(enrollment_cache(), enrollment_key(user_id, challenge_id), value,
      expire: @default_ttl
    )
  end

  @doc "Invalidate a cached enrollment timestamp."
  def delete_enrollment(user_id, challenge_id) do
    Cachex.del(enrollment_cache(), enrollment_key(user_id, challenge_id))
  end

  # ── Utilities ──

  @doc "Clear all caches (for testing)."
  def clear_all do
    Cachex.clear(challenge_cache())
    Cachex.clear(completion_cache())
    Cachex.clear(enrollment_cache())
    :ok
  end

  # Normalize Cachex.fetch results to {:ok, value} | {:error, reason}
  defp normalize_fetch_result({:commit, value}), do: {:ok, value}
  defp normalize_fetch_result({:ok, value}), do: {:ok, value}
  defp normalize_fetch_result({:ignore, {:error, reason}}), do: {:error, reason}
  defp normalize_fetch_result({:ignore, error}), do: {:error, error}
  defp normalize_fetch_result({:error, reason}), do: {:error, reason}
end
