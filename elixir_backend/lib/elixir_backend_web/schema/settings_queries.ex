defmodule ElixirBackendWeb.Schema.SettingsQueries do
  use Absinthe.Schema.Notation
  @moduledoc false

  object :settings_queries do
    field :frontend_config, non_null(:json) do
      resolve(fn _, _, _ ->
        # Returns an empty config by default; can be extended later
        {:ok, %{}}
      end)
    end
  end
end
