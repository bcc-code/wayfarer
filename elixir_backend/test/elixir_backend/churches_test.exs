defmodule ElixirBackend.ChurchesTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Churches

  import ElixirBackend.TestHelpers

  describe "get_church/1" do
    test "returns church by id" do
      church = create_church(%{name: "Oslo Church"})
      assert {:ok, found} = Churches.get_church(church.id)
      assert found.id == church.id
      assert found.name == "Oslo Church"
    end

    test "returns error for nonexistent church" do
      assert {:error, :not_found} = Churches.get_church("CH00000000000000000000000000")
    end
  end

  describe "list_churches/2" do
    test "returns paginated churches" do
      create_church(%{name: "Church A"})
      create_church(%{name: "Church B"})
      create_church(%{name: "Church C"})

      assert {:ok, result} = Churches.list_churches(%{}, %{first: 2})
      assert length(result.edges) == 2
      assert result.total_count >= 3
      assert result.page_info.has_next_page == true
    end

    test "filters by country" do
      create_church(%{name: "Norwegian", country: "NO"})
      create_church(%{name: "Swedish", country: "SE"})

      assert {:ok, result} = Churches.list_churches(%{country: "NO"}, %{first: 10})
      assert Enum.all?(result.edges, fn e -> e.node.country == "NO" end)
    end

    test "filters by category" do
      create_church(%{name: "Small", category: "S"})
      create_church(%{name: "Large", category: "L"})

      assert {:ok, result} = Churches.list_churches(%{category: "S"}, %{first: 10})
      assert Enum.all?(result.edges, fn e -> e.node.category == "S" end)
    end

    test "filters by ids" do
      c1 = create_church(%{name: "Church 1"})
      _c2 = create_church(%{name: "Church 2"})

      assert {:ok, result} = Churches.list_churches(%{ids: [c1.id]}, %{first: 10})
      assert result.total_count == 1
      assert hd(result.edges).node.id == c1.id
    end
  end

  describe "update_church/2" do
    test "updates church fields" do
      church = create_church(%{name: "Old Name"})
      assert {:ok, updated} = Churches.update_church(church.id, %{name: "New Name"})
      assert updated.name == "New Name"
    end

    test "validates category on update" do
      church = create_church()
      assert {:error, changeset} = Churches.update_church(church.id, %{category: "INVALID"})
      assert errors_on(changeset)[:category] != nil
    end

    test "returns error for nonexistent church" do
      assert {:error, :not_found} =
               Churches.update_church("CH00000000000000000000000000", %{name: "X"})
    end
  end
end
