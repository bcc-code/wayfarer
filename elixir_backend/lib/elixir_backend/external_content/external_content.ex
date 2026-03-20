defmodule ElixirBackend.ExternalContent.Content do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  @valid_types ~w(MEDIA ARTICLE BOOK_CHAPTER BIBLE_VERSE EXTERNAL_LINK QUIZ SONG TEXT)

  schema "external_content" do
    field :plan_id, :string
    field :task_id, :string
    field :content_id, :string
    field :content_type, :string
    field :published_at, :utc_datetime
    field :synced_at, :utc_datetime
    field :source, :string
    field :url, :string
    field :complete_by, :utc_datetime

    has_many :translations, ElixirBackend.ExternalContent.Translation,
      foreign_key: :external_content_id

    timestamps(type: :utc_datetime)
  end

  def changeset(content, attrs) do
    content
    |> cast(attrs, [
      :id,
      :plan_id,
      :task_id,
      :content_id,
      :content_type,
      :published_at,
      :synced_at,
      :source,
      :url,
      :complete_by
    ])
    |> validate_required([:id, :plan_id, :task_id, :content_type, :source])
    |> validate_inclusion(:content_type, @valid_types)
    |> unique_constraint([:plan_id, :task_id])
  end

  def update_changeset(content, attrs) do
    content
    |> cast(attrs, [
      :content_id,
      :content_type,
      :published_at,
      :synced_at,
      :source,
      :url,
      :complete_by
    ])
    |> validate_inclusion(:content_type, @valid_types)
  end
end
