defmodule ElixirBackend.EventsTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Events

  import ElixirBackend.TestHelpers

  describe "get_event/1" do
    test "returns event by id" do
      project = create_project()
      event = create_event(project, %{name: "My Event"})
      assert {:ok, found} = Events.get_event(event.id)
      assert found.id == event.id
      assert found.name == "My Event"
    end

    test "returns error for nonexistent event" do
      assert {:error, :not_found} = Events.get_event("EV00000000000000000000000000")
    end
  end

  describe "list_events/2" do
    test "returns paginated events" do
      project = create_project()
      create_event(project, %{name: "Event A"})
      create_event(project, %{name: "Event B"})
      create_event(project, %{name: "Event C"})

      assert {:ok, result} = Events.list_events(%{}, %{first: 2})
      assert length(result.edges) == 2
      assert result.total_count >= 3
      assert result.page_info.has_next_page == true
    end

    test "filters by project_id" do
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})
      create_event(p1, %{name: "E1"})
      create_event(p2, %{name: "E2"})

      assert {:ok, result} = Events.list_events(%{project_id: p1.id}, %{first: 10})
      assert result.total_count == 1
      assert hd(result.edges).node.name == "E1"
    end

    test "filters by ids" do
      project = create_project()
      e1 = create_event(project, %{name: "Event 1"})
      _e2 = create_event(project, %{name: "Event 2"})

      assert {:ok, result} = Events.list_events(%{ids: [e1.id]}, %{first: 10})
      assert result.total_count == 1
    end
  end

  describe "create_event/2" do
    test "creates an event with valid attrs" do
      project = create_project()

      assert {:ok, event} =
               Events.create_event(project.id, %{
                 name: "New Event",
                 description: "Description",
                 start_date: ~U[2026-06-01 00:00:00Z],
                 end_date: ~U[2026-06-30 23:59:59Z]
               })

      assert String.starts_with?(event.id, "EV")
      assert event.name == "New Event"
      assert event.project_id == project.id
    end

    test "fails without required name" do
      project = create_project()
      assert {:error, _changeset} = Events.create_event(project.id, %{description: "No name"})
    end
  end

  describe "update_event/2" do
    test "updates event fields" do
      project = create_project()
      event = create_event(project)
      assert {:ok, updated} = Events.update_event(event.id, %{name: "Updated Event"})
      assert updated.name == "Updated Event"
    end

    test "updates description" do
      project = create_project()
      event = create_event(project)
      assert {:ok, updated} = Events.update_event(event.id, %{description: "New desc"})
      assert updated.description == "New desc"
    end

    test "returns error for nonexistent event" do
      assert {:error, :not_found} =
               Events.update_event("EV00000000000000000000000000", %{name: "X"})
    end
  end

  describe "delete_event/1" do
    test "deletes an event" do
      project = create_project()
      event = create_event(project)
      assert {:ok, _} = Events.delete_event(event.id)
      assert {:error, :not_found} = Events.get_event(event.id)
    end

    test "returns error for nonexistent event" do
      assert {:error, :not_found} = Events.delete_event("EV00000000000000000000000000")
    end
  end

  describe "move_event/2" do
    test "moves event to a different project" do
      p1 = create_project(%{name: "Source"})
      p2 = create_project(%{name: "Target"})
      event = create_event(p1)

      assert {:ok, moved} = Events.move_event(event.id, p2.id)
      assert moved.project_id == p2.id
    end
  end

  describe "my_events/2" do
    test "returns events user has joined" do
      user = create_user()
      project = create_project()
      e1 = create_event(project, %{name: "Joined"})
      _e2 = create_event(project, %{name: "Not Joined"})
      create_user_event(user, e1)

      assert {:ok, events} = Events.my_events(user.id)
      assert length(events) == 1
      assert hd(events).id == e1.id
    end

    test "filters by project_id" do
      user = create_user()
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})
      e1 = create_event(p1, %{name: "E1"})
      e2 = create_event(p2, %{name: "E2"})
      create_user_event(user, e1)
      create_user_event(user, e2)

      assert {:ok, events} = Events.my_events(user.id, p1.id)
      assert length(events) == 1
      assert hd(events).project_id == p1.id
    end

    test "returns empty list for user with no events" do
      user = create_user()
      assert {:ok, []} = Events.my_events(user.id)
    end
  end

  describe "join_event/2" do
    test "joins a user to an event" do
      user = create_user()
      project = create_project()
      event = create_event(project)

      assert {:ok, joined} = Events.join_event(user.id, event.id)
      assert joined.id == event.id

      assert {:ok, [e]} = Events.my_events(user.id)
      assert e.id == event.id
    end

    test "idempotent - joining twice does not error" do
      user = create_user()
      project = create_project()
      event = create_event(project)

      assert {:ok, _} = Events.join_event(user.id, event.id)
      assert {:ok, _} = Events.join_event(user.id, event.id)

      assert {:ok, [_]} = Events.my_events(user.id)
    end
  end
end
