defmodule ElixirBackend.ExternalContent do
  @moduledoc """
  Context module for external content management.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Pagination
  alias ElixirBackend.ExternalContent.{Content, Translation}

  # ── Read ──

  def get_content(id) do
    case Repo.get(Content, id) do
      nil -> {:error, :not_found}
      content -> {:ok, content}
    end
  end

  def get_content!(id), do: Repo.get!(Content, id)

  def list_contents(filter \\ %{}, sort_by \\ nil, pagination_opts \\ %{}) do
    base_query = from(c in Content)

    query = apply_filter(base_query, filter)
    total_count = Repo.aggregate(query, :count)

    sort_field = sort_field_from(sort_by)
    sort_order = sort_order_from(sort_by)

    pagination_opts =
      Map.merge(pagination_opts, %{sort_field: sort_field, sort_order: sort_order})

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    {:ok, Pagination.build_connection(items, pagination_opts, total_count)}
  end

  def get_translations(content_id) do
    query =
      from(t in Translation,
        where: t.external_content_id == ^content_id,
        order_by: [asc: t.language_code]
      )

    Repo.all(query)
  end

  def get_title(content_id, language_code) do
    query =
      from(t in Translation,
        where: t.external_content_id == ^content_id and t.language_code == ^language_code,
        select: t.title
      )

    Repo.one(query)
  end

  # ── Write ──

  def upsert_content(attrs) do
    plan_id = attrs[:plan_id]
    task_id = attrs[:task_id]

    existing =
      if plan_id && task_id do
        Repo.one(from(c in Content, where: c.plan_id == ^plan_id and c.task_id == ^task_id))
      end

    case existing do
      nil ->
        id = attrs[:id] || ULID.new_external_content_id()
        synced_at = attrs[:synced_at] || DateTime.utc_now() |> DateTime.truncate(:second)

        %Content{}
        |> Content.changeset(Map.merge(attrs, %{id: id, synced_at: synced_at}))
        |> Repo.insert()

      content ->
        content
        |> Content.update_changeset(attrs)
        |> Repo.update()
    end
  end

  def upsert_translation(attrs) do
    content_id = attrs[:external_content_id]
    lang = attrs[:language_code]

    existing =
      if content_id && lang do
        Repo.one(
          from(t in Translation,
            where: t.external_content_id == ^content_id and t.language_code == ^lang
          )
        )
      end

    case existing do
      nil ->
        %Translation{}
        |> Translation.changeset(attrs)
        |> Repo.insert()

      _translation ->
        now = DateTime.utc_now() |> DateTime.truncate(:second)

        from(t in Translation,
          where: t.external_content_id == ^content_id and t.language_code == ^lang
        )
        |> Repo.update_all(set: [title: attrs[:title], updated_at: now])

        {:ok,
         Repo.one(
           from(t in Translation,
             where: t.external_content_id == ^content_id and t.language_code == ^lang
           )
         )}
    end
  end

  def delete_content(id) do
    with {:ok, content} <- get_content(id) do
      Repo.delete(content)
    end
  end

  # ── Private ──

  defp apply_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:plan_id, v}, q when is_binary(v) ->
        where(q, [c], c.plan_id == ^v)

      {:task_id, v}, q when is_binary(v) ->
        where(q, [c], c.task_id == ^v)

      {:content_id, v}, q when is_binary(v) ->
        where(q, [c], c.content_id == ^v)

      {:content_type, v}, q when is_binary(v) ->
        where(q, [c], c.content_type == ^v)

      {:source, v}, q when is_binary(v) ->
        where(q, [c], c.source == ^v)

      {:published_after, %DateTime{} = dt}, q ->
        where(q, [c], c.published_at >= ^dt)

      {:published_before, %DateTime{} = dt}, q ->
        where(q, [c], c.published_at <= ^dt)

      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [c], c.id in ^ids)

      _, q ->
        q
    end)
  end

  defp apply_filter(query, _), do: query

  defp sort_field_from("PUBLISHED_AT_ASC"), do: :published_at
  defp sort_field_from("PUBLISHED_AT_DESC"), do: :published_at
  defp sort_field_from(_), do: :created_at

  defp sort_order_from("CREATED_AT_DESC"), do: :desc
  defp sort_order_from("PUBLISHED_AT_DESC"), do: :desc
  defp sort_order_from("CREATED_AT_ASC"), do: :asc
  defp sort_order_from("PUBLISHED_AT_ASC"), do: :asc
  defp sort_order_from(_), do: :asc
end
