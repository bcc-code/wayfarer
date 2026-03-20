defmodule ElixirBackend.PushNotifications.Preference do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "push_notification_preferences" do
    field :user_id, :string
    field :notification_type, :string
    field :enabled, :boolean, default: true

    timestamps(type: :utc_datetime)
  end

  def changeset(pref, attrs) do
    pref
    |> cast(attrs, [:user_id, :notification_type, :enabled])
    |> validate_required([:user_id, :notification_type])
    |> unique_constraint([:user_id, :notification_type])
  end
end
