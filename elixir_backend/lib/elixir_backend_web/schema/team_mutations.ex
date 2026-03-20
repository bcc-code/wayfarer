defmodule ElixirBackendWeb.Schema.TeamMutations do
  @moduledoc "GraphQL mutation resolvers for teams and super teams."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Teams
  alias ElixirBackendWeb.Schema.Middleware.RequireRole

  object :team_mutations do
    # ── Team mutations ──

    field :join_team, non_null(:team) do
      arg(:code, non_null(:id))

      resolve(fn _parent, %{code: code}, %{context: context} ->
        case context do
          %{current_user_id: user_id} ->
            Teams.join_team(user_id, code)

          _ ->
            {:error, "authentication required"}
        end
      end)
    end

    field :create_team, non_null(:team) do
      arg(:project_id, non_null(:id))
      arg(:input, non_null(:create_team_input))

      middleware(RequireRole, roles: ["admin", "superadmin", "church_admin"])

      resolve(fn _parent, %{project_id: project_id, input: input}, _resolution ->
        Teams.create_team(project_id, input)
      end)
    end

    field :update_team, non_null(:team) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_team_input))

      middleware(RequireRole, roles: ["admin", "superadmin", "church_admin", "team_lead"])

      resolve(fn _parent, %{id: id, input: input}, _resolution ->
        Teams.update_team(id, input)
      end)
    end

    field :delete_team, non_null(:boolean) do
      arg(:id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin", "church_admin"])

      resolve(fn _parent, %{id: id}, _resolution ->
        case Teams.delete_team(id) do
          {:ok, _} -> {:ok, true}
          {:error, _} -> {:error, "failed to delete team"}
        end
      end)
    end

    field :add_team_members, non_null(:team) do
      arg(:team_id, non_null(:id))
      arg(:user_ids, non_null(list_of(non_null(:id))))
      arg(:force, :boolean)

      middleware(RequireRole, roles: ["admin", "superadmin", "church_admin"])

      resolve(fn _parent, %{team_id: team_id, user_ids: user_ids}, _resolution ->
        Teams.add_members(team_id, user_ids)
      end)
    end

    field :remove_team_members, non_null(:team) do
      arg(:team_id, non_null(:id))
      arg(:user_ids, non_null(list_of(non_null(:id))))

      middleware(RequireRole, roles: ["admin", "superadmin", "church_admin"])

      resolve(fn _parent, %{team_id: team_id, user_ids: user_ids}, _resolution ->
        Teams.remove_members(team_id, user_ids)
      end)
    end

    field :regenerate_join_code, non_null(:team) do
      arg(:team_id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin", "church_admin"])

      resolve(fn _parent, %{team_id: team_id}, _resolution ->
        Teams.regenerate_join_code(team_id)
      end)
    end

    field :assign_team_lead, non_null(:team) do
      arg(:team_id, non_null(:id))
      arg(:user_id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin", "church_admin", "team_lead"])

      resolve(fn _parent, %{team_id: team_id, user_id: user_id}, _resolution ->
        Teams.assign_team_lead(team_id, user_id)
      end)
    end

    # ── SuperTeam mutations ──

    field :create_super_team, non_null(:super_team) do
      arg(:project_id, non_null(:id))
      arg(:input, non_null(:create_super_team_input))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{project_id: project_id, input: input}, _resolution ->
        Teams.create_super_team(project_id, input)
      end)
    end

    field :update_super_team, non_null(:super_team) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_super_team_input))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id, input: input}, _resolution ->
        Teams.update_super_team(id, input)
      end)
    end

    field :delete_super_team, non_null(:boolean) do
      arg(:id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id}, _resolution ->
        case Teams.delete_super_team(id) do
          {:ok, _} -> {:ok, true}
          {:error, _} -> {:error, "failed to delete super team"}
        end
      end)
    end

    field :assign_teams_to_super_team, non_null(:super_team) do
      arg(:super_team_id, non_null(:id))
      arg(:team_ids, non_null(list_of(non_null(:id))))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{super_team_id: super_team_id, team_ids: team_ids}, _resolution ->
        Teams.assign_teams_to_super_team(super_team_id, team_ids)
      end)
    end
  end
end
