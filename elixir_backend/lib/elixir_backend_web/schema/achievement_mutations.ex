defmodule ElixirBackendWeb.Schema.AchievementMutations do
  use Absinthe.Schema.Notation

  alias ElixirBackend.Achievements

  object :achievement_mutations do
    # ── Create ──

    field :create_simple_achievement, non_null(:simple_achievement) do
      arg(:input, non_null(:create_simple_achievement_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{input: input}, _ ->
        Achievements.create_simple_achievement(input)
      end)
    end

    field :create_content_achievement, non_null(:content_achievement) do
      arg(:input, non_null(:create_content_achievement_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{input: input}, _ ->
        Achievements.create_content_achievement(input)
      end)
    end

    field :create_streak_achievement, non_null(:streak_achievement) do
      arg(:input, non_null(:create_streak_achievement_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{input: input}, _ ->
        Achievements.create_streak_achievement(input)
      end)
    end

    # ── Update ──

    field :update_achievement, non_null(:achievement) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_achievement_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id, input: input}, _ ->
        attrs =
          input
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Achievements.update_achievement(id, attrs)
      end)
    end

    field :update_content_achievement, non_null(:content_achievement) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_content_achievement_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id, input: input}, _ ->
        attrs =
          input
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Achievements.update_content_achievement(id, attrs)
      end)
    end

    field :update_streak_achievement, non_null(:streak_achievement) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_streak_achievement_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id, input: input}, _ ->
        attrs =
          input
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Achievements.update_streak_achievement(id, attrs)
      end)
    end

    field :update_quiz_achievement, non_null(:quiz_achievement) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_quiz_achievement_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id, input: input}, _ ->
        attrs =
          input
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Achievements.update_quiz_achievement(id, attrs)
      end)
    end

    # ── Delete ──

    field :delete_achievement, non_null(:boolean) do
      arg(:id, non_null(:id))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id}, _ ->
        case Achievements.delete_achievement(id) do
          {:ok, _} -> {:ok, true}
          {:error, reason} -> {:error, reason}
        end
      end)
    end

    field :link_achievement_to_challenge, non_null(:achievement) do
      arg(:achievement_id, non_null(:id))
      arg(:challenge_id, non_null(:id))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{achievement_id: aid, challenge_id: cid}, _ ->
        Achievements.link_to_challenge(aid, cid)
      end)
    end

    # ── Reorder ──

    field :reorder_achievements, non_null(list_of(non_null(:achievement))) do
      arg(:project_id, non_null(:id))
      arg(:achievement_ids, non_null(list_of(non_null(:id))))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{project_id: pid, achievement_ids: ids}, _ ->
        Achievements.reorder_achievements(pid, ids)
      end)
    end

    # ── Award/Revoke ──

    field :award_achievement, non_null(:achievement) do
      arg(:user_id, non_null(:id))
      arg(:achievement_id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["m2m", "admin", "superadmin"]
      )

      resolve(fn _, %{user_id: uid, achievement_id: aid}, _ ->
        Achievements.award_achievement(uid, aid)
      end)
    end

    field :revoke_achievement, non_null(:boolean) do
      arg(:user_id, non_null(:id))
      arg(:achievement_id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["m2m", "admin", "superadmin"]
      )

      resolve(fn _, %{user_id: uid, achievement_id: aid}, _ ->
        Achievements.revoke_achievement(uid, aid)
      end)
    end

    # ── Content Progress ──

    field :mark_content_item_completed, non_null(list_of(non_null(:content_achievement))) do
      arg(:user_id, non_null(:id))
      arg(:external_content_id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["m2m", "admin", "superadmin"]
      )

      resolve(fn _, %{user_id: uid, external_content_id: ecid}, _ ->
        Achievements.mark_content_completed(uid, ecid)
      end)
    end

    field :unmark_content_item_completed, non_null(list_of(non_null(:content_achievement))) do
      arg(:user_id, non_null(:id))
      arg(:external_content_id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["m2m", "admin", "superadmin"]
      )

      resolve(fn _, %{user_id: uid, external_content_id: ecid}, _ ->
        Achievements.unmark_content_completed(uid, ecid)
      end)
    end

    # ── Celebrated ──

    field :mark_achievement_celebrated, non_null(:boolean) do
      arg(:achievement_id, non_null(:id))

      resolve(fn _, %{achievement_id: aid}, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:error, "unauthorized"}
          user_id -> Achievements.mark_celebrated(user_id, aid)
        end
      end)
    end
  end
end
