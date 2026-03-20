defmodule ElixirBackendWeb.Schema.FeedbackQueries do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Feedback

  object :feedback_queries do
    field :feedback, non_null(:feedback_connection) do
      arg(:filter, :feedback_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, args, _ ->
        filter = args[:filter] || %{}
        pagination = Map.take(args, [:first, :after, :last, :before])
        Feedback.list_feedback(filter, pagination)
      end)
    end

    field :feedback_tags, non_null(list_of(non_null(:string))) do
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, _, _ ->
        {:ok, Feedback.get_tags()}
      end)
    end

    field :feedback_platforms, non_null(list_of(non_null(:string))) do
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, _, _ ->
        {:ok, Feedback.get_platforms()}
      end)
    end
  end
end
