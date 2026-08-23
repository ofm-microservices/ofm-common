// Package metadata carries request correlation values across transport
// boundaries without coupling application contracts to HTTP headers.
package metadata

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const (
	requestIDKey     contextKey = "ofm.request_id"
	correlationIDKey contextKey = "ofm.correlation_id"
	idempotencyKey   contextKey = "ofm.idempotency_key"
	testRunIDKey     contextKey = "ofm.test_run_id"
	testScenarioKey  contextKey = "ofm.test_scenario_id"
)

// Values stores transport correlation metadata in a context.
type Values struct {
	RequestID      string
	CorrelationID  string
	IdempotencyKey string
	TestRunID      string
	TestScenarioID string
}

// WithValues attaches correlation values to a context.
func WithValues(ctx context.Context, values Values) context.Context {
	// Fiber may reuse the backing bytes of values returned by c.Get after the
	// handler returns. Clone before retaining them in a context that can outlive
	// the request and be encoded by a long-lived gRPC HTTP/2 connection.
	ctx = context.WithValue(ctx, requestIDKey, strings.Clone(strings.TrimSpace(values.RequestID)))
	ctx = context.WithValue(ctx, correlationIDKey, strings.Clone(strings.TrimSpace(values.CorrelationID)))
	ctx = context.WithValue(ctx, idempotencyKey, strings.Clone(strings.TrimSpace(values.IdempotencyKey)))
	ctx = context.WithValue(ctx, testRunIDKey, strings.Clone(strings.TrimSpace(values.TestRunID)))
	return context.WithValue(ctx, testScenarioKey, strings.Clone(strings.TrimSpace(values.TestScenarioID)))
}

// FromContext reads correlation values from a context.
func FromContext(ctx context.Context) Values {
	return Values{RequestID: value(ctx, requestIDKey), CorrelationID: value(ctx, correlationIDKey), IdempotencyKey: value(ctx, idempotencyKey), TestRunID: value(ctx, testRunIDKey), TestScenarioID: value(ctx, testScenarioKey)}
}

// OutgoingHeaders returns the bounded correlation headers that should be
// copied to asynchronous messages. Entity IDs and request-specific values are
// deliberately excluded to keep broker metadata low-cardinality.
func OutgoingHeaders(ctx context.Context) map[string]string {
	values := FromContext(ctx)
	result := make(map[string]string, 5)
	if values.RequestID != "" {
		result["x-request-id"] = values.RequestID
	}
	if values.CorrelationID != "" {
		result["x-correlation-id"] = values.CorrelationID
	}
	if values.IdempotencyKey != "" {
		result["idempotency-key"] = values.IdempotencyKey
	}
	if values.TestRunID != "" {
		result["x-test-run-id"] = values.TestRunID
	}
	if values.TestScenarioID != "" {
		result["x-test-scenario"] = values.TestScenarioID
	}
	return result
}

func value(ctx context.Context, key contextKey) string {
	result, _ := ctx.Value(key).(string)
	return result
}

// UnaryClientInterceptor propagates correlation metadata to downstream gRPC
// services while preserving existing outgoing metadata.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		values := FromContext(ctx)
		pairs := make([]string, 0, 10)
		if values.RequestID != "" {
			pairs = append(pairs, "x-request-id", values.RequestID)
		}
		if values.CorrelationID != "" {
			pairs = append(pairs, "x-correlation-id", values.CorrelationID)
		}
		if values.IdempotencyKey != "" {
			pairs = append(pairs, "idempotency-key", values.IdempotencyKey)
		}
		if values.TestRunID != "" {
			pairs = append(pairs, "x-test-run-id", values.TestRunID)
		}
		if values.TestScenarioID != "" {
			pairs = append(pairs, "x-test-scenario", values.TestScenarioID)
		}
		if len(pairs) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryServerInterceptor extracts correlation metadata from incoming gRPC
// calls and makes it available to application and event-publishing layers.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		incoming, _ := metadata.FromIncomingContext(ctx)
		ctx = WithValues(ctx, Values{RequestID: first(incoming.Get("x-request-id")), CorrelationID: first(incoming.Get("x-correlation-id")), IdempotencyKey: first(incoming.Get("idempotency-key")), TestRunID: first(incoming.Get("x-test-run-id")), TestScenarioID: first(incoming.Get("x-test-scenario"))})
		return handler(ctx, req)
	}
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
