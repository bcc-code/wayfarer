defmodule ElixirBackendWeb.Schema.QuizMutations do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Quizzes

  object :quiz_mutations do
    # ── Quiz CRUD ──

    field :create_quiz, non_null(:quiz) do
      arg(:input, non_null(:create_quiz_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{input: input}, _ ->
        attrs = Map.put(input, :image_url, input[:image])
        Quizzes.create_quiz(attrs)
      end)
    end

    field :update_quiz, non_null(:quiz) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_quiz_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id, input: input}, _ ->
        attrs =
          input
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()
          |> then(fn a ->
            if Map.has_key?(a, :image), do: Map.put(a, :image_url, a[:image]), else: a
          end)

        Quizzes.update_quiz(id, attrs)
      end)
    end

    field :delete_quiz, non_null(:boolean) do
      arg(:id, non_null(:id))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id}, _ ->
        case Quizzes.delete_quiz(id) do
          {:ok, _} -> {:ok, true}
          {:error, reason} -> {:error, reason}
        end
      end)
    end

    # ── Question management ──

    field :add_quiz_question, non_null(:quiz_question) do
      arg(:quiz_id, non_null(:id))
      arg(:input, non_null(:create_quiz_question_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{quiz_id: qid, input: input}, _ ->
        Quizzes.add_question(qid, input)
      end)
    end

    field :update_quiz_question, non_null(:quiz_question) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_quiz_question_input))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id, input: input}, _ ->
        attrs =
          input
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Quizzes.update_question(id, attrs)
      end)
    end

    field :delete_quiz_question, non_null(:boolean) do
      arg(:id, non_null(:id))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id}, _ ->
        case Quizzes.delete_question(id) do
          {:ok, _} -> {:ok, true}
          {:error, reason} -> {:error, reason}
        end
      end)
    end

    field :reorder_quiz_questions, non_null(list_of(non_null(:quiz_question))) do
      arg(:quiz_id, non_null(:id))
      arg(:question_ids, non_null(list_of(non_null(:id))))
      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{quiz_id: qid, question_ids: ids}, _ ->
        Quizzes.reorder_questions(qid, ids)
      end)
    end

    # ── Taking quizzes ──

    field :submit_quiz_answer, non_null(:quiz_response) do
      arg(:submission_id, non_null(:id))
      arg(:input, non_null(:submit_quiz_answer_input))

      resolve(fn _, %{submission_id: sid, input: input}, _ ->
        Quizzes.submit_answer(sid, input)
      end)
    end

    field :finalize_quiz, non_null(:quiz_submission) do
      arg(:submission_id, non_null(:id))

      resolve(fn _, %{submission_id: sid}, _ ->
        Quizzes.finalize_submission(sid)
      end)
    end

    # ── M2M: External submission ──

    field :create_quiz_submission, non_null(:quiz_submission) do
      arg(:quiz_id, non_null(:id))
      arg(:user_id, non_null(:id))
      arg(:responses, non_null(list_of(non_null(:submit_quiz_answer_input))))
      arg(:completed_at, :datetime)

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["m2m", "admin", "superadmin"]
      )

      resolve(fn _, %{quiz_id: qid, user_id: uid, responses: responses} = args, _ ->
        with {:ok, submission} <- Quizzes.create_submission(qid, uid) do
          Enum.each(responses, fn resp ->
            Quizzes.submit_answer(submission.id, resp)
          end)

          if args[:completed_at] do
            Quizzes.finalize_submission(submission.id)
          else
            {:ok, submission}
          end
        end
      end)
    end
  end
end
