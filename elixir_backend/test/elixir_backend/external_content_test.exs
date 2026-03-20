defmodule ElixirBackend.ExternalContentTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.ExternalContent, as: EC

  defp create_content(attrs \\ %{}) do
    defaults = %{
      plan_id: "plan-#{System.unique_integer([:positive])}",
      task_id: "task-#{System.unique_integer([:positive])}",
      content_type: "ARTICLE",
      source: "ssf"
    }

    {:ok, content} = EC.upsert_content(Map.merge(defaults, attrs))
    content
  end

  describe "upsert_content/1" do
    test "creates new content" do
      assert {:ok, content} =
               EC.upsert_content(%{
                 plan_id: "plan-1",
                 task_id: "task-1",
                 content_type: "ARTICLE",
                 source: "ssf"
               })

      assert String.starts_with?(content.id, "EC")
      assert content.plan_id == "plan-1"
      assert content.content_type == "ARTICLE"
    end

    test "upserts on conflict" do
      {:ok, c1} =
        EC.upsert_content(%{
          plan_id: "plan-1",
          task_id: "task-1",
          content_type: "ARTICLE",
          source: "ssf"
        })

      {:ok, c2} =
        EC.upsert_content(%{
          plan_id: "plan-1",
          task_id: "task-1",
          content_type: "MEDIA",
          source: "other"
        })

      # Should update the existing record
      assert c2.id == c1.id
      assert c2.content_type == "MEDIA"
    end

    test "validates content_type" do
      assert {:error, _} =
               EC.upsert_content(%{
                 plan_id: "plan-1",
                 task_id: "task-1",
                 content_type: "INVALID",
                 source: "ssf"
               })
    end
  end

  describe "get_content/1" do
    test "returns content by id" do
      content = create_content()
      assert {:ok, found} = EC.get_content(content.id)
      assert found.id == content.id
    end

    test "returns error for non-existent id" do
      assert {:error, :not_found} = EC.get_content("EC00000000000000000000000000")
    end
  end

  describe "list_contents/3" do
    test "lists all content" do
      create_content(%{plan_id: "plan-a", task_id: "task-a"})
      create_content(%{plan_id: "plan-b", task_id: "task-b"})

      {:ok, connection} = EC.list_contents()
      assert connection.total_count >= 2
    end

    test "filters by plan_id" do
      create_content(%{plan_id: "plan-x", task_id: "task-1"})
      create_content(%{plan_id: "plan-y", task_id: "task-2"})

      {:ok, conn} = EC.list_contents(%{plan_id: "plan-x"})
      assert conn.total_count == 1
    end

    test "filters by content_type" do
      create_content(%{content_type: "MEDIA"})
      create_content(%{content_type: "SONG"})

      {:ok, conn} = EC.list_contents(%{content_type: "MEDIA"})
      assert conn.total_count == 1
    end

    test "filters by ids" do
      c1 = create_content()
      _c2 = create_content()

      {:ok, conn} = EC.list_contents(%{ids: [c1.id]})
      assert conn.total_count == 1
    end
  end

  describe "translations" do
    test "upserts a translation" do
      content = create_content()

      assert {:ok, _} =
               EC.upsert_translation(%{
                 external_content_id: content.id,
                 language_code: "en",
                 title: "Test Title"
               })

      translations = EC.get_translations(content.id)
      assert length(translations) == 1
      assert hd(translations).title == "Test Title"
    end

    test "upserts translation on conflict" do
      content = create_content()

      EC.upsert_translation(%{
        external_content_id: content.id,
        language_code: "en",
        title: "Old Title"
      })

      EC.upsert_translation(%{
        external_content_id: content.id,
        language_code: "en",
        title: "New Title"
      })

      translations = EC.get_translations(content.id)
      assert length(translations) == 1
      assert hd(translations).title == "New Title"
    end

    test "get_title returns title for language" do
      content = create_content()

      EC.upsert_translation(%{
        external_content_id: content.id,
        language_code: "no",
        title: "Norsk Tittel"
      })

      assert EC.get_title(content.id, "no") == "Norsk Tittel"
      assert EC.get_title(content.id, "en") == nil
    end
  end

  describe "delete_content/1" do
    test "deletes content" do
      content = create_content()
      assert {:ok, _} = EC.delete_content(content.id)
      assert {:error, :not_found} = EC.get_content(content.id)
    end
  end
end
