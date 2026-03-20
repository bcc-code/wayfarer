defmodule ElixirBackend.Roles do
  @moduledoc """
  Context module for user role management.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Cache
  alias ElixirBackend.Roles.UserRole

  # ── Read ──

  def list_user_roles(user_id) do
    roles =
      Cache.fetch_raw(Cache.user_roles_key(user_id), fn ->
        from(ur in UserRole,
          where: ur.user_id == ^user_id,
          order_by: [asc: ur.assigned_at]
        )
        |> Repo.all()
      end)

    {:ok, roles}
  end

  def users_with_role(role, opts \\ []) do
    scope_type = opts[:scope_type]
    scope_id = opts[:scope_id]

    query =
      from(ur in UserRole,
        where: ur.role == ^role,
        join: u in assoc(ur, :user),
        select: u
      )

    query =
      case {scope_type, scope_id} do
        {"CHURCH", id} when is_binary(id) ->
          where(query, [ur], ur.church_id == ^id)

        {"PROJECT", id} when is_binary(id) ->
          where(query, [ur], ur.project_id == ^id)

        {"TEAM", id} when is_binary(id) ->
          where(query, [ur], ur.team_id == ^id)

        _ ->
          query
      end

    {:ok, Repo.all(query)}
  end

  # ── Write ──

  def assign_role(attrs) do
    role_attrs = scope_attrs(attrs)

    # Check if this exact role+scope already exists
    if role_exists?(role_attrs) do
      {:error, :already_assigned}
    else
      id = ULID.new_user_role_id()
      assigned_at = DateTime.utc_now() |> DateTime.truncate(:second)

      role_attrs =
        role_attrs
        |> Map.put(:id, id)
        |> Map.put(:assigned_at, assigned_at)

      result =
        %UserRole{}
        |> UserRole.changeset(role_attrs)
        |> Repo.insert()

      with {:ok, role} <- result do
        Cache.invalidate_user_roles(role_attrs.user_id)
        {:ok, role}
      end
    end
  end

  defp role_exists?(attrs) do
    query =
      from(ur in UserRole,
        where: ur.user_id == ^attrs.user_id and ur.role == ^attrs.role
      )

    query =
      cond do
        attrs[:church_id] ->
          where(query, [ur], ur.church_id == ^attrs.church_id)

        attrs[:project_id] ->
          where(query, [ur], ur.project_id == ^attrs.project_id)

        attrs[:team_id] ->
          where(query, [ur], ur.team_id == ^attrs.team_id)

        true ->
          where(
            query,
            [ur],
            is_nil(ur.church_id) and is_nil(ur.project_id) and is_nil(ur.team_id)
          )
      end

    Repo.exists?(query)
  end

  def revoke_role(attrs) do
    query =
      from(ur in UserRole,
        where: ur.user_id == ^attrs.user_id and ur.role == ^attrs.role
      )

    query = apply_scope_filter(query, attrs)

    case Repo.delete_all(query) do
      {count, _} when count > 0 ->
        Cache.invalidate_user_roles(attrs.user_id)
        {:ok, true}

      _ ->
        {:ok, false}
    end
  end

  # ── Private helpers ──

  defp scope_attrs(attrs) do
    case attrs[:scope_type] do
      "CHURCH" -> Map.put(attrs, :church_id, attrs[:scope_id])
      "PROJECT" -> Map.put(attrs, :project_id, attrs[:scope_id])
      "TEAM" -> Map.put(attrs, :team_id, attrs[:scope_id])
      _ -> attrs
    end
  end

  defp apply_scope_filter(query, attrs) do
    case attrs[:scope_type] do
      "CHURCH" ->
        where(query, [ur], ur.church_id == ^attrs[:scope_id])

      "PROJECT" ->
        where(query, [ur], ur.project_id == ^attrs[:scope_id])

      "TEAM" ->
        where(query, [ur], ur.team_id == ^attrs[:scope_id])

      _ ->
        where(
          query,
          [ur],
          is_nil(ur.church_id) and is_nil(ur.project_id) and is_nil(ur.team_id)
        )
    end
  end
end
