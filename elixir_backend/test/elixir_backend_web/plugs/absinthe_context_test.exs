defmodule ElixirBackendWeb.Plugs.AbsintheContextTest do
  use ElixirBackendWeb.ConnCase

  describe "language parsing from Accept-Language header" do
    test "extracts simple language code" do
      conn =
        build_conn()
        |> put_req_header("accept-language", "en")
        |> ElixirBackendWeb.Plugs.AbsintheContext.call([])

      assert conn.private.absinthe.context.language == "en"
    end

    test "extracts language from locale with region" do
      conn =
        build_conn()
        |> put_req_header("accept-language", "en-US")
        |> ElixirBackendWeb.Plugs.AbsintheContext.call([])

      assert conn.private.absinthe.context.language == "en"
    end

    test "extracts first language from weighted list" do
      conn =
        build_conn()
        |> put_req_header("accept-language", "de-DE,en-US;q=0.9,en;q=0.8")
        |> ElixirBackendWeb.Plugs.AbsintheContext.call([])

      assert conn.private.absinthe.context.language == "de"
    end

    test "defaults to no when header is missing" do
      conn =
        build_conn()
        |> ElixirBackendWeb.Plugs.AbsintheContext.call([])

      assert conn.private.absinthe.context.language == "no"
    end

    test "defaults to no for empty header" do
      conn =
        build_conn()
        |> put_req_header("accept-language", "")
        |> ElixirBackendWeb.Plugs.AbsintheContext.call([])

      assert conn.private.absinthe.context.language == "no"
    end
  end
end
