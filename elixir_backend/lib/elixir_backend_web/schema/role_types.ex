defmodule ElixirBackendWeb.Schema.RoleTypes do
  @moduledoc "Absinthe types for user roles."
  use Absinthe.Schema.Notation

  import Absinthe.Resolution.Helpers, only: [dataloader: 1]

  # ── Enums ──

  enum :role_type do
    value(:superadmin, as: "SUPERADMIN")
    value(:admin, as: "ADMIN")
    value(:church_admin, as: "CHURCH_ADMIN")
    value(:project_admin, as: "PROJECT_ADMIN")
    value(:team_lead, as: "TEAM_LEAD")
    value(:user, as: "USER")
    value(:m2m, as: "M2M")
  end

  enum :scope_type do
    value(:church, as: "CHURCH")
    value(:project, as: "PROJECT")
    value(:team, as: "TEAM")
  end

  # ── Objects ──

  object :role_scope do
    field :type, non_null(:scope_type) do
      resolve(fn scope, _, _ -> {:ok, scope.type} end)
    end

    field :id, non_null(:id) do
      resolve(fn scope, _, _ -> {:ok, scope.id} end)
    end

    field :church, :church do
      resolve(fn scope, _, _ -> {:ok, scope[:church]} end)
    end

    field :project, :project do
      resolve(fn scope, _, _ -> {:ok, scope[:project]} end)
    end

    field :team, :team do
      resolve(fn scope, _, _ -> {:ok, scope[:team]} end)
    end
  end

  object :user_role do
    field :id, non_null(:id)
    field :role, non_null(:role_type)

    field :user, non_null(:user), resolve: dataloader(ElixirBackend.Repo)

    field :scope, :role_scope do
      resolve(fn role, _, _ ->
        cond do
          role.church_id ->
            {:ok, %{type: "CHURCH", id: role.church_id}}

          role.project_id ->
            {:ok, %{type: "PROJECT", id: role.project_id}}

          role.team_id ->
            {:ok, %{type: "TEAM", id: role.team_id}}

          true ->
            {:ok, nil}
        end
      end)
    end
  end

  # ── Input types ──

  input_object :assign_role_input do
    field :user_id, non_null(:id)
    field :role, non_null(:role_type)
    field :scope_type, :scope_type
    field :scope_id, :id
  end

  input_object :revoke_role_input do
    field :user_id, non_null(:id)
    field :role, non_null(:role_type)
    field :scope_type, :scope_type
    field :scope_id, :id
  end
end
