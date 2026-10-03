package resilience

import (
	"encoding/json"
	"time"
)

// DLQRecord is the transport-neutral envelope persisted or published when an
// event cannot be processed safely after retry exhaustion.
type DLQRecord struct {
	OriginalKey       []byte    `json:"original_key,omitempty"`
	OriginalValue     []byte    `json:"original_value"`
	OriginalTopic     string    `json:"original_topic"`
	OriginalPartition int       `json:"original_partition"`
	OriginalOffset    int64     `json:"original_offset"`
	EventID           string    `json:"event_id,omitempty"`
	EventType         string    `json:"event_type,omitempty"`
	AggregateType     string    `json:"aggregate_type,omitempty"`
	AggregateID       string    `json:"aggregate_id,omitempty"`
	SchemaVersion     int       `json:"schema_version,omitempty"`
	Attempts          int       `json:"attempts"`
	ErrorClass        string    `json:"error_class"`
	Error             string    `json:"error"`
	FailedAt          time.Time `json:"failed_at"`
}

// MarshalDLQ serializes a replayable DLQ envelope.
func MarshalDLQ(record DLQRecord) ([]byte, error) { return json.Marshal(record) }

// UnmarshalDLQ decodes a replayable DLQ envelope for inspection or replay.
func UnmarshalDLQ(payload []byte) (DLQRecord, error) {
	var record DLQRecord
	err := json.Unmarshal(payload, &record)
	return record, err
}
