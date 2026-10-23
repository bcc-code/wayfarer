package api

import (
	"context"
	"log/slog"

	"github.com/99designs/gqlgen/graphql"
)

const (
	// maxLoggedErrors caps how many errors of one response are logged.
	maxLoggedErrors = 5
	// apqNotFound is the message of gqlgen's (unexported) APQ cache-miss error.
	apqNotFound = "PersistedQueryNotFound"
)

// ErrorLogger is a gqlgen extension that logs one Warn line per GraphQL
// response carrying errors: operation name/type plus each error's message and
// path. Variables are never logged (they can hold user data). APQ cache
// misses are skipped — they are part of the normal persisted-query handshake.
//
// Logger is meant to be export-only (VictoriaLogs), so client-caused errors
// at load-test rates never turn into synchronous stdout/journald writes.
type ErrorLogger struct {
	Logger *slog.Logger
}

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
} = ErrorLogger{}

func (ErrorLogger) ExtensionName() string { return "ErrorLogger" }

func (ErrorLogger) Validate(graphql.ExecutableSchema) error { return nil }

func (l ErrorLogger) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	resp := next(ctx)
	if l.Logger == nil || resp == nil || len(resp.Errors) == 0 {
		return resp
	}

	errs := make([]map[string]string, 0, min(len(resp.Errors), maxLoggedErrors))
	for _, e := range resp.Errors {
		if e.Message == apqNotFound {
			continue
		}
		if len(errs) == maxLoggedErrors {
			break
		}
		entry := map[string]string{"message": e.Message}
		if len(e.Path) > 0 {
			entry["path"] = e.Path.String()
		}
		if code, ok := e.Extensions["code"].(string); ok {
			entry["code"] = code
		}
		errs = append(errs, entry)
	}
	if len(errs) == 0 {
		return resp
	}

	attrs := []any{"error_count", len(resp.Errors), "errors", errs}
	if graphql.HasOperationContext(ctx) {
		oc := graphql.GetOperationContext(ctx)
		attrs = append(attrs, "operation", oc.OperationName)
		if oc.Operation != nil {
			attrs = append(attrs, "operation_type", string(oc.Operation.Operation))
		}
	}
	l.Logger.WarnContext(ctx, "graphql response has errors", attrs...)
	return resp
}
