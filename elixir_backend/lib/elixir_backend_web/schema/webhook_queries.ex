defmodule ElixirBackendWeb.Schema.WebhookQueries do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Webhooks

  object :webhook_queries do
    field :webhook, non_null(:webhook) do
      arg(:id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["superadmin"])

      resolve(fn _, %{id: id}, _ ->
        Webhooks.get_webhook(id)
      end)
    end

    field :webhooks, non_null(list_of(non_null(:webhook))) do
      arg(:project_id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["superadmin"])

      resolve(fn _, %{project_id: pid}, _ ->
        {:ok, Webhooks.list_webhooks(pid)}
      end)
    end
  end
end
