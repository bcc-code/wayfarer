defmodule ElixirBackendWeb.Schema.PushNotificationMutations do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.PushNotifications

  object :push_notification_mutations do
    field :register_push_subscription, non_null(:push_subscription) do
      arg(:input, non_null(:register_push_subscription_input))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, %{input: input}, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:error, "not authenticated"}
          user_id -> PushNotifications.register_subscription(user_id, input)
        end
      end)
    end

    field :unregister_push_subscription, non_null(:boolean) do
      arg(:endpoint, non_null(:string))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, %{endpoint: endpoint}, _ ->
        case PushNotifications.unregister_subscription(endpoint) do
          {:ok, _} -> {:ok, true}
          {:error, reason} -> {:error, reason}
        end
      end)
    end

    field :set_notification_preference, non_null(:push_notification_preference) do
      arg(:notification_type, non_null(:notification_type))
      arg(:enabled, non_null(:boolean))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, %{notification_type: type, enabled: enabled}, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:error, "not authenticated"}
          user_id -> PushNotifications.set_preference(user_id, type, enabled)
        end
      end)
    end
  end
end
