defmodule ElixirBackendWeb.Schema.FileUploadTypes do
  use Absinthe.Schema.Notation
  @moduledoc false

  object :file_upload do
    field :id, non_null(:id)
    field :filename, non_null(:string)
    field :stored_filename, non_null(:string)
    field :file_size, non_null(:integer)
    field :mime_type, non_null(:string)
    field :public_url, non_null(:string)
    field :uploaded_by, non_null(:id)
    field :width, :integer
    field :height, :integer
    field :blurhash, :string
    field :created_at, non_null(:datetime)
  end
end
