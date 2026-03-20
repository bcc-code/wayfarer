defmodule ElixirBackendWeb.Schema.ConsentMutations do
  use Absinthe.Schema.Notation

  alias ElixirBackend.Consents

  object :consent_mutations do
    field :accept_consent, non_null(:user_consent_history_entry) do
      arg(:consent_id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, %{consent_id: cid}, %{context: context} ->
        Consents.accept_consent(context[:current_user_id], cid)
      end)
    end

    field :reject_consent, non_null(:user_consent_history_entry) do
      arg(:consent_id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, %{consent_id: cid}, %{context: context} ->
        Consents.reject_consent(context[:current_user_id], cid)
      end)
    end

    field :create_consent, non_null(:consent) do
      arg(:input, non_null(:create_consent_input))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{input: input}, _ ->
        Consents.create_consent(input)
      end)
    end

    field :update_consent, non_null(:consent) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_consent_input))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id, input: input}, _ ->
        attrs =
          input
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Consents.update_consent(id, attrs)
      end)
    end
  end
end
