defmodule ElixirBackend.Consents.Consent do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "consents" do
    field :key, :string
    field :version, :integer, default: 1
    field :title, :string
    field :short_text, :string
    field :body, :string
    field :url, :string
    field :published_at, :utc_datetime
    field :managed_by, :string
    field :is_remote, :boolean, default: false

    timestamps(type: :utc_datetime)
  end

  def changeset(consent, attrs) do
    consent
    |> cast(attrs, [
      :id,
      :key,
      :version,
      :title,
      :short_text,
      :body,
      :url,
      :published_at,
      :managed_by,
      :is_remote
    ])
    |> validate_required([:id, :key, :title])
    |> unique_constraint([:key, :version])
  end
end
