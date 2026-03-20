defmodule ElixirBackendWeb.Schema.AdminMutations do
  use Absinthe.Schema.Notation

  object :admin_mutations do
    field :clear_all_cache, non_null(:boolean) do
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, _, _ ->
        # Cache clearing implementation can be added later
        {:ok, true}
      end)
    end
  end
end
