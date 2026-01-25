package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName    = "go-adventure"
	serviceVersion = "1.0.0"
)

var (
	tracer trace.Tracer
)

// Config holds telemetry configuration
type Config struct {
	Enabled      bool
	OTLPEndpoint string // e.g., "localhost:4318" for Jaeger
	UseStdout    bool   // If true, export to stdout instead of OTLP
}

// InitTracing initializes OpenTelemetry tracing
func InitTracing(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if !cfg.Enabled {
		// Return a no-op tracer
		tracer = otel.Tracer(serviceName)
		return func(ctx context.Context) error { return nil }, nil
	}

	// Create resource with service info
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
		attribute.String("game.type", "text-adventure"),
	)

	var exporter sdktrace.SpanExporter
	var err error

	if cfg.UseStdout {
		// Export to stdout (useful for debugging)
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
			stdouttrace.WithWriter(os.Stdout),
		)
	} else {
		// Export via OTLP HTTP (works with Jaeger, etc.)
		endpoint := cfg.OTLPEndpoint
		if endpoint == "" {
			endpoint = "localhost:4318"
		}
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithInsecure(),
		)
	}
	if err != nil {
		return nil, err
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Register as global tracer provider
	otel.SetTracerProvider(tp)

	// Get tracer for our service
	tracer = tp.Tracer(serviceName)

	// Return shutdown function
	return tp.Shutdown, nil
}

// Tracer returns the configured tracer
func Tracer() trace.Tracer {
	if tracer == nil {
		tracer = otel.Tracer(serviceName)
	}
	return tracer
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, opts...)
}

// Common attribute keys for the game
var (
	AttrLocation     = attribute.Key("game.location")
	AttrLocationName = attribute.Key("game.location_name")
	AttrCommand      = attribute.Key("game.command")
	AttrVerb         = attribute.Key("game.verb")
	AttrObject       = attribute.Key("game.object")
	AttrScore        = attribute.Key("game.score")
	AttrTurns        = attribute.Key("game.turns")
	AttrInventory    = attribute.Key("game.inventory")
	AttrMovedFrom    = attribute.Key("game.moved_from")
	AttrMovedTo      = attribute.Key("game.moved_to")
)

// AddGameEvent adds an event to the current span
func AddGameEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SpanFromContext extracts the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	if ctx == nil {
		return nil
	}
	return trace.SpanFromContext(ctx)
}
