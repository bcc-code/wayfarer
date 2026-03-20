defmodule ElixirBackendWeb.Schema.QuizQueries do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Quizzes
  alias ElixirBackend.Translations

  object :quiz_queries do
    field :quiz, non_null(:quiz) do
      arg(:id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, %{id: id}, resolution ->
        Quizzes.get_quiz(id)
        |> Translations.translate_result(:quiz, resolution)
      end)
    end

    field :quizzes, non_null(:quiz_connection) do
      arg(:filter, :quiz_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, args, resolution ->
        filter = args[:filter] || %{}
        pagination = Map.take(args, [:first, :after, :last, :before])

        Quizzes.list_quizzes(filter, pagination)
        |> Translations.translate_connection(:quiz, resolution)
      end)
    end

    field :quiz_submission, non_null(:quiz_submission) do
      arg(:id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, %{id: id}, _ ->
        Quizzes.get_submission(id)
      end)
    end

    field :quiz_submissions, non_null(:quiz_submission_connection) do
      arg(:quiz_id, non_null(:id))
      arg(:user_id, :id)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, args, _ ->
        opts =
          args
          |> Map.take([:first, :after, :last, :before, :user_id])

        Quizzes.list_submissions(args.quiz_id, opts)
      end)
    end
  end
end
