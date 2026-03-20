defmodule ElixirBackend.Pagination do
  @moduledoc """
  Cursor-based pagination using composite cursors (published_at, id).
  Encodes/decodes cursors as Base64-encoded JSON.
  """

  import Ecto.Query

  @default_page_size 10

  @doc "Encode a cursor from a challenge's sort key and id."
  def encode_cursor(%{published_at: published_at, id: id}) do
    sort_key = published_at || nil

    %{s: sort_key && DateTime.to_iso8601(sort_key), i: id}
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
  Uses COALESCE(published_at, created_at) as the sort key.
  """
  def paginate(query, opts) do
    {direction, limit, cursor} = pagination_params(opts)

    query
    |> apply_cursor(cursor, direction)
    |> apply_order(direction)
    |> limit(^(limit + 1))
  end

  @doc "Build a connection response from paginated results."
  def build_connection(items, opts, total_count) do
    {direction, limit, _cursor} = pagination_params(opts)
    backward? = direction == :backward

    has_more = length(items) > limit
    items = Enum.take(items, limit)
    items = if(backward?, do: Enum.reverse(items), else: items)

    edges = Enum.map(items, fn item -> %{cursor: encode_cursor(item), node: item} end)

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

  defp apply_cursor(query, nil, _direction), do: query

  defp apply_cursor(query, cursor_string, direction) do
    case decode_cursor(cursor_string) do
      {:ok, %{sort_key: sort_key, id: id}} ->
        apply_cursor_where(query, sort_key, id, direction)

      _ ->
        query
    end
  end

  defp apply_cursor_where(query, sort_key, id, :forward) do
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

  defp apply_cursor_where(query, sort_key, id, :backward) do
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

  defp apply_order(query, :forward) do
    order_by(query, [c],
      desc: fragment("COALESCE(?, ?)", c.published_at, c.inserted_at),
      desc: c.id
    )
  end

  defp apply_order(query, :backward) do
    order_by(query, [c],
      asc: fragment("COALESCE(?, ?)", c.published_at, c.inserted_at),
      asc: c.id
    )
  end
end
