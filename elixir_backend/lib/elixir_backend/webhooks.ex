defmodule ElixirBackend.Webhooks do
  @moduledoc """
  Context for webhook management and logging.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Webhooks.{Webhook, WebhookLog}

  # ── CRUD ──

  def get_webhook(id) do
    case Repo.get(Webhook, id) do
      nil -> {:error, :not_found}
      wh -> {:ok, wh}
    end
  end

  def list_webhooks(project_id) do
    from(w in Webhook, where: w.project_id == ^project_id, order_by: [asc: w.name])
    |> Repo.all()
  end

  def create_webhook(attrs) do
    id = ULID.new_webhook_id()

    %Webhook{}
    |> Webhook.changeset(Map.put(attrs, :id, id))
    |> Repo.insert()
  end

  def update_webhook(id, attrs) do
    with {:ok, webhook} <- get_webhook(id) do
      webhook
      |> Webhook.update_changeset(attrs)
      |> Repo.update()
    end
  end

  def delete_webhook(id) do
    with {:ok, webhook} <- get_webhook(id) do
      Repo.delete(webhook)
    end
  end

  # ── Logs ──

  def get_recent_logs(webhook_id, limit \\ 10) do
    from(l in WebhookLog,
      where: l.webhook_id == ^webhook_id,
      order_by: [desc: l.created_at],
      limit: ^limit
    )
    |> Repo.all()
  end

  def create_log(attrs) do
    id = ULID.new_webhook_log_id()
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    %WebhookLog{}
    |> WebhookLog.changeset(
      attrs
      |> Map.put(:id, id)
      |> Map.put(:created_at, now)
    )
    |> Repo.insert()
  end
end
