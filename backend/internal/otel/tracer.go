package otel

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcoauth "google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/credentials/insecure"
)

// TracerProvider holds the OpenTelemetry tracer provider and related resources
type TracerProvider struct {
	provider *sdktrace.TracerProvider
}

// Config holds configuration for initializing OpenTelemetry
type Config struct {
	Enabled          bool
	ServiceName      string
	ServiceVersion   string
	ExporterEndpoint string
	ExporterInsecure bool
	SamplingRatio    float64
}

// InitTracer initializes the OpenTelemetry tracer with the given configuration
// Returns nil if tracing is disabled
func InitTracer(ctx context.Context, cfg Config) (*TracerProvider, error) {
	if !cfg.Enabled {
		slog.Info("OpenTelemetry tracing is disabled")
		return nil, nil
	}

	// Build resource attributes
	resourceAttrs := []resource.Option{
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	}

	// Add GCP cloud attributes if project ID is available
	gcpProjectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if gcpProjectID != "" {
		resourceAttrs = append(resourceAttrs, resource.WithAttributes(
			semconv.CloudProviderGCP,
			semconv.CloudAccountID(gcpProjectID),
		))
	}

	// Create resource with service information
	res, err := resource.New(ctx, resourceAttrs...)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Configure OTLP exporter with gRPC
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
	}

	if cfg.ExporterInsecure {
		opts = append(opts,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		)
	} else {
		// Use TLS with Google Application Default Credentials
		tokenSource, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/trace.append")
		if err != nil {
			return nil, fmt.Errorf("failed to get Google default token source: %w", err)
		}
		opts = append(opts,
			otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))),
			otlptracegrpc.WithDialOption(grpc.WithPerRPCCredentials(grpcoauth.TokenSource{TokenSource: tokenSource})),
		)
		slog.Info("OpenTelemetry: using Google Cloud credentials")
	}

	exporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(opts...))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Configure sampling based on ratio
	var sampler sdktrace.Sampler
	if cfg.SamplingRatio >= 1.0 {
		sampler = sdktrace.AlwaysSample()
		slog.Info("OpenTelemetry: sampling all traces (100%)")
	} else if cfg.SamplingRatio <= 0.0 {
		sampler = sdktrace.NeverSample()
		slog.Info("OpenTelemetry: sampling disabled (0%)")
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplingRatio)
		slog.Info("OpenTelemetry: probabilistic sampling enabled",
			"ratio", cfg.SamplingRatio,
			"percentage", cfg.SamplingRatio*100)
	}

	// Create tracer provider with batch span processor
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
	)

	// Set global tracer provider
	otel.SetTracerProvider(provider)

	// Set global propagator to extract and inject trace context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logAttrs := []any{
		"service", cfg.ServiceName,
		"version", cfg.ServiceVersion,
		"endpoint", cfg.ExporterEndpoint,
		"insecure", cfg.ExporterInsecure,
	}
	if gcpProjectID != "" {
		logAttrs = append(logAttrs, "gcp_project", gcpProjectID)
	}
	slog.Info("OpenTelemetry tracer initialized successfully", logAttrs...)

	return &TracerProvider{
		provider: provider,
	}, nil
}

// Shutdown gracefully shuts down the tracer provider
// This should be called when the application is shutting down
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp == nil || tp.provider == nil {
		return nil
	}

	slog.Info("Shutting down OpenTelemetry tracer provider...")
	if err := tp.provider.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown tracer provider: %w", err)
	}

	slog.Info("OpenTelemetry tracer provider shutdown successfully")
	return nil
}
