defmodule ElixirBackend.FileUploads do
  @moduledoc "Context for file upload tracking."

  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.FileUploads.FileUpload

  def get_upload(id) do
    case Repo.get(FileUpload, id) do
      nil -> {:error, :not_found}
      upload -> {:ok, upload}
    end
  end

  def create_upload(attrs) do
    id = ULID.new_file_upload_id()
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    %FileUpload{}
    |> FileUpload.changeset(
      attrs
      |> Map.put(:id, id)
      |> Map.put(:created_at, now)
    )
    |> Repo.insert()
  end
end
