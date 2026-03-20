defmodule ElixirBackend.WebhooksTest do
  use ElixirBackend.DataCase, async: true

  alias ElixirBackend.Webhooks
  import ElixirBackend.TestHelpers

  setup do
    project = create_project()
    %{project: project}
  end

  describe "create_webhook/1" do
    test "creates a webhook", %{project: project} do
      {:ok, wh} =
        Webhooks.create_webhook(%{
          project_id: project.id,
          name: "Test Hook",
          url: "https://example.com/hook",
          event_type: "CHALLENGE_COMPLETED"
        })

      assert wh.name == "Test Hook"
      assert String.starts_with?(wh.id, "WH")
    end
  end

  describe "get_webhook/1" do
    test "returns webhook by id", %{project: project} do
      {:ok, wh} =
        Webhooks.create_webhook(%{
          project_id: project.id,
          name: "Hook",
          url: "https://example.com/hook",
          event_type: "CUSTOM"
        })

      assert {:ok, found} = Webhooks.get_webhook(wh.id)
      assert found.id == wh.id
    end
  end

  describe "list_webhooks/1" do
    test "lists webhooks for project", %{project: project} do
      {:ok, _} =
        Webhooks.create_webhook(%{
          project_id: project.id,
          name: "Hook1",
          url: "https://example.com/1",
          event_type: "CUSTOM"
        })

      hooks = Webhooks.list_webhooks(project.id)
      assert length(hooks) == 1
    end
  end

  describe "update_webhook/2" do
    test "updates webhook fields", %{project: project} do
      {:ok, wh} =
        Webhooks.create_webhook(%{
          project_id: project.id,
          name: "Old",
          url: "https://example.com",
          event_type: "CUSTOM"
        })

      {:ok, updated} = Webhooks.update_webhook(wh.id, %{name: "New", active: false})
      assert updated.name == "New"
      assert updated.active == false
    end
  end

  describe "delete_webhook/1" do
    test "deletes a webhook", %{project: project} do
      {:ok, wh} =
        Webhooks.create_webhook(%{
          project_id: project.id,
          name: "Del",
          url: "https://example.com",
          event_type: "CUSTOM"
        })

      assert {:ok, _} = Webhooks.delete_webhook(wh.id)
      assert {:error, :not_found} = Webhooks.get_webhook(wh.id)
    end
  end

  describe "create_log/1" do
    test "creates a webhook log", %{project: project} do
      {:ok, wh} =
        Webhooks.create_webhook(%{
          project_id: project.id,
          name: "Log Test",
          url: "https://example.com",
          event_type: "CUSTOM"
        })

      {:ok, log} =
        Webhooks.create_log(%{
          webhook_id: wh.id,
          event_type: "CUSTOM",
          request_payload: %{"data" => "test"},
          response_status_code: 200,
          duration_ms: 150
        })

      assert log.response_status_code == 200
      assert String.starts_with?(log.id, "WL")

      logs = Webhooks.get_recent_logs(wh.id)
      assert length(logs) == 1
    end
  end
end
