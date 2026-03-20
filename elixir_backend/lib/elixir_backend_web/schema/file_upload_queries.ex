defmodule ElixirBackendWeb.Schema.FileUploadQueries do
  use Absinthe.Schema.Notation

  alias ElixirBackend.FileUploads

  object :file_upload_queries do
    field :file_upload, :file_upload do
      arg(:id, non_null(:id))

      resolve(fn _, %{id: id}, _ ->
        FileUploads.get_upload(id)
      end)
    end
  end
end
