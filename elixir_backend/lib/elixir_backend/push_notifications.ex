defmodule ElixirBackend.PushNotifications do
  @moduledoc """
  Context for push notification subscriptions and preferences.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.PushNotifications.{PushSubscription, Preference}

  # ── Subscriptions ──

  def register_subscription(user_id, attrs) do
    id = ULID.new_push_subscription_id()

    %PushSubscription{}
    |> PushSubscription.changeset(
      attrs
      |> Map.put(:id, id)
      |> Map.put(:user_id, user_id)
    )
    |> Repo.insert(
      on_conflict: {:replace, [:p256dh_key, :auth_key, :user_agent]},
      conflict_target: [:endpoint]
    )
  end

  def unregister_subscription(endpoint) do
    case Repo.get_by(PushSubscription, endpoint: endpoint) do
      nil -> {:error, :not_found}
      sub -> Repo.delete(sub)
    end
  end

  def get_user_subscriptions(user_id) do
    from(s in PushSubscription, where: s.user_id == ^user_id)
    |> Repo.all()
  end

  # ── Preferences ──

  def get_preferences(user_id) do
    from(p in Preference, where: p.user_id == ^user_id)
    |> Repo.all()
  end

  def set_preference(user_id, notification_type, enabled) do
    case Repo.get_by(Preference, user_id: user_id, notification_type: notification_type) do
      nil ->
        %Preference{}
        |> Preference.changeset(%{
          user_id: user_id,
          notification_type: notification_type,
          enabled: enabled
        })
        |> Repo.insert()

      _pref ->
        from(p in Preference,
          where: p.user_id == ^user_id and p.notification_type == ^notification_type
        )
        |> Repo.update_all(set: [enabled: enabled])

        {:ok, Repo.get_by!(Preference, user_id: user_id, notification_type: notification_type)}
    end
  end

  def notifications_enabled?(user_id) do
    count =
      from(s in PushSubscription, where: s.user_id == ^user_id)
      |> Repo.aggregate(:count)

    count > 0
  end
end
