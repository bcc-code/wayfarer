defmodule ElixirBackendWeb.Schema.AdminQueries do
  use Absinthe.Schema.Notation

  alias ElixirBackend.Admin

  object :admin_queries do
    field :admin_dashboard_stats, non_null(:admin_dashboard_stats) do
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, _, _ ->
        Admin.dashboard_stats()
      end)
    end

    field :church_admin_statistics, non_null(:church_admin_statistics) do
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["church_admin", "admin", "superadmin"]
      )

      resolve(fn _, _, %{context: context} ->
        Admin.church_admin_statistics(context[:current_user_id])
      end)
    end
  end
end
