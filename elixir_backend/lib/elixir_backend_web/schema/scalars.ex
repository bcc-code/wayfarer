defmodule ElixirBackendWeb.Schema.Scalars do
  @moduledoc "Custom GraphQL scalar types: DateTime and HTML."
  use Absinthe.Schema.Notation

  scalar :datetime, name: "DateTime" do
    serialize(fn
      %DateTime{} = dt -> DateTime.to_iso8601(dt)
      nil -> nil
    end)

    parse(fn
      %Absinthe.Blueprint.Input.String{value: value} ->
        case DateTime.from_iso8601(value) do
          {:ok, dt, _} -> {:ok, dt}
          _ -> :error
        end

      %Absinthe.Blueprint.Input.Null{} ->
        {:ok, nil}

      _ ->
        :error
    end)
  end

  scalar :date, name: "Date" do
    serialize(fn
      %Date{} = d -> Date.to_iso8601(d)
      nil -> nil
    end)

    parse(fn
      %Absinthe.Blueprint.Input.String{value: value} ->
        case Date.from_iso8601(value) do
          {:ok, d} -> {:ok, d}
          _ -> :error
        end

      %Absinthe.Blueprint.Input.Null{} ->
        {:ok, nil}

      _ ->
        :error
    end)
  end

  scalar :json, name: "JSON" do
    serialize(fn value -> value end)

    parse(fn
      %Absinthe.Blueprint.Input.String{value: value} ->
        case Jason.decode(value) do
          {:ok, decoded} -> {:ok, decoded}
          _ -> :error
        end

      %Absinthe.Blueprint.Input.Null{} ->
        {:ok, nil}

      # Already parsed maps/lists from variables
      %{value: value} ->
        {:ok, value}

      _ ->
        :error
    end)
  end

  scalar :html, name: "HTML" do
    serialize(fn value -> value end)

    parse(fn
      %Absinthe.Blueprint.Input.String{value: value} -> {:ok, value}
      %Absinthe.Blueprint.Input.Null{} -> {:ok, nil}
      _ -> :error
    end)
  end
end
