// Package realtime contains the client-facing notification contract shared by
// services that publish user-visible asynchronous outcomes.
package realtime

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metadata"
)

// Notification is the stable envelope sent through the realtime fanout topic.
// It intentionally contains business outcome metadata, not CDC, recovery,
// projection, consumer-offset, or outbox implementation details.
type Notification struct {
	EventID       string          `json:"event_id,omitempty"`
	OperationID   string          `json:"operation_id,omitempty"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	AggregateType string          `json:"aggregate_type,omitempty"`
	AggregateID   string          `json:"aggregate_id,omitempty"`
	Type          string          `json:"type"`
	Status        string          `json:"status,omitempty"`
	ErrorCode     string          `json:"error_code,omitempty"`
	Retryable     *bool           `json:"retryable,omitempty"`
	OccurredAt    string          `json:"occurred_at,omitempty"`
	TestRunID     string          `json:"test_run_id,omitempty"`
	Payload       json.RawMessage `json:"payload,omitempty"`
	UserID        string          `json:"user_id,omitempty"`
	DeliveryScope string          `json:"delivery_scope,omitempty"`
	ConnectionID  string          `json:"connection_id,omitempty"`
}

// Status values describe the outcome of an asynchronous business operation.
const (
	StatusAccepted  = "accepted"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// New creates a notification with correlation values inherited from the
// current request context. Callers may override OperationID when the command
// already has a durable idempotency key.
func New(ctx context.Context, eventType, aggregateType, aggregateID, userID string, payload json.RawMessage) Notification {
	values := metadata.FromContext(ctx)
	eventID := uuid.Must(uuid.NewV7()).String()
	operationID := values.IdempotencyKey
	if strings.TrimSpace(operationID) == "" {
		operationID = values.RequestID
	}
	if strings.TrimSpace(operationID) == "" {
		operationID = eventID
	}
	return Notification{
		EventID: eventID, OperationID: operationID, CorrelationID: values.CorrelationID,
		AggregateType: aggregateType, AggregateID: aggregateID, Type: eventType,
		Status: StatusAccepted, OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
		TestRunID: values.TestRunID, Payload: payload, UserID: userID, DeliveryScope: "user",
	}
}

// WithOutcome derives a terminal notification while preserving the operation
// identity from the command that started the asynchronous workflow.
func WithOutcome(ctx context.Context, eventType, aggregateType, aggregateID, userID, operationID, correlationID, status, errorCode string, retryable *bool, payload json.RawMessage) Notification {
	notification := New(ctx, eventType, aggregateType, aggregateID, userID, payload)
	if strings.TrimSpace(operationID) != "" {
		notification.OperationID = strings.TrimSpace(operationID)
	}
	if strings.TrimSpace(correlationID) != "" {
		notification.CorrelationID = strings.TrimSpace(correlationID)
	}
	notification.Status = status
	notification.ErrorCode = strings.TrimSpace(errorCode)
	notification.Retryable = retryable
	return notification
}
