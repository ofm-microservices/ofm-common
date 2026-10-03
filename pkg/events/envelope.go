// Package events contains transport-neutral contracts for integration events.
package events

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EventEnvelope is the common Kafka event wrapper. Producers create it once,
// before the first publish attempt, so retries preserve the same EventID.
// Payload contains the event-specific JSON representation.
type EventEnvelope struct {
	EventID             string          `json:"event_id"`
	CommandID           string          `json:"command_id,omitempty"`
	CommandMethod       string          `json:"command_method,omitempty"`
	CommandPath         string          `json:"command_path,omitempty"`
	IdempotencyKey      string          `json:"idempotency_key,omitempty"`
	TestRunID           string          `json:"test_run_id,omitempty"`
	RecoveryPrincipalID string          `json:"recovery_principal_id,omitempty"`
	RecoveryUsername    string          `json:"recovery_username,omitempty"`
	RecoveryEmail       string          `json:"recovery_email,omitempty"`
	RecoveryAccessToken string          `json:"recovery_access_token,omitempty"`
	EventType           string          `json:"event_type"`
	Operation           string          `json:"operation,omitempty"`
	OccurredAt          time.Time       `json:"occurred_at"`
	AggregateType       string          `json:"aggregate_type,omitempty"`
	AggregateID         string          `json:"aggregate_id,omitempty"`
	AggregateVersion    int64           `json:"aggregate_version,omitempty"`
	CorrelationID       string          `json:"correlation_id,omitempty"`
	CausationID         string          `json:"causation_id,omitempty"`
	SessionID           string          `json:"session_id,omitempty"`
	TraceParent         string          `json:"traceparent,omitempty"`
	TraceState          string          `json:"tracestate,omitempty"`
	SourceService       string          `json:"source_service,omitempty"`
	SchemaVersion       int             `json:"schema_version"`
	Payload             json.RawMessage `json:"payload"`
}

// New creates an envelope with a fresh UUIDv7. The caller must retain the
// returned envelope and reuse it when publishing retries.
func New(eventType string, payload json.RawMessage) (EventEnvelope, error) {
	if strings.TrimSpace(eventType) == "" {
		return EventEnvelope{}, fmt.Errorf("event type is required")
	}
	if len(payload) == 0 || !json.Valid(payload) {
		return EventEnvelope{}, fmt.Errorf("event payload must be valid JSON")
	}

	eventID, err := uuid.NewV7()
	if err != nil {
		return EventEnvelope{}, fmt.Errorf("generate event id: %w", err)
	}

	return EventEnvelope{
		EventID:       eventID.String(),
		EventType:     strings.TrimSpace(eventType),
		OccurredAt:    time.Now().UTC(),
		SchemaVersion: 1,
		Payload:       payload,
	}, nil
}

// Marshal serializes the envelope for a Kafka value.
func (e EventEnvelope) Marshal() ([]byte, error) {
	if strings.TrimSpace(e.EventID) == "" {
		return nil, fmt.Errorf("event id is required")
	}
	if strings.TrimSpace(e.EventType) == "" {
		return nil, fmt.Errorf("event type is required")
	}
	if e.SchemaVersion == 0 {
		e.SchemaVersion = 1
	}
	if len(e.Payload) == 0 || !json.Valid(e.Payload) {
		return nil, fmt.Errorf("event payload must be valid JSON")
	}
	return json.Marshal(e)
}

// Unmarshal parses a Kafka event envelope and validates its required fields.
func Unmarshal(data []byte) (EventEnvelope, error) {
	var envelope EventEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return EventEnvelope{}, fmt.Errorf("decode event envelope: %w", err)
	}
	if strings.TrimSpace(envelope.EventID) == "" {
		return EventEnvelope{}, fmt.Errorf("event id is required")
	}
	if strings.TrimSpace(envelope.EventType) == "" {
		return EventEnvelope{}, fmt.Errorf("event type is required")
	}
	if len(envelope.Payload) == 0 || !json.Valid(envelope.Payload) {
		return EventEnvelope{}, fmt.Errorf("event payload must be valid JSON")
	}
	return envelope, nil
}

// Wrap creates a Kafka envelope for an arbitrary JSON event payload. The
// topic is used as the default event type when the producer has no richer
// domain name available at the transport boundary.
func Wrap(topic string, payload []byte) ([]byte, EventEnvelope, error) {
	if existing, err := Unmarshal(payload); err == nil {
		return payload, existing, nil
	}
	envelope, err := New(topic, json.RawMessage(payload))
	if err != nil {
		return nil, EventEnvelope{}, err
	}
	data, err := envelope.Marshal()
	if err != nil {
		return nil, EventEnvelope{}, err
	}
	return data, envelope, nil
}

// Unwrap returns the application payload and envelope metadata. Raw payloads
// are accepted during the rolling migration so old Kafka records remain
// consumable while producers are upgraded.
func Unwrap(data []byte) ([]byte, *EventEnvelope, error) {
	envelope, err := Unmarshal(data)
	if err != nil {
		return data, nil, nil
	}
	return envelope.Payload, &envelope, nil
}
