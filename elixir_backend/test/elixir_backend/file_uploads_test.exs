defmodule ElixirBackend.FileUploadsTest do
  use ElixirBackend.DataCase, async: true

  alias ElixirBackend.FileUploads
  alias ElixirBackend.TestHelpers

  describe "create_upload/1" do
    test "creates a file upload record" do
      user = TestHelpers.create_user()

      {:ok, upload} =
        FileUploads.create_upload(%{
          filename: "photo.jpg",
          stored_filename: "abc123.jpg",
          file_size: 12_345,
          mime_type: "image/jpeg",
          public_url: "https://cdn.example.com/abc123.jpg",
          uploaded_by: user.id,
          width: 1920,
          height: 1080,
          blurhash: "LEHV6nWB2yk8pyo0adR*.7kCMdnj"
        })

      assert upload.filename == "photo.jpg"
      assert upload.file_size == 12_345
      assert upload.width == 1920
      assert upload.blurhash == "LEHV6nWB2yk8pyo0adR*.7kCMdnj"
    end
  end

  describe "get_upload/1" do
    test "returns upload by id" do
      user = TestHelpers.create_user()

      {:ok, upload} =
        FileUploads.create_upload(%{
          filename: "doc.pdf",
          stored_filename: "xyz789.pdf",
          file_size: 5000,
          mime_type: "application/pdf",
          public_url: "https://cdn.example.com/xyz789.pdf",
          uploaded_by: user.id
        })

      {:ok, found} = FileUploads.get_upload(upload.id)
      assert found.id == upload.id
      assert found.filename == "doc.pdf"
    end

    test "returns error for missing upload" do
      assert {:error, :not_found} = FileUploads.get_upload("FL00000000000000000000000000")
    end
  end
end
