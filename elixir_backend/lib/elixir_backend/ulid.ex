defmodule ElixirBackend.ULID do
  @moduledoc """
  Generates prefixed ULIDs matching the Go backend format.
  Format: 2-char prefix + 26-char Crockford Base32 ULID = 28 chars total.
  """

  import Bitwise

  @doc "Generate a new ULID string (26 chars, Crockford Base32)."
  def generate do
    Uniq.UUID.uuid7()
    |> encode_crockford()
  end

  @doc "Generate a prefixed ULID for a given entity type."
  def new_id(prefix) when is_binary(prefix) and byte_size(prefix) == 2 do
    prefix <> generate()
  end

  def new_challenge_id, do: new_id("CL")
  def new_project_id, do: new_id("PR")
  def new_user_id, do: new_id("US")
  def new_event_id, do: new_id("EV")
  def new_team_id, do: new_id("TM")
  def new_super_team_id, do: new_id("ST")

  # Encode a UUID v7 binary into 26-char Crockford Base32 (ULID format)
  defp encode_crockford(uuid_string) do
    uuid_string
    |> String.replace("-", "")
    |> Base.decode16!(case: :mixed)
    |> crockford_encode()
  end

  @crockford_alphabet ~c"0123456789ABCDEFGHJKMNPQRSTVWXYZ"

  defp crockford_encode(<<value::unsigned-128>>) do
    encode_crockford_chars(value, 26, [])
    |> List.to_string()
  end

  defp encode_crockford_chars(_value, 0, acc), do: acc

  defp encode_crockford_chars(value, remaining, acc) do
    char = Enum.at(@crockford_alphabet, value &&& 0x1F)
    encode_crockford_chars(value >>> 5, remaining - 1, [char | acc])
  end
end

defmodule ElixirBackend.PrefixedID do
  @moduledoc """
  Custom Ecto type for prefixed ULID strings (28-char CHAR fields).
  """
  use Ecto.Type

  def type, do: :string

  def cast(id) when is_binary(id) and byte_size(id) == 28, do: {:ok, id}
  def cast(_), do: :error

  def load(id) when is_binary(id), do: {:ok, String.trim(id)}
  def load(_), do: :error

  def dump(id) when is_binary(id) and byte_size(id) == 28, do: {:ok, id}
  def dump(_), do: :error
end
