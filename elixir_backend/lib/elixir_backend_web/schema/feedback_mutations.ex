defmodule ElixirBackendWeb.Schema.FeedbackMutations do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Feedback

  object :feedback_mutations do
    field :submit_feedback, non_null(:user_feedback) do
      arg(:input, non_null(:submit_feedback_input))

      resolve(fn _, %{input: input}, %{context: context} ->
        attrs = Map.put(input, :user_id, context[:current_user_id])
        Feedback.submit_feedback(attrs)
      end)
    end

    field :delete_feedback, non_null(:boolean) do
      arg(:id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id}, _ ->
        case Feedback.delete_feedback(id) do
          {:ok, _} -> {:ok, true}
          {:error, reason} -> {:error, reason}
        end
      end)
    end

    field :mark_feedback_handled, non_null(:user_feedback) do
      arg(:feedback_id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{feedback_id: id}, _ ->
        Feedback.mark_handled(id)
      end)
    end

    field :update_feedback_tags, non_null(:user_feedback) do
      arg(:feedback_id, non_null(:id))
      arg(:tags, non_null(list_of(non_null(:string))))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{feedback_id: id, tags: tags}, _ ->
        Feedback.update_tags(id, tags)
      end)
    end
  end
end
