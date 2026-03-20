defmodule ElixirBackendWeb.Schema.ConsentQueries do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Consents

  object :consent_queries do
    field :consents, non_null(list_of(non_null(:consent))) do
      resolve(fn _, _, _ ->
        {:ok, Consents.list_consents()}
      end)
    end

    field :consent, non_null(:consent) do
      arg(:id, non_null(:id))

      resolve(fn _, %{id: id}, _ ->
        Consents.get_consent(id)
      end)
    end

    field :pending_consents, non_null(list_of(non_null(:consent))) do
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, _, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:ok, []}
          user_id -> {:ok, Consents.pending_consents(user_id)}
        end
      end)
    end
  end
end
