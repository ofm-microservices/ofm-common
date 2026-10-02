package resilience

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// RetryEnvelope is the transport-neutral record placed on a service retry
// topic. The original Kafka offset is retained for diagnostics only; the
// source offset is committed after this record is durably published.
type RetryEnvelope struct {
	OriginalKey       []byte        `json:"original_key,omitempty"`
	OriginalValue     []byte        `json:"original_value"`
	OriginalTopic     string        `json:"original_topic"`
	OriginalPartition int           `json:"original_partition"`
	OriginalOffset    int64         `json:"original_offset"`
	OriginalHeaders   []RetryHeader `json:"original_headers,omitempty"`
	Attempt           int           `json:"attempt"`
	AvailableAt       time.Time     `json:"available_at"`
	LastError         string        `json:"last_error"`
}

// RetryHeader preserves source metadata needed when a record is republished.
type RetryHeader struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

// RetryTopic returns the service-scoped Kafka topic used for delayed retry.
func RetryTopic(group string) string {
	return strings.TrimSpace(group) + ".retry"
}

// DeadLetterTopic returns the service-scoped Kafka topic for exhausted or
// permanently invalid messages.
func DeadLetterTopic(group string) string {
	return strings.TrimSpace(group) + ".dead-letter"
}

// MarshalRetry serializes a retry record and validates its required source
// topic and payload fields.
func MarshalRetry(record RetryEnvelope) ([]byte, error) {
	if strings.TrimSpace(record.OriginalTopic) == "" {
		return nil, fmt.Errorf("retry source topic is required")
	}
	if len(record.OriginalValue) == 0 {
		return nil, fmt.Errorf("retry payload is required")
	}
	if record.Attempt < 1 {
		return nil, fmt.Errorf("retry attempt must be positive")
	}
	return json.Marshal(record)
}

// UnmarshalRetry decodes a retry record produced by MarshalRetry.
func UnmarshalRetry(payload []byte) (RetryEnvelope, error) {
	var record RetryEnvelope
	if err := json.Unmarshal(payload, &record); err != nil {
		return RetryEnvelope{}, fmt.Errorf("decode retry envelope: %w", err)
	}
	if _, err := MarshalRetry(record); err != nil {
		return RetryEnvelope{}, err
	}
	return record, nil
}
