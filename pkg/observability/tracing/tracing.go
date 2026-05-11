package tracing

import (
	"context"
	"fmt"
	"strings"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// NewTracerProvider builds an OTEL tracer provider for the given service.
func NewTracerProvider(ctx context.Context, cfg Config) (*sdktrace.TracerProvider, error) {
	if !cfg.Enabled {
		return sdktrace.NewTracerProvider(), nil
	}
	if strings.TrimSpace(cfg.ServiceName) == "" {
		return nil, fmt.Errorf("service name is empty")
	}

	exp, err := newExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}

	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceNamespace("ofm"),
			attribute.String("service.env", cfg.ServiceEnv),
			attribute.String("service.version", cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))
	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exp),
	), nil
}

func newExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = "http://127.0.0.1:9099"
	}
	protocol := strings.TrimSpace(cfg.Protocol)
	if protocol == "" {
		protocol = "http/protobuf"
	}

	switch protocol {
	case "http/protobuf", "http":
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(endpoint)}
		return otlptracehttp.New(ctx, opts...)
	default:
		return nil, fmt.Errorf("unsupported otlp protocol %q", protocol)
	}
}

// Propagator returns the W3C trace-context propagator used by OFM services.
func Propagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// Fields extracts trace correlation fields from the context.
func Fields(ctx context.Context) []logging.Field {
	if ctx == nil {
		return nil
	}
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return nil
	}
	fields := []logging.Field{
		logging.String("trace_id", sc.TraceID().String()),
		logging.String("span_id", sc.SpanID().String()),
	}
	return fields
}

// SetGlobal installs the tracing provider and propagator for the process.
func SetGlobal(tp *sdktrace.TracerProvider) {
	if tp != nil {
		otel.SetTracerProvider(tp)
	}
	otel.SetTextMapPropagator(Propagator())
}
