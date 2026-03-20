defmodule ElixirBackend.Consents.UserConsentHistory do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "user_consent_history" do
    field :consent_key, :string
    field :action, :string
    field :occurred_at, :utc_datetime
    field :source, :string
    field :external_consent_id, :string
    field :external_timestamp, :utc_datetime

    belongs_to :user, ElixirBackend.Accounts.User, type: :string
    belongs_to :consent, ElixirBackend.Consents.Consent, type: :string
  end

  @valid_actions ~w(ACCEPTED REJECTED)

  def changeset(entry, attrs) do
    entry
    |> cast(attrs, [
      :id,
      :user_id,
      :consent_id,
      :consent_key,
      :action,
      :occurred_at,
      :source,
      :external_consent_id,
      :external_timestamp
    ])
    |> validate_required([:id, :user_id, :consent_id, :consent_key, :action])
    |> validate_inclusion(:action, @valid_actions)
    |> foreign_key_constraint(:user_id)
    |> foreign_key_constraint(:consent_id)
  end
end
