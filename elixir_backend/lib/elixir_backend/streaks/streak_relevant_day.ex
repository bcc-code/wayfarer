defmodule ElixirBackend.Streaks.StreakRelevantDay do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "streak_relevant_days" do
    field :start_date, :date
    field :end_date, :date

    belongs_to :streak, ElixirBackend.Streaks.Streak, type: :string
  end

  def changeset(day, attrs) do
    day
    |> cast(attrs, [:id, :start_date, :end_date, :streak_id])
    |> validate_required([:id, :start_date, :end_date, :streak_id])
    |> validate_date_range()
    |> foreign_key_constraint(:streak_id)
  end

  defp validate_date_range(changeset) do
    start_date = get_field(changeset, :start_date)
    end_date = get_field(changeset, :end_date)

    if start_date && end_date && Date.compare(end_date, start_date) == :lt do
      add_error(changeset, :end_date, "must be on or after start_date")
    else
      changeset
    end
  end
end
