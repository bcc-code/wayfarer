package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracerName is the name used for all spans created by this package
const TracerName = "wayfarer-backend"

// Tracer returns the global tracer for the wayfarer backend
func Tracer() trace.Tracer {
	return otel.Tracer(TracerName)
}

// StartSpan creates a new span with the given name and optional attributes
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, trace.WithAttributes(attrs...))
}

// StartDataloaderSpan creates a span for dataloader batch operations
func StartDataloaderSpan(ctx context.Context, loaderName string, batchSize int) (context.Context, trace.Span) {
	return StartSpan(ctx, "dataloader/"+loaderName,
		attribute.Int("dataloader.batch_size", batchSize),
		attribute.String("dataloader.name", loaderName),
	)
}

// RecordCacheHitMiss records cache hit/miss statistics on an existing span
func RecordCacheHitMiss(span trace.Span, hits, misses int) {
	span.SetAttributes(
		attribute.Int("cache.hits", hits),
		attribute.Int("cache.misses", misses),
	)
}

// RecordError records an error on the span and sets the status to error
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetUserID adds the user ID attribute to a span
func SetUserID(span trace.Span, userID string) {
	span.SetAttributes(attribute.String("user.id", userID))
}

// SetQuizID adds the quiz ID attribute to a span
func SetQuizID(span trace.Span, quizID string) {
	span.SetAttributes(attribute.String("quiz.id", quizID))
}

// SetSubmissionID adds the submission ID attribute to a span
func SetSubmissionID(span trace.Span, submissionID string) {
	span.SetAttributes(attribute.String("submission.id", submissionID))
}

// SetProjectID adds the project ID attribute to a span
func SetProjectID(span trace.Span, projectID string) {
	span.SetAttributes(attribute.String("project.id", projectID))
}

// SetTransactionSuccess records whether a transaction succeeded
func SetTransactionSuccess(span trace.Span, success bool) {
	span.SetAttributes(attribute.Bool("transaction.success", success))
}
