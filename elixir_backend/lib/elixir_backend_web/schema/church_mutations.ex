defmodule ElixirBackendWeb.Schema.ChurchMutations do
  @moduledoc "GraphQL mutation resolvers for churches."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Churches
  alias ElixirBackendWeb.Schema.Middleware.RequireRole

  object :church_mutations do
    field :update_church, non_null(:church) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_church_input))

      middleware(RequireRole, roles: ["superadmin"])

      resolve(fn _parent, %{id: id, input: input}, _resolution ->
        Churches.update_church(id, input)
      end)
    end
  end
end
