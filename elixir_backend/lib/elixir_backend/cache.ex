defmodule ElixirBackend.Cache do
  @moduledoc """
  Caching layer using Cachex. Mirrors the Go backend's cache key structure.

  Cache instances:
  - `:challenges` — Challenge structs by ID
  - `:user_challenge_completions` — Completion timestamps by {user_id, challenge_id}
  - `:user_challenge_enrollments` — Enrollment timestamps by {user_id, challenge_id}
  - `:users` — User structs by ID
  - `:general` — All other entity lookups, relationships, roles, leaderboards
  """

  # Default TTL: 15 minutes (matching Go backend)
  @default_ttl :timer.minutes(15)

  # ── Cache names ──

  def challenge_cache, do: :challenges
  def completion_cache, do: :user_challenge_completions
  def enrollment_cache, do: :user_challenge_enrollments
  def user_cache, do: :users
  def general_cache, do: :general

  @doc "All cache specs for the supervision tree."
  def child_specs do
    [
      Supervisor.child_spec({Cachex, challenge_cache()}, id: challenge_cache()),
      Supervisor.child_spec({Cachex, completion_cache()}, id: completion_cache()),
      Supervisor.child_spec({Cachex, enrollment_cache()}, id: enrollment_cache()),
      Supervisor.child_spec({Cachex, user_cache()}, id: user_cache()),
      Supervisor.child_spec({Cachex, general_cache()}, id: general_cache())
    ]
  end

  # ══════════════════════════════════════════════════════════
  # General cache: get / put / del / fetch
  # ══════════════════════════════════════════════════════════

  @doc "Get a value from the general cache by key."
  def get(key) do
    case Cachex.get(general_cache(), key) do
      {:ok, nil} -> :miss
      {:ok, value} -> {:ok, value}
      _ -> :miss
    end
  end

  @doc "Store a value in the general cache."
  def put(key, value, opts \\ []) do
    ttl = opts[:ttl] || @default_ttl
    Cachex.put(general_cache(), key, value, expire: ttl)
  end

  @doc "Delete a key from the general cache."
  def del(key) do
    Cachex.del(general_cache(), key)
  end

  @doc "Get-or-compute from general cache."
  def fetch(key, fallback) do
    Cachex.fetch(general_cache(), key, fn _key ->
      case fallback.() do
        {:ok, value} -> {:commit, value, expire: @default_ttl}
        error -> {:ignore, error}
      end
    end)
    |> normalize_fetch_result()
  end

  @doc "Get-or-compute from general cache, wrapping raw values in {:ok, value}."
  def fetch_raw(key, fallback) do
    Cachex.fetch(general_cache(), key, fn _key ->
      {:commit, fallback.(), expire: @default_ttl}
    end)
    |> normalize_fetch_raw_result()
  end

  # ══════════════════════════════════════════════════════════
  # Entity cache keys
  # ══════════════════════════════════════════════════════════

  def church_key(id), do: "church:#{id}"
  def project_key(id), do: "project:#{id}"
  def event_key(id), do: "event:#{id}"
  def team_key(id), do: "team:#{id}"
  def super_team_key(id), do: "superteam:#{id}"
  def achievement_key(id), do: "achievement:#{id}"
  def streak_key(id), do: "streak:#{id}"
  def quiz_key(id), do: "quiz:#{id}"
  def consent_key(id), do: "consent:#{id}"

  # ══════════════════════════════════════════════════════════
  # Relationship cache keys
  # ══════════════════════════════════════════════════════════

  def challenges_by_project_key(pid), do: "challenges:project:#{pid}"
  def challenges_by_event_key(eid), do: "challenges:event:#{eid}"
  def events_by_project_key(pid), do: "events:project:#{pid}"
  def teams_by_project_key(pid), do: "teams:project:#{pid}"
  def teams_by_superteam_key(stid), do: "teams:superteam:#{stid}"
  def teams_by_user_key(uid), do: "teams:user:#{uid}"
  def projects_by_user_key(uid), do: "projects:user:#{uid}"
  def events_by_user_key(uid), do: "events:user:#{uid}"
  def achievements_by_project_key(pid), do: "achievements:project:#{pid}"
  def streaks_by_project_key(pid), do: "streaks:project:#{pid}"
  def quiz_questions_key(qid), do: "quiz:questions:#{qid}"
  def quiz_answers_key(qqid), do: "quiz:answers:#{qqid}"
  def consents_latest_key, do: "consents:latest"
  def team_members_key(tid), do: "team:members:#{tid}"

  # ══════════════════════════════════════════════════════════
  # Translation cache keys
  # ══════════════════════════════════════════════════════════

  def translation_key(entity_type, entity_id, language_code),
    do: "translation:#{entity_type}:#{entity_id}:#{language_code}"

  # ══════════════════════════════════════════════════════════
  # Role cache keys
  # ══════════════════════════════════════════════════════════

  def user_roles_key(uid), do: "userroles:#{uid}"

  # ══════════════════════════════════════════════════════════
  # User progress cache keys
  # ══════════════════════════════════════════════════════════

  def user_achieved_at_key(uid, aid), do: "userachievement:#{uid}:#{aid}"
  def user_streak_activity_key(uid, sid), do: "userstreak:#{uid}:#{sid}"

  # ══════════════════════════════════════════════════════════
  # Leaderboard cache keys
  # ══════════════════════════════════════════════════════════

  def leaderboard_key(scope, scope_id, entity_type),
    do: "leaderboard:#{scope}:#{scope_id}:#{entity_type}"

  # ══════════════════════════════════════════════════════════
  # Invalidation helpers
  # ══════════════════════════════════════════════════════════

  @doc "Invalidate all caches related to a church."
  def invalidate_church(id) do
    del(church_key(id))
  end

  @doc "Invalidate all caches related to a project."
  def invalidate_project(id) do
    del(project_key(id))
    del(challenges_by_project_key(id))
    del(events_by_project_key(id))
    del(teams_by_project_key(id))
    del(achievements_by_project_key(id))
    del(streaks_by_project_key(id))
    invalidate_leaderboards(id)
  end

  @doc "Invalidate all caches related to an event."
  def invalidate_event(event_id, project_id \\ nil) do
    del(event_key(event_id))
    del(challenges_by_event_key(event_id))
    if project_id, do: del(events_by_project_key(project_id))
  end

  @doc "Invalidate all caches related to a team."
  def invalidate_team(team_id, project_id \\ nil) do
    del(team_key(team_id))
    del(team_members_key(team_id))
    if project_id, do: del(teams_by_project_key(project_id))
  end

  @doc "Invalidate all caches related to a super team."
  def invalidate_super_team(id, project_id \\ nil) do
    del(super_team_key(id))
    del(teams_by_superteam_key(id))
    if project_id, do: del(teams_by_project_key(project_id))
  end

  @doc "Invalidate all caches related to an achievement."
  def invalidate_achievement(id, project_id \\ nil) do
    del(achievement_key(id))
    if project_id, do: del(achievements_by_project_key(project_id))
  end

  @doc "Invalidate all caches related to a streak."
  def invalidate_streak(id, project_id \\ nil) do
    del(streak_key(id))
    if project_id, do: del(streaks_by_project_key(project_id))
  end

  @doc "Invalidate all caches related to a quiz."
  def invalidate_quiz(id) do
    del(quiz_key(id))
    del(quiz_questions_key(id))
  end

  @doc "Invalidate quiz question answer caches."
  def invalidate_quiz_answers(question_id) do
    del(quiz_answers_key(question_id))
  end

  @doc "Invalidate consent caches."
  def invalidate_consent(id) do
    del(consent_key(id))
    del(consents_latest_key())
  end

  @doc "Invalidate user role caches."
  def invalidate_user_roles(user_id) do
    del(user_roles_key(user_id))
  end

  @doc "Invalidate user team membership caches."
  def invalidate_user_teams(user_id) do
    del(teams_by_user_key(user_id))
  end

  @doc "Invalidate user project/event membership caches."
  def invalidate_user_memberships(user_id) do
    del(projects_by_user_key(user_id))
    del(events_by_user_key(user_id))
  end

  @doc "Invalidate leaderboard caches for a given scope."
  def invalidate_leaderboards(scope_id) do
    for entity_type <- ["persons", "teams", "superteams", "churches"],
        scope <- ["project", "event"] do
      del(leaderboard_key(scope, scope_id, entity_type))
    end

    :ok
  end

  @doc "Invalidate user achievement progress."
  def invalidate_user_achievement(user_id, achievement_id) do
    del(user_achieved_at_key(user_id, achievement_id))
  end

  @doc "Invalidate user streak activity."
  def invalidate_user_streak(user_id, streak_id) do
    del(user_streak_activity_key(user_id, streak_id))
  end

  # ══════════════════════════════════════════════════════════
  # Challenge cache (existing, kept for backward compat)
  # ══════════════════════════════════════════════════════════

  def challenge_key(challenge_id), do: "challenge:#{challenge_id}"

  def get_challenge(challenge_id) do
    case Cachex.get(challenge_cache(), challenge_key(challenge_id)) do
      {:ok, nil} -> :miss
      {:ok, value} -> {:ok, value}
      _ -> :miss
    end
  end

  def put_challenge(challenge) do
    Cachex.put(challenge_cache(), challenge_key(challenge.id), challenge, expire: @default_ttl)
    # Also invalidate relationship caches
    if Map.get(challenge, :project_id), do: del(challenges_by_project_key(challenge.project_id))
    if Map.get(challenge, :event_id), do: del(challenges_by_event_key(challenge.event_id))
  end

  def delete_challenge(challenge_id) do
    Cachex.del(challenge_cache(), challenge_key(challenge_id))
  end

  def fetch_challenge(challenge_id, fallback) do
    Cachex.fetch(challenge_cache(), challenge_key(challenge_id), fn _key ->
      case fallback.() do
        {:ok, challenge} -> {:commit, challenge, expire: @default_ttl}
        error -> {:ignore, error}
      end
    end)
    |> normalize_fetch_result()
  end

  # ══════════════════════════════════════════════════════════
  # User challenge completion cache (existing)
  # ══════════════════════════════════════════════════════════

  def completion_key(user_id, challenge_id), do: "#{user_id}:#{challenge_id}"

  def get_completion(user_id, challenge_id) do
    case Cachex.get(completion_cache(), completion_key(user_id, challenge_id)) do
      {:ok, nil} -> :miss
      {:ok, :cached_nil} -> {:ok, nil}
      {:ok, value} -> {:ok, value}
      _ -> :miss
    end
  end

  def put_completion(user_id, challenge_id, timestamp) do
    value = if is_nil(timestamp), do: :cached_nil, else: timestamp

    Cachex.put(completion_cache(), completion_key(user_id, challenge_id), value,
      expire: @default_ttl
    )
  end

  def delete_completion(user_id, challenge_id) do
    Cachex.del(completion_cache(), completion_key(user_id, challenge_id))
  end

  # ══════════════════════════════════════════════════════════
  # User challenge enrollment cache (existing)
  # ══════════════════════════════════════════════════════════

  def enrollment_key(user_id, challenge_id), do: "#{user_id}:#{challenge_id}"

  def get_enrollment(user_id, challenge_id) do
    case Cachex.get(enrollment_cache(), enrollment_key(user_id, challenge_id)) do
      {:ok, nil} -> :miss
      {:ok, :cached_nil} -> {:ok, nil}
      {:ok, value} -> {:ok, value}
      _ -> :miss
    end
  end

  def put_enrollment(user_id, challenge_id, timestamp) do
    value = if is_nil(timestamp), do: :cached_nil, else: timestamp

    Cachex.put(enrollment_cache(), enrollment_key(user_id, challenge_id), value,
      expire: @default_ttl
    )
  end

  def delete_enrollment(user_id, challenge_id) do
    Cachex.del(enrollment_cache(), enrollment_key(user_id, challenge_id))
  end

  # ══════════════════════════════════════════════════════════
  # User cache (existing)
  # ══════════════════════════════════════════════════════════

  def user_key(user_id), do: "user:#{user_id}"

  def get_user(user_id) do
    case Cachex.get(user_cache(), user_key(user_id)) do
      {:ok, nil} -> :miss
      {:ok, value} -> {:ok, value}
      _ -> :miss
    end
  end

  def put_user(user) do
    Cachex.put(user_cache(), user_key(user.id), user, expire: @default_ttl)
  end

  def delete_user(user_id) do
    Cachex.del(user_cache(), user_key(user_id))
  end

  def fetch_user(user_id, fallback) do
    Cachex.fetch(user_cache(), user_key(user_id), fn _key ->
      case fallback.() do
        {:ok, user} -> {:commit, user, expire: @default_ttl}
        error -> {:ignore, error}
      end
    end)
    |> normalize_fetch_result()
  end

  # ══════════════════════════════════════════════════════════
  # Utilities
  # ══════════════════════════════════════════════════════════

  @doc "Clear all caches (for testing)."
  def clear_all do
    Cachex.clear(challenge_cache())
    Cachex.clear(completion_cache())
    Cachex.clear(enrollment_cache())
    Cachex.clear(user_cache())
    Cachex.clear(general_cache())
    :ok
  end

  # Normalize Cachex.fetch results to {:ok, value} | {:error, reason}
  defp normalize_fetch_result({:commit, value}), do: {:ok, value}
  defp normalize_fetch_result({:ok, value}), do: {:ok, value}
  defp normalize_fetch_result({:ignore, {:error, reason}}), do: {:error, reason}
  defp normalize_fetch_result({:ignore, error}), do: {:error, error}
  defp normalize_fetch_result({:error, reason}), do: {:error, reason}

  defp normalize_fetch_raw_result({:commit, value}), do: value
  defp normalize_fetch_raw_result({:ok, value}), do: value
  defp normalize_fetch_raw_result({:error, _reason}), do: nil
end
