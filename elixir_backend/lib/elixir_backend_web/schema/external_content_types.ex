defmodule ElixirBackendWeb.Schema.ExternalContentTypes do
  use Absinthe.Schema.Notation

  alias ElixirBackend.ExternalContent, as: EC

  enum :external_content_type do
    value(:media, as: "MEDIA")
    value(:article, as: "ARTICLE")
    value(:book_chapter, as: "BOOK_CHAPTER")
    value(:bible_verse, as: "BIBLE_VERSE")
    value(:external_link, as: "EXTERNAL_LINK")
    value(:quiz, as: "QUIZ")
    value(:song, as: "SONG")
    value(:text, as: "TEXT")
  end

  enum :external_content_sort_by do
    value(:created_at_asc, as: "CREATED_AT_ASC")
    value(:created_at_desc, as: "CREATED_AT_DESC")
    value(:published_at_asc, as: "PUBLISHED_AT_ASC")
    value(:published_at_desc, as: "PUBLISHED_AT_DESC")
  end

  object :external_content do
    field :id, non_null(:id)
    field :plan_id, non_null(:string)
    field :task_id, non_null(:string)
    field :content_id, :string
    field :content_type, non_null(:external_content_type)
    field :published_at, :datetime
    field :source, non_null(:string)
    field :synced_at, non_null(:datetime)
    field :url, :string
    field :created_at, non_null(:datetime)
    field :updated_at, non_null(:datetime)

    field :translations, non_null(list_of(non_null(:external_content_translation))) do
      resolve(fn content, _, _ ->
        {:ok, EC.get_translations(content.id)}
      end)
    end

    field :title, :string do
      resolve(fn content, _, %{context: context} ->
        lang = context[:language] || "en"
        {:ok, EC.get_title(content.id, lang)}
      end)
    end
  end

  object :external_content_translation do
    field :language_code, non_null(:string)
    field :title, :string
  end

  input_object :external_content_filter do
    field :plan_id, :string
    field :task_id, :string
    field :content_id, :string
    field :content_type, :external_content_type
    field :source, :string
    field :published_after, :datetime
    field :published_before, :datetime
    field :ids, list_of(non_null(:id))
  end

  # Pagination types
  object :external_content_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:external_content)
  end

  object :external_content_connection do
    field :edges, non_null(list_of(non_null(:external_content_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end
end
