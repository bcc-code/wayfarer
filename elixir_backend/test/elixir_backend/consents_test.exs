defmodule ElixirBackend.ConsentsTest do
  use ElixirBackend.DataCase, async: true

  alias ElixirBackend.Consents
  alias ElixirBackend.TestHelpers

  describe "create_consent/1" do
    test "creates a consent with valid attrs" do
      {:ok, consent} =
        Consents.create_consent(%{
          key: "terms-of-service",
          title: "Terms of Service",
          body: "You agree to our terms.",
          published_at: ~U[2026-01-01 00:00:00Z]
        })

      assert consent.key == "terms-of-service"
      assert consent.title == "Terms of Service"
      assert consent.version == 1
    end

    test "enforces unique key+version" do
      {:ok, _} =
        Consents.create_consent(%{
          key: "privacy",
          title: "Privacy Policy",
          version: 1
        })

      {:error, changeset} =
        Consents.create_consent(%{
          key: "privacy",
          title: "Privacy v2",
          version: 1
        })

      assert errors_on(changeset)[:key] || errors_on(changeset)[:version] ||
               changeset.errors != []
    end
  end

  describe "list_consents/0" do
    test "returns only published consents" do
      {:ok, _published} =
        Consents.create_consent(%{
          key: "published-one",
          title: "Published",
          published_at: ~U[2026-01-01 00:00:00Z]
        })

      {:ok, _draft} =
        Consents.create_consent(%{
          key: "draft-one",
          title: "Draft"
        })

      consents = Consents.list_consents()
      assert length(consents) >= 1
      assert Enum.all?(consents, fn c -> c.published_at != nil end)
    end
  end

  describe "get_consent/1" do
    test "returns consent by id" do
      {:ok, consent} =
        Consents.create_consent(%{key: "test-get", title: "Test"})

      {:ok, found} = Consents.get_consent(consent.id)
      assert found.id == consent.id
    end

    test "returns error for missing consent" do
      assert {:error, :not_found} = Consents.get_consent("CN00000000000000000000000000")
    end
  end

  describe "update_consent/2" do
    test "updates consent fields" do
      {:ok, consent} =
        Consents.create_consent(%{key: "update-test", title: "Old Title"})

      {:ok, updated} = Consents.update_consent(consent.id, %{title: "New Title"})
      assert updated.title == "New Title"
    end
  end

  describe "accept_consent/2 and reject_consent/2" do
    test "records consent acceptance" do
      user = TestHelpers.create_user()

      {:ok, consent} =
        Consents.create_consent(%{
          key: "accept-test",
          title: "Accept Test",
          published_at: ~U[2026-01-01 00:00:00Z]
        })

      {:ok, entry} = Consents.accept_consent(user.id, consent.id)
      assert entry.action == "ACCEPTED"
      assert entry.consent_key == "accept-test"
    end

    test "records consent rejection" do
      user = TestHelpers.create_user()

      {:ok, consent} =
        Consents.create_consent(%{
          key: "reject-test",
          title: "Reject Test",
          published_at: ~U[2026-01-01 00:00:00Z]
        })

      {:ok, entry} = Consents.reject_consent(user.id, consent.id)
      assert entry.action == "REJECTED"
    end
  end

  describe "pending_consents/1" do
    test "returns consents user has not acted on" do
      user = TestHelpers.create_user()

      {:ok, c1} =
        Consents.create_consent(%{
          key: "pending-1",
          title: "Pending 1",
          published_at: ~U[2026-01-01 00:00:00Z]
        })

      {:ok, _c2} =
        Consents.create_consent(%{
          key: "pending-2",
          title: "Pending 2",
          published_at: ~U[2026-01-01 00:00:00Z]
        })

      Consents.accept_consent(user.id, c1.id)

      pending = Consents.pending_consents(user.id)
      pending_keys = Enum.map(pending, & &1.key)
      assert "pending-2" in pending_keys
      refute "pending-1" in pending_keys
    end
  end
end
