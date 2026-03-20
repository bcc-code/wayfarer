defmodule ElixirBackend.FeedbackTest do
  use ElixirBackend.DataCase, async: true

  alias ElixirBackend.Feedback
  alias ElixirBackend.TestHelpers

  describe "submit_feedback/1" do
    test "creates feedback with required fields" do
      user = TestHelpers.create_user()

      {:ok, fb} =
        Feedback.submit_feedback(%{
          user_id: user.id,
          message: "Great app!",
          can_contact_me: true,
          platform: "web"
        })

      assert fb.message == "Great app!"
      assert fb.can_contact_me == true
      assert fb.platform == "web"
    end
  end

  describe "list_feedback/2" do
    test "returns paginated feedback" do
      user = TestHelpers.create_user()

      for i <- 1..3 do
        Feedback.submit_feedback(%{
          user_id: user.id,
          message: "Feedback #{i}",
          can_contact_me: false
        })
      end

      {:ok, result} = Feedback.list_feedback()
      assert result.total_count >= 3
      assert length(result.edges) >= 3
    end

    test "filters by platform" do
      user = TestHelpers.create_user()

      Feedback.submit_feedback(%{
        user_id: user.id,
        message: "Web feedback",
        can_contact_me: false,
        platform: "web"
      })

      Feedback.submit_feedback(%{
        user_id: user.id,
        message: "iOS feedback",
        can_contact_me: false,
        platform: "ios"
      })

      {:ok, result} = Feedback.list_feedback(%{platform: "web"})
      assert Enum.all?(result.edges, fn e -> e.node.platform == "web" end)
    end

    test "filters by handled status" do
      user = TestHelpers.create_user()

      {:ok, fb} =
        Feedback.submit_feedback(%{
          user_id: user.id,
          message: "Handled one",
          can_contact_me: false
        })

      Feedback.mark_handled(fb.id)

      {:ok, handled} = Feedback.list_feedback(%{handled: true})
      assert Enum.all?(handled.edges, fn e -> e.node.handled_at != nil end)

      {:ok, unhandled} = Feedback.list_feedback(%{handled: false})
      assert Enum.all?(unhandled.edges, fn e -> e.node.handled_at == nil end)
    end
  end

  describe "mark_handled/1" do
    test "sets handled_at timestamp" do
      user = TestHelpers.create_user()

      {:ok, fb} =
        Feedback.submit_feedback(%{
          user_id: user.id,
          message: "Handle me",
          can_contact_me: false
        })

      assert fb.handled_at == nil
      {:ok, handled} = Feedback.mark_handled(fb.id)
      assert handled.handled_at != nil
    end
  end

  describe "update_tags/2" do
    test "updates feedback tags" do
      user = TestHelpers.create_user()

      {:ok, fb} =
        Feedback.submit_feedback(%{
          user_id: user.id,
          message: "Tag me",
          can_contact_me: false
        })

      {:ok, updated} = Feedback.update_tags(fb.id, ["bug", "ui"])
      assert updated.tags == ["bug", "ui"]
    end
  end

  describe "delete_feedback/1" do
    test "deletes feedback" do
      user = TestHelpers.create_user()

      {:ok, fb} =
        Feedback.submit_feedback(%{
          user_id: user.id,
          message: "Delete me",
          can_contact_me: false
        })

      {:ok, _} = Feedback.delete_feedback(fb.id)
      assert {:error, :not_found} = Feedback.get_feedback(fb.id)
    end
  end

  describe "get_tags/0 and get_platforms/0" do
    test "returns distinct tags and platforms" do
      user = TestHelpers.create_user()

      {:ok, fb1} =
        Feedback.submit_feedback(%{
          user_id: user.id,
          message: "One",
          can_contact_me: false,
          platform: "web",
          tags: ["bug"]
        })

      {:ok, _fb2} =
        Feedback.submit_feedback(%{
          user_id: user.id,
          message: "Two",
          can_contact_me: false,
          platform: "ios",
          tags: ["feature", "bug"]
        })

      # Ensure tags saved correctly
      assert fb1.tags == ["bug"]

      tags = Feedback.get_tags()
      assert "bug" in tags
      assert "feature" in tags

      platforms = Feedback.get_platforms()
      assert "web" in platforms
      assert "ios" in platforms
    end
  end
end
