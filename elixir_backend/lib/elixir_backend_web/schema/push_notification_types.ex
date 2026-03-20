defmodule ElixirBackendWeb.Schema.PushNotificationTypes do
  use Absinthe.Schema.Notation
  @moduledoc false

  enum :notification_type do
    value(:achievement_unlocked, as: "ACHIEVEMENT_UNLOCKED")
    value(:challenge_available, as: "CHALLENGE_AVAILABLE")
    value(:generic, as: "GENERIC")
  end

  object :push_subscription do
    field :id, non_null(:id)
    field :endpoint, non_null(:string)
    field :user_agent, :string
  end

  object :push_notification_preference do
    field :notification_type, non_null(:notification_type)
    field :enabled, non_null(:boolean)
  end

  input_object :register_push_subscription_input do
    field :endpoint, non_null(:string)
    field :p256dh_key, non_null(:string)
    field :auth_key, non_null(:string)
    field :user_agent, :string
  end
end
