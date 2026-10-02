package resilience

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
)

type retryWriter struct{ messages []kafka.Message }

func (w *retryWriter) WriteMessages(_ context.Context, messages ...kafka.Message) error {
	w.messages = append(w.messages, messages...)
	return nil
}

func TestKafkaRetryQueueEnqueuePreservesSourceRecord(t *testing.T) {
	w := &retryWriter{}
	q := KafkaRetryQueue{Group: "realtime-service", Writer: w}
	err := q.Enqueue(context.Background(), kafka.Message{Key: []byte("k"), Value: []byte("v"), Partition: 2, Offset: 19, Headers: []kafka.Header{{Key: "run_id", Value: []byte("run-1")}}}, "realtime", 2, errors.New("redis unavailable"))
	if err != nil {
		t.Fatal(err)
	}
	if len(w.messages) != 1 {
		t.Fatalf("messages=%d", len(w.messages))
	}
	record, err := UnmarshalRetry(w.messages[0].Value)
	if err != nil {
		t.Fatal(err)
	}
	if record.OriginalTopic != "realtime" || string(record.OriginalValue) != "v" || record.Attempt != 2 {
		t.Fatalf("unexpected retry record: %#v", record)
	}
	if len(record.OriginalHeaders) != 1 || record.OriginalHeaders[0].Key != "run_id" || string(record.OriginalHeaders[0].Value) != "run-1" {
		t.Fatalf("source headers were not preserved: %#v", record.OriginalHeaders)
	}
}
