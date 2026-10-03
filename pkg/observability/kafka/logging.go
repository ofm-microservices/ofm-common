package kafka

import (
	"encoding/json"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
)

// Consumed records safe metadata for a Kafka message at the broker boundary.
// Payload contents are deliberately excluded in every mode.
func Consumed(topic string, partition int, offset int64, attempt int, payload []byte) {
	lg := logging.ProcessLogger()
	if lg == nil {
		return
	}
	fields := []logging.Field{
		logging.String("transport", "kafka"), logging.String("topic", topic),
		logging.Int("partition", partition), logging.Int64("offset", offset),
		logging.Int("attempt", attempt), logging.Int("payload_bytes", len(payload)),
	}
	var envelope struct {
		EventID       string `json:"event_id"`
		Type          string `json:"event_type"`
		AggregateID   string `json:"aggregate_id"`
		CorrelationID string `json:"correlation_id"`
	}
	if json.Unmarshal(payload, &envelope) == nil {
		if envelope.EventID != "" {
			fields = append(fields, logging.String("event_id", envelope.EventID))
		}
		if envelope.Type != "" {
			fields = append(fields, logging.String("event_type", envelope.Type))
		}
		if envelope.AggregateID != "" {
			fields = append(fields, logging.String("aggregate_id", envelope.AggregateID))
		}
		if envelope.CorrelationID != "" {
			fields = append(fields, logging.String("correlation_id", envelope.CorrelationID))
		}
	}
	lg.Info("kafka event consumed", fields...)
}

// Published records safe metadata before a Kafka producer writes an envelope.
func Published(topic string, payload []byte) {
	lg := logging.ProcessLogger()
	if lg == nil {
		return
	}
	fields := []logging.Field{logging.String("transport", "kafka"), logging.String("topic", topic), logging.Int("payload_bytes", len(payload))}
	var envelope struct {
		EventID       string `json:"event_id"`
		Type          string `json:"event_type"`
		AggregateID   string `json:"aggregate_id"`
		CorrelationID string `json:"correlation_id"`
	}
	if json.Unmarshal(payload, &envelope) == nil {
		if envelope.EventID != "" {
			fields = append(fields, logging.String("event_id", envelope.EventID))
		}
		if envelope.Type != "" {
			fields = append(fields, logging.String("event_type", envelope.Type))
		}
		if envelope.AggregateID != "" {
			fields = append(fields, logging.String("aggregate_id", envelope.AggregateID))
		}
		if envelope.CorrelationID != "" {
			fields = append(fields, logging.String("correlation_id", envelope.CorrelationID))
		}
	}
	lg.Info("kafka event publish started", fields...)
}
