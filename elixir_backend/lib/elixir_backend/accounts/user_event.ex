defmodule ElixirBackend.Accounts.UserEvent do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "user_events" do
    field :user_id, :string
    field :event_id, :string
    field :joined_at, :utc_datetime
  end

  def changeset(user_event, attrs) do
    user_event
    |> cast(attrs, [:user_id, :event_id, :joined_at])
    |> validate_required([:user_id, :event_id, :joined_at])
    |> unique_constraint([:user_id, :event_id])
    |> foreign_key_constraint(:user_id)
    |> foreign_key_constraint(:event_id)
  end
end
