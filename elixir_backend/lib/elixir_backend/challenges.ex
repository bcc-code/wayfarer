defmodule ElixirBackend.Challenges do
  @moduledoc """
  Context module for challenge business logic.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Pagination
  alias ElixirBackend.Challenges.{Challenge, Completion, Enrollment}

  # ── Read ──

  def get_challenge(id) do
    case Repo.get(Challenge, id) do
      nil -> {:error, :not_found}
      challenge -> {:ok, challenge}
    end
  end

  def get_challenge!(id), do: Repo.get!(Challenge, id)

  @doc """
  Get a challenge with visibility filtering.
  - Admins see all challenges.
  - Enrolled users see challenges regardless of visible_at.
  - Other users only see challenges with visible_at in the past (or published_at in the past if no visible_at).
  """
  def get_visible_challenge(id, opts \\ []) do
    user_id = opts[:user_id]
    roles = opts[:roles] || []

    with {:ok, challenge} <- get_challenge(id) do
      cond do
        has_admin_role?(roles) ->
          {:ok, challenge}

        user_id && user_enrolled?(user_id, id) ->
          {:ok, challenge}

        visible_to_public?(challenge) ->
          {:ok, challenge}

        true ->
          {:error, :not_found}
      end
    end
  end

  @doc """
  List challenges with visibility filtering.
  """
  def list_visible_challenges(filter \\ %{}, pagination_opts \\ %{}, opts \\ []) do
    user_id = opts[:user_id]
    roles = opts[:roles] || []

    base_query = from(c in Challenge)

    query =
      base_query
      |> apply_filter(filter)
      |> apply_visibility_filter(user_id, roles)

    total_count = Repo.aggregate(query, :count)

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    Pagination.build_connection(items, pagination_opts, total_count)
  end

  # ── Paginated listing (no visibility filter — for tests/admin) ──

  def list_challenges(filter \\ %{}, pagination_opts \\ %{}) do
    base_query = from(c in Challenge)

    query =
      base_query
      |> apply_filter(filter)

    total_count = Repo.aggregate(query, :count)

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    Pagination.build_connection(items, pagination_opts, total_count)
  end

  # ── Create ──

  def create_challenge(attrs) do
    id = ULID.new_challenge_id()

    # Default published_at to now if not provided
    attrs =
      if Map.has_key?(attrs, :published_at) || Map.has_key?(attrs, "published_at") do
        attrs
      else
        Map.put(attrs, :published_at, DateTime.utc_now() |> DateTime.truncate(:second))
      end

    %Challenge{}
    |> Challenge.create_changeset(Map.put(attrs, :id, id))
    |> Repo.insert()
  end

  # ── Update ──

  def update_challenge(%Challenge{} = challenge, attrs) do
    challenge
    |> Challenge.update_changeset(attrs)
    |> Repo.update()
  end

  def update_challenge(id, attrs) when is_binary(id) do
    with {:ok, challenge} <- get_challenge(id) do
      update_challenge(challenge, attrs)
    end
  end

  # ── Delete ──

  def delete_challenge(%Challenge{} = challenge) do
    Repo.delete(challenge)
  end

  def delete_challenge(id) when is_binary(id) do
    with {:ok, challenge} <- get_challenge(id) do
      delete_challenge(challenge)
    end
  end

  # ── Publish ──

  def publish_challenge(id, published_at) do
    with {:ok, challenge} <- get_challenge(id) do
      challenge
      |> Ecto.Changeset.change(published_at: published_at)
      |> Repo.update()
    end
  end

  # ── Assign to event ──

  def assign_challenge_to_event(challenge_id, event_id) do
    with {:ok, challenge} <- get_challenge(challenge_id) do
      challenge
      |> Ecto.Changeset.change(event_id: event_id)
      |> Ecto.Changeset.foreign_key_constraint(:event_id)
      |> Repo.update()
    end
  end

  # ── Set visibility ──

  def set_challenge_visibility(id, visible_at, started_at \\ nil) do
    with {:ok, challenge} <- get_challenge(id) do
      changes = %{visible_at: visible_at}
      changes = if started_at, do: Map.put(changes, :started_at, started_at), else: changes

      challenge
      |> Ecto.Changeset.change(changes)
      |> Repo.update()
    end
  end

  # ── Set requirements ──

  def set_challenge_requirements(id, opts) do
    with {:ok, challenge} <- get_challenge(id) do
      changes =
        %{}
        |> maybe_put(:requires_team_membership, opts[:requires_team_membership])
        |> maybe_put(:requires_super_team_membership, opts[:requires_super_team_membership])

      challenge
      |> Ecto.Changeset.change(changes)
      |> Repo.update()
    end
  end

  # ── Completions ──

  def complete_challenge(user_id, challenge_id, completed_at \\ nil) do
    completed_at = completed_at || DateTime.utc_now() |> DateTime.truncate(:second)

    %Completion{}
    |> Completion.changeset(%{
      user_id: user_id,
      challenge_id: challenge_id,
      completed_at: completed_at
    })
    |> Repo.insert(on_conflict: :nothing, conflict_target: [:user_id, :challenge_id])
  end

  def uncomplete_challenge(user_id, challenge_id) do
    query =
      from(c in Completion,
        where: c.user_id == ^user_id and c.challenge_id == ^challenge_id
      )

    case Repo.delete_all(query) do
      {count, _} when count > 0 -> {:ok, true}
      _ -> {:ok, false}
    end
  end

  def self_complete_challenge(user_id, challenge_id) do
    with {:ok, challenge} <- get_challenge(challenge_id),
         :ok <- validate_self_completion(challenge) do
      complete_challenge(user_id, challenge_id)
      {:ok, challenge}
    end
  end

  defp validate_self_completion(
         %Challenge{challenge_type: "SIMPLE", allow_self_completion: true} = challenge
       ) do
    validate_challenge_active(challenge)
  end

  defp validate_self_completion(%Challenge{challenge_type: "SIMPLE"}),
    do: {:error, "self-completion not allowed for this challenge"}

  defp validate_self_completion(_),
    do: {:error, "self-completion is only available for simple challenges"}

  # ── Enrollments ──

  def enroll_in_challenge(user_id, challenge_id) do
    with {:ok, challenge} <- get_challenge(challenge_id),
         :ok <- validate_enrollment(challenge) do
      enrolled_at = DateTime.utc_now() |> DateTime.truncate(:second)

      result =
        %Enrollment{}
        |> Enrollment.changeset(%{
          user_id: user_id,
          challenge_id: challenge_id,
          enrolled_at: enrolled_at
        })
        |> Repo.insert(on_conflict: :nothing, conflict_target: [:user_id, :challenge_id])

      case result do
        {:ok, _} -> {:ok, challenge}
        error -> error
      end
    end
  end

  def unenroll_from_challenge(user_id, challenge_id) do
    query =
      from(e in Enrollment,
        where: e.user_id == ^user_id and e.challenge_id == ^challenge_id
      )

    case Repo.delete_all(query) do
      {count, _} when count > 0 -> {:ok, true}
      _ -> {:ok, false}
    end
  end

  def admin_complete_challenge(user_id, challenge_id, completed_at \\ nil) do
    complete_challenge(user_id, challenge_id, completed_at)
  end

  # ── Enrollment per-user lookups (for resolvers) ──

  def get_user_completed_at(user_id, challenge_id) do
    query =
      from(c in Completion,
        where: c.user_id == ^user_id and c.challenge_id == ^challenge_id,
        select: c.completed_at
      )

    {:ok, Repo.one(query)}
  end

  def get_user_enrolled_at(user_id, challenge_id) do
    query =
      from(e in Enrollment,
        where: e.user_id == ^user_id and e.challenge_id == ^challenge_id,
        select: e.enrolled_at
      )

    {:ok, Repo.one(query)}
  end

  # ── Private helpers ──

  defp validate_enrollment(challenge), do: validate_challenge_active(challenge)

  defp validate_challenge_active(challenge) do
    now = DateTime.utc_now()

    cond do
      challenge.published_at == nil ->
        {:error, "challenge is not published"}

      DateTime.compare(challenge.published_at, now) == :gt ->
        {:error, "challenge is not yet published"}

      challenge.end_time != nil && DateTime.compare(challenge.end_time, now) == :lt ->
        {:error, "challenge has ended"}

      true ->
        :ok
    end
  end

  defp apply_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:project_id, project_id}, q when is_binary(project_id) ->
        where(q, [c], c.project_id == ^project_id)

      {:event_id, event_id}, q when is_binary(event_id) ->
        where(q, [c], c.event_id == ^event_id)

      {:challenge_type, type}, q when is_binary(type) ->
        where(q, [c], c.challenge_type == ^type)

      {:ids, ids}, q when is_list(ids) ->
        where(q, [c], c.id in ^ids)

      {:published_after, dt}, q when not is_nil(dt) ->
        where(q, [c], c.published_at >= ^dt)

      {:published_before, dt}, q when not is_nil(dt) ->
        where(q, [c], c.published_at <= ^dt)

      _, q ->
        q
    end)
  end

  defp apply_filter(query, _), do: query

  defp apply_visibility_filter(query, _user_id, roles) when is_list(roles) do
    if has_admin_role?(roles) do
      query
    else
      apply_user_visibility_filter(query)
    end
  end

  defp apply_visibility_filter(query, _user_id, _roles) do
    apply_user_visibility_filter(query)
  end

  defp apply_user_visibility_filter(query) do
    now = DateTime.utc_now()

    where(
      query,
      [c],
      # Challenge is visible if visible_at is set and in the past
      # Or if no visible_at, fall back to published_at being in the past
      # Or if user is enrolled (checked by subquery)
      (not is_nil(c.visible_at) and c.visible_at <= ^now) or
        (is_nil(c.visible_at) and not is_nil(c.published_at) and c.published_at <= ^now) or
        c.id in subquery(
          from(e in Enrollment,
            select: e.challenge_id
          )
        )
    )
  end

  defp has_admin_role?(roles) do
    Enum.any?(roles, &(&1 in ["admin", "superadmin"]))
  end

  defp user_enrolled?(user_id, challenge_id) do
    query =
      from(e in Enrollment,
        where: e.user_id == ^user_id and e.challenge_id == ^challenge_id
      )

    Repo.exists?(query)
  end

  defp visible_to_public?(challenge) do
    now = DateTime.utc_now()

    cond do
      challenge.visible_at != nil ->
        DateTime.compare(challenge.visible_at, now) != :gt

      challenge.published_at != nil ->
        DateTime.compare(challenge.published_at, now) != :gt

      true ->
        false
    end
  end

  defp maybe_put(map, _key, nil), do: map
  defp maybe_put(map, key, value), do: Map.put(map, key, value)
end
