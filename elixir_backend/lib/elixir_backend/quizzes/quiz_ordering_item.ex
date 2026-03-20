defmodule ElixirBackend.Quizzes.QuizOrderingItem do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "quiz_ordering_items" do
    field :item_text, :string
    field :correct_order, :integer

    belongs_to :question, ElixirBackend.Quizzes.QuizQuestion, type: :string

    timestamps(type: :utc_datetime)
  end

  def changeset(item, attrs) do
    item
    |> cast(attrs, [:id, :item_text, :correct_order, :question_id])
    |> validate_required([:id, :item_text, :correct_order, :question_id])
    |> foreign_key_constraint(:question_id)
  end
end
