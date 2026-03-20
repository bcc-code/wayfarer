defmodule ElixirBackendWeb.Schema.ChurchTypes do
  @moduledoc "Absinthe types for churches: object, inputs, pagination."
  use Absinthe.Schema.Notation

  # ── Enums ──

  enum :church_category do
    value(:s, as: "S")
    value(:l, as: "L")
    value(:xl, as: "XL")
  end

  # ── Church object ──

  object :church do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :country, non_null(:string)
    field :category, non_null(:church_category)
  end

  # ── Pagination ──

  object :church_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:church)
  end

  object :church_connection do
    field :edges, non_null(list_of(non_null(:church_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end

  # ── Input types ──

  input_object :update_church_input do
    field :name, :string
    field :country, :string
    field :category, :church_category
  end

  input_object :church_filter do
    field :ids, list_of(non_null(:id))
    field :country, :string
    field :category, :church_category
  end
end
