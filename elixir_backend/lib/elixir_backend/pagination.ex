defmodule ElixirBackend.Pagination do
  @moduledoc """
  Cursor-based pagination with configurable sort fields.

  Sort modes:
  - `:published_at` (default) — uses COALESCE(published_at, inserted_at)
  - `:created_at` — uses inserted_at directly
  """

  import Ecto.Query

  @default_page_size 10

  @doc "Encode a cursor from a record's sort key and id."
  def encode_cursor(record, sort_field \\ :published_at) do
    sort_key =
      case sort_field do
        :created_at -> Map.get(record, :inserted_at)
        _ -> Map.get(record, :published_at) || Map.get(record, :inserted_at)
      end

    %{s: sort_key && DateTime.to_iso8601(sort_key), i: record.id}
    |> Jason.encode!()
    |> Base.url_encode64(padding: false)
  end

  @doc "Decode a cursor string back to sort key + id."
  def decode_cursor(nil), do: nil

  def decode_cursor(cursor) when is_binary(cursor) do
    with {:ok, json} <- Base.url_decode64(cursor, padding: false),
         {:ok, %{"s" => sort_key_str, "i" => id}} <- Jason.decode(json) do
      {:ok, %{sort_key: parse_sort_key(sort_key_str), id: id}}
    else
      _ -> {:error, :invalid_cursor}
    end
  end

  defp parse_sort_key(nil), do: nil

  defp parse_sort_key(str) do
    case DateTime.from_iso8601(str) do
      {:ok, dt, _} -> dt
      _ -> nil
    end
  end

  @doc """
  Apply cursor pagination to a query.
  Options:
  - `:sort_field` — `:published_at` (default) or `:created_at`
  - `:first`, `:after`, `:last`, `:before` — standard cursor pagination args
  """
  def paginate(query, opts) do
    {direction, limit, cursor} = pagination_params(opts)
    sort_field = opts[:sort_field] || :published_at

    query
    |> apply_cursor(cursor, direction, sort_field)
    |> apply_order(direction, sort_field)
    |> limit(^(limit + 1))
  end

  @doc "Build a connection response from paginated results."
  def build_connection(items, opts, total_count) do
    {direction, limit, _cursor} = pagination_params(opts)
    sort_field = opts[:sort_field] || :published_at
    backward? = direction == :backward

    has_more = length(items) > limit
    items = Enum.take(items, limit)
    items = if(backward?, do: Enum.reverse(items), else: items)

    edges =
      Enum.map(items, fn item ->
        %{cursor: encode_cursor(item, sort_field), node: item}
      end)

    %{
      edges: edges,
      page_info: build_page_info(edges, has_more, backward?, opts),
      total_count: total_count
    }
  end

  defp pagination_params(opts) do
    first = opts[:first]
    last = opts[:last]

    cond do
      first -> {:forward, min(first, 100), opts[:after]}
      last -> {:backward, min(last, 100), opts[:before]}
      true -> {:forward, @default_page_size, opts[:after]}
    end
  end

  defp build_page_info(edges, has_more, backward?, opts) do
    %{
      has_next_page: if(backward?, do: opts[:before] != nil, else: has_more),
      has_previous_page: if(backward?, do: has_more, else: opts[:after] != nil),
      start_cursor: if(edges != [], do: List.first(edges).cursor),
      end_cursor: if(edges != [], do: List.last(edges).cursor)
    }
  end

  defp apply_cursor(query, nil, _direction, _sort_field), do: query

  defp apply_cursor(query, cursor_string, direction, sort_field) do
    case decode_cursor(cursor_string) do
      {:ok, %{sort_key: sort_key, id: id}} ->
        apply_cursor_where(query, sort_key, id, direction, sort_field)

      _ ->
        query
    end
  end

  # ── :created_at sort (uses inserted_at directly) ──

  defp apply_cursor_where(query, sort_key, id, :forward, :created_at) do
    if sort_key do
      where(query, [c], c.inserted_at < ^sort_key or (c.inserted_at == ^sort_key and c.id < ^id))
    else
      where(query, [c], c.id < ^id)
    end
  end

  defp apply_cursor_where(query, sort_key, id, :backward, :created_at) do
    if sort_key do
      where(query, [c], c.inserted_at > ^sort_key or (c.inserted_at == ^sort_key and c.id > ^id))
    else
      where(query, [c], c.id > ^id)
    end
  end

  # ── :published_at sort (COALESCE(published_at, inserted_at)) ──

  defp apply_cursor_where(query, sort_key, id, :forward, _sort_field) do
    if sort_key do
      where(
        query,
        [c],
        fragment("COALESCE(?, ?)", c.published_at, c.inserted_at) < ^sort_key or
          (fragment("COALESCE(?, ?)", c.published_at, c.inserted_at) == ^sort_key and c.id < ^id)
      )
    else
      where(query, [c], c.id < ^id)
    end
  end

  defp apply_cursor_where(query, sort_key, id, :backward, _sort_field) do
    if sort_key do
      where(
        query,
        [c],
        fragment("COALESCE(?, ?)", c.published_at, c.inserted_at) > ^sort_key or
          (fragment("COALESCE(?, ?)", c.published_at, c.inserted_at) == ^sort_key and c.id > ^id)
      )
    else
      where(query, [c], c.id > ^id)
    end
  end

  # ── Ordering ──

  defp apply_order(query, :forward, :created_at) do
    order_by(query, [c], desc: c.inserted_at, desc: c.id)
  end

  defp apply_order(query, :backward, :created_at) do
    order_by(query, [c], asc: c.inserted_at, asc: c.id)
  end

  defp apply_order(query, :forward, _sort_field) do
    order_by(query, [c],
      desc: fragment("COALESCE(?, ?)", c.published_at, c.inserted_at),
      desc: c.id
    )
  end

  defp apply_order(query, :backward, _sort_field) do
    order_by(query, [c],
      asc: fragment("COALESCE(?, ?)", c.published_at, c.inserted_at),
      asc: c.id
    )
  end
end
