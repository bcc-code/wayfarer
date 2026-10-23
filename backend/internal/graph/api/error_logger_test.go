package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func newCapturingErrorLogger() (ErrorLogger, *bytes.Buffer) {
	var buf bytes.Buffer
	return ErrorLogger{Logger: slog.New(slog.NewJSONHandler(&buf, nil))}, &buf
}

func respondWith(resp *graphql.Response) graphql.ResponseHandler {
	return func(context.Context) *graphql.Response { return resp }
}

func TestErrorLoggerLogsErroredResponses(t *testing.T) {
	el, buf := newCapturingErrorLogger()
	ctx := graphql.WithOperationContext(context.Background(), &graphql.OperationContext{
		OperationName: "MyTeam",
		Operation:     &ast.OperationDefinition{Operation: ast.Query},
		Variables:     map[string]any{"secret": "do-not-log"},
	})
	resp := &graphql.Response{Errors: gqlerror.List{
		{Message: "user not authenticated", Path: ast.Path{ast.PathName("myTeam")}, Extensions: map[string]any{"code": "UNAUTHENTICATED"}},
	}}

	got := el.InterceptResponse(ctx, respondWith(resp))
	assert.Same(t, resp, got)

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	assert.Equal(t, "WARN", line["level"])
	assert.Equal(t, "MyTeam", line["operation"])
	assert.Equal(t, "query", line["operation_type"])
	errs := line["errors"].([]any)
	require.Len(t, errs, 1)
	e := errs[0].(map[string]any)
	assert.Equal(t, "user not authenticated", e["message"])
	assert.Equal(t, "myTeam", e["path"])
	assert.Equal(t, "UNAUTHENTICATED", e["code"])
	assert.NotContains(t, buf.String(), "do-not-log")
}

func TestErrorLoggerSkipsCleanAndAPQResponses(t *testing.T) {
	el, buf := newCapturingErrorLogger()
	el.InterceptResponse(context.Background(), respondWith(&graphql.Response{}))
	el.InterceptResponse(context.Background(), respondWith(nil))
	el.InterceptResponse(context.Background(), respondWith(&graphql.Response{
		Errors: gqlerror.List{{Message: apqNotFound}},
	}))
	assert.Empty(t, buf.String())

	assert.NotPanics(t, func() {
		ErrorLogger{}.InterceptResponse(context.Background(), respondWith(&graphql.Response{
			Errors: gqlerror.List{{Message: "boom"}},
		}))
	}, "nil Logger is a no-op")
}

func TestErrorLoggerCapsErrors(t *testing.T) {
	el, buf := newCapturingErrorLogger()
	var list gqlerror.List
	for range 20 {
		list = append(list, &gqlerror.Error{Message: "boom"})
	}
	el.InterceptResponse(context.Background(), respondWith(&graphql.Response{Errors: list}))

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	assert.Len(t, line["errors"], maxLoggedErrors)
	assert.Equal(t, float64(20), line["error_count"])
}
