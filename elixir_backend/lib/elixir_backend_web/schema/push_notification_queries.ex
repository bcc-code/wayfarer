defmodule ElixirBackendWeb.Schema.PushNotificationQueries do
  use Absinthe.Schema.Notation

  alias ElixirBackend.PushNotifications

  object :push_notification_queries do
    field :my_push_notification_preferences,
          non_null(list_of(non_null(:push_notification_preference))) do
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, _, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:ok, []}
          user_id -> {:ok, PushNotifications.get_preferences(user_id)}
        end
      end)
    end

    field :push_notifications_enabled, non_null(:boolean) do
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, _, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:ok, false}
          user_id -> {:ok, PushNotifications.notifications_enabled?(user_id)}
        end
      end)
    end

    field :vapid_public_key, :string do
      resolve(fn _, _, _ ->
        {:ok, Application.get_env(:elixir_backend, :vapid_public_key)}
      end)
    end
  end
end
