defmodule ElixirBackend.Translations do
  @moduledoc """
  Central context module for entity translations.
  Provides generic CRUD operations and translation application for all translatable entities.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.Cache

  alias ElixirBackend.Projects.ProjectTranslation
  alias ElixirBackend.Events.EventTranslation
  alias ElixirBackend.Streaks.StreakTranslation
  alias ElixirBackend.Challenges.ChallengeTranslation
  alias ElixirBackend.Achievements.AchievementTranslation
  alias ElixirBackend.Consents.ConsentTranslation

  alias ElixirBackend.Quizzes.{
    QuizTranslation,
    QuizQuestionTranslation,
    QuizPredefinedAnswerTranslation
  }

  @default_language "no"

  @entity_config %{
    project: {ProjectTranslation, :project_id, [:name, :description, :rules]},
    event: {EventTranslation, :event_id, [:name, :description]},
    streak: {StreakTranslation, :streak_id, [:name, :description]},
    challenge:
      {ChallengeTranslation, :challenge_id,
       [:name, :description, :button_text, :notification_text]},
    achievement:
      {AchievementTranslation, :achievement_id,
       [:name, :description_pending, :description_completed, :notification_text]},
    consent: {ConsentTranslation, :consent_id, [:title, :short_text, :body]},
    quiz: {QuizTranslation, :quiz_id, [:name, :description]},
    quiz_question: {QuizQuestionTranslation, :question_id, [:question_text]},
    quiz_answer: {QuizPredefinedAnswerTranslation, :answer_id, [:answer_text]}
  }

  @doc "Returns the default language code."
  def default_language, do: @default_language

  @doc "Returns the entity config for a given entity type."
  def entity_config(entity_type), do: Map.fetch!(@entity_config, entity_type)

  @doc "Get a translation for a given entity type, entity ID, and language code."
  def get_translation(entity_type, entity_id, language_code) do
    {schema, fk_field, _fields} = entity_config(entity_type)

    Cache.fetch(Cache.translation_key(entity_type, entity_id, language_code), fn ->
      query =
        from(t in schema,
          where: field(t, ^fk_field) == ^entity_id and t.language_code == ^language_code
        )

      case Repo.one(query) do
        nil -> {:error, :not_found}
        translation -> {:ok, translation}
      end
    end)
  end

  @doc "Upsert a translation (insert or update on conflict)."
  def upsert_translation(entity_type, attrs) do
    {schema, fk_field, fields} = entity_config(entity_type)
    entity_id = attrs[fk_field]
    language_code = attrs[:language_code]

    existing =
      if entity_id && language_code do
        Repo.one(
          from(t in schema,
            where: field(t, ^fk_field) == ^entity_id and t.language_code == ^language_code
          )
        )
      end

    result =
      case existing do
        nil ->
          struct(schema)
          |> schema.changeset(attrs)
          |> Repo.insert()

        translation ->
          translation
          |> schema.changeset(Map.take(attrs, fields))
          |> Repo.update()
      end

    with {:ok, translation} <- result do
      Cache.del(Cache.translation_key(entity_type, entity_id, language_code))
      {:ok, translation}
    end
  end

  @doc "Delete a translation."
  def delete_translation(entity_type, entity_id, language_code) do
    {schema, fk_field, _fields} = entity_config(entity_type)

    query =
      from(t in schema,
        where: field(t, ^fk_field) == ^entity_id and t.language_code == ^language_code
      )

    case Repo.delete_all(query) do
      {count, _} when count > 0 ->
        Cache.del(Cache.translation_key(entity_type, entity_id, language_code))
        :ok

      _ ->
        {:error, :not_found}
    end
  end

  @doc """
  Returns the translation status for an entity — which languages have translations
  and which fields are populated for each.
  """
  def translation_status(entity_type, entity_id) do
    {schema, fk_field, fields} = entity_config(entity_type)

    translations =
      from(t in schema, where: field(t, ^fk_field) == ^entity_id)
      |> Repo.all()

    Enum.map(translations, fn translation ->
      populated_fields =
        fields
        |> Enum.filter(fn f ->
          value = Map.get(translation, f)
          value != nil && value != ""
        end)
        |> Enum.map(&Atom.to_string/1)

      %{language_code: translation.language_code, fields: populated_fields}
    end)
  end

  @doc """
  Apply translation to an entity struct. Overlays non-nil, non-empty translated fields
  onto the base entity. Skips if language is the default language.
  """
  def apply_translation(entity, _entity_type, @default_language), do: entity
  def apply_translation(entity, _entity_type, nil), do: entity

  def apply_translation(entity, entity_type, language_code) do
    {_schema, fk_field, fields} = entity_config(entity_type)
    entity_id = Map.get(entity, :id) || Map.get(entity, fk_field)

    case get_translation(entity_type, entity_id, language_code) do
      {:ok, translation} -> overlay_fields(entity, translation, fields)
      _ -> entity
    end
  end

  defp overlay_fields(entity, translation, fields) do
    Enum.reduce(fields, entity, fn field, acc ->
      case Map.get(translation, field) do
        value when is_binary(value) and value != "" -> Map.put(acc, field, value)
        _ -> acc
      end
    end)
  end
end
