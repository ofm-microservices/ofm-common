// Package kafka provides bounded correlation propagation for Kafka messages.
package kafka

import (
	"context"

	requestmetadata "github.com/ofm-microservices/ofm-common/pkg/observability/metadata"
	segmentkafka "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Headers serializes request correlation values into Kafka headers.
func Headers(ctx context.Context) []segmentkafka.Header {
	result := make([]segmentkafka.Header, 0, 5)
	for key, value := range requestmetadata.OutgoingHeaders(ctx) {
		result = append(result, segmentkafka.Header{Key: key, Value: []byte(value)})
	}
	return result
}

// Context extracts request correlation values from Kafka headers.
func Context(ctx context.Context, headers []segmentkafka.Header) context.Context {
	carrier := propagation.MapCarrier{}
	values := requestmetadata.Values{}
	for _, header := range headers {
		carrier[header.Key] = string(header.Value)
		switch header.Key {
		case "x-request-id":
			values.RequestID = string(header.Value)
		case "x-correlation-id":
			values.CorrelationID = string(header.Value)
		case "idempotency-key":
			values.IdempotencyKey = string(header.Value)
		case "x-test-run-id":
			values.TestRunID = string(header.Value)
		case "x-test-scenario":
			values.TestScenarioID = string(header.Value)
		}
	}
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	return requestmetadata.WithValues(ctx, values)
}
