defmodule ElixirBackendWeb.Schema.WebhookMutations do
  use Absinthe.Schema.Notation

  alias ElixirBackend.Webhooks

  object :webhook_mutations do
    field :create_webhook, non_null(:webhook) do
      arg(:input, non_null(:create_webhook_input))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["superadmin"])

      resolve(fn _, %{input: input}, _ ->
        Webhooks.create_webhook(input)
      end)
    end

    field :update_webhook, non_null(:webhook) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_webhook_input))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["superadmin"])

      resolve(fn _, %{id: id, input: input}, _ ->
        attrs =
          input
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Webhooks.update_webhook(id, attrs)
      end)
    end

    field :delete_webhook, non_null(:boolean) do
      arg(:id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["superadmin"])

      resolve(fn _, %{id: id}, _ ->
        case Webhooks.delete_webhook(id) do
          {:ok, _} -> {:ok, true}
          {:error, reason} -> {:error, reason}
        end
      end)
    end
  end
end
