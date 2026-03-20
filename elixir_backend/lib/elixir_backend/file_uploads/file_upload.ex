defmodule ElixirBackend.FileUploads.FileUpload do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "file_uploads" do
    field :filename, :string
    field :stored_filename, :string
    field :file_size, :integer
    field :mime_type, :string
    field :public_url, :string
    field :uploaded_by, :string
    field :width, :integer
    field :height, :integer
    field :blurhash, :string
    field :created_at, :utc_datetime
  end

  def changeset(upload, attrs) do
    upload
    |> cast(attrs, [
      :id,
      :filename,
      :stored_filename,
      :file_size,
      :mime_type,
      :public_url,
      :uploaded_by,
      :width,
      :height,
      :blurhash,
      :created_at
    ])
    |> validate_required([:id, :filename, :stored_filename])
  end
end
