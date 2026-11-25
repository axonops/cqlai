// Package telemetry sends OpenTelemetry spans for the CQL cqlai runs.
//
// It is off unless asked for, with --otel, CQLAI_OTEL or "otel": {"enabled":
// true} in cqlai.json. Off, nothing here is built: no tracer provider, no
// exporter, no observer on the cluster. The OTEL_* variables on their own do
// not turn it on, since they may be set in the environment for something else.
//
// These are client-side spans sent to an OTLP collector. They have nothing to
// do with TRACING ON, which asks Cassandra to record its own trace of a request.
package telemetry

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

// ServiceName is service.name unless OTEL_SERVICE_NAME or
// OTEL_RESOURCE_ATTRIBUTES say otherwise.
const ServiceName = "cqlai"

var (
	mu       sync.Mutex
	provider *sdktrace.TracerProvider
	version  = "dev"
)

// SetServiceVersion sets service.version on the spans. It has to be called
// before Start.
func SetServiceVersion(v string) {
	mu.Lock()
	defer mu.Unlock()
	version = v
}

// Start sets up the tracer provider and the exporter. A second call does
// nothing, so it can be made for every connection.
func Start(ctx context.Context) error {
	mu.Lock()
	defer mu.Unlock()
	if provider != nil {
		return nil
	}

	// The default handler and logger write to stderr, which is the terminal
	// the UI is drawn on, and in batch mode is the output.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.DebugfToFile("Telemetry", "OpenTelemetry error: %v", err)
	}))
	otel.SetLogger(logr.Discard())

	exporter, err := newExporter(ctx)
	if err != nil {
		return err
	}

	// Later options win, so the environment overrides the name given here.
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(ServiceName), semconv.ServiceVersion(version)),
		resource.WithTelemetrySDK(),
		resource.WithFromEnv(),
	)
	if err != nil {
		_ = exporter.Shutdown(ctx)
		return fmt.Errorf("OpenTelemetry resource: %w", err)
	}

	provider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)
	logger.DebugfToFile("Telemetry", "OpenTelemetry tracing started")
	return nil
}

// Started reports whether Start has set tracing up.
func Started() bool {
	mu.Lock()
	defer mu.Unlock()
	return provider != nil
}

// TracerProvider is the provider Start set up, nil before then.
func TracerProvider() trace.TracerProvider {
	mu.Lock()
	defer mu.Unlock()
	if provider == nil {
		return nil
	}
	return provider
}

// Shutdown sends the spans still waiting in the batch and stops the exporter.
// It does nothing if tracing was never started. It has to run before the
// process exits: a run of `cqlai -e` is over well before the batch would have
// been sent on its own.
func Shutdown(ctx context.Context) error {
	mu.Lock()
	defer mu.Unlock()
	if provider == nil {
		return nil
	}
	err := provider.Shutdown(ctx)
	provider = nil
	return err
}

// newExporter picks the OTLP exporter for the protocol the environment asks
// for. The exporters read the endpoint, headers, TLS and timeout settings from
// OTEL_EXPORTER_OTLP_* themselves, but neither chooses between gRPC and HTTP.
func newExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
	if protocol == "" {
		protocol = os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	}
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "", "http/protobuf":
		return otlptracehttp.New(ctx)
	case "grpc":
		return otlptracegrpc.New(ctx)
	default:
		return nil, fmt.Errorf("unsupported OTLP protocol %q (supported: grpc, http/protobuf)", protocol)
	}
}
