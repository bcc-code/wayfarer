defmodule ElixirBackend.PushNotificationsTest do
  use ElixirBackend.DataCase, async: true

  alias ElixirBackend.PushNotifications
  import ElixirBackend.TestHelpers

  setup do
    user = create_user()
    %{user: user}
  end

  describe "register_subscription/2" do
    test "registers a push subscription", %{user: user} do
      {:ok, sub} =
        PushNotifications.register_subscription(user.id, %{
          endpoint: "https://push.example.com/sub1",
          p256dh_key: "test-p256dh",
          auth_key: "test-auth"
        })

      assert sub.endpoint == "https://push.example.com/sub1"
      assert String.starts_with?(sub.id, "PS")
    end

    test "upserts on same endpoint", %{user: user} do
      attrs = %{
        endpoint: "https://push.example.com/sub2",
        p256dh_key: "key1",
        auth_key: "auth1"
      }

      {:ok, _} = PushNotifications.register_subscription(user.id, attrs)

      {:ok, sub2} =
        PushNotifications.register_subscription(user.id, %{attrs | p256dh_key: "key2"})

      assert sub2.p256dh_key == "key2"
    end
  end

  describe "unregister_subscription/1" do
    test "removes subscription by endpoint", %{user: user} do
      {:ok, _} =
        PushNotifications.register_subscription(user.id, %{
          endpoint: "https://push.example.com/del",
          p256dh_key: "key",
          auth_key: "auth"
        })

      assert {:ok, _} = PushNotifications.unregister_subscription("https://push.example.com/del")
      assert PushNotifications.get_user_subscriptions(user.id) == []
    end
  end

  describe "preferences" do
    test "set and get preferences", %{user: user} do
      {:ok, pref} = PushNotifications.set_preference(user.id, "ACHIEVEMENT_UNLOCKED", false)
      assert pref.enabled == false

      prefs = PushNotifications.get_preferences(user.id)
      assert length(prefs) == 1
    end

    test "updates existing preference", %{user: user} do
      {:ok, _} = PushNotifications.set_preference(user.id, "GENERIC", true)
      {:ok, updated} = PushNotifications.set_preference(user.id, "GENERIC", false)
      assert updated.enabled == false
    end
  end

  describe "notifications_enabled?/1" do
    test "returns false when no subscriptions", %{user: user} do
      assert PushNotifications.notifications_enabled?(user.id) == false
    end

    test "returns true with subscription", %{user: user} do
      {:ok, _} =
        PushNotifications.register_subscription(user.id, %{
          endpoint: "https://push.example.com/check",
          p256dh_key: "key",
          auth_key: "auth"
        })

      assert PushNotifications.notifications_enabled?(user.id) == true
    end
  end
end
