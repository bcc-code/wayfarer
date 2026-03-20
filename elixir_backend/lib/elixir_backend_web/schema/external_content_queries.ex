defmodule ElixirBackendWeb.Schema.ExternalContentQueries do
  use Absinthe.Schema.Notation

  alias ElixirBackend.ExternalContent, as: EC

  object :external_content_queries do
    field :external_content, non_null(:external_content) do
      arg(:id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id}, _ ->
        EC.get_content(id)
      end)
    end

    field :external_contents, non_null(:external_content_connection) do
      arg(:filter, non_null(:external_content_filter))
      arg(:sort_by, :external_content_sort_by)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, args, _ ->
        filter = args.filter
        sort_by = args[:sort_by]
        pagination = Map.take(args, [:first, :after, :last, :before])
        EC.list_contents(filter, sort_by, pagination)
      end)
    end
  end
end
