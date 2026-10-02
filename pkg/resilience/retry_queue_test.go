package resilience

import (
	"testing"
	"time"
)

func TestRetryTopicNames(t *testing.T) {
	if got := RetryTopic(" realtime-service "); got != "realtime-service.retry" {
		t.Fatalf("retry topic=%q", got)
	}
	if got := DeadLetterTopic("realtime-service"); got != "realtime-service.dead-letter" {
		t.Fatalf("dead-letter topic=%q", got)
	}
}

func TestRetryEnvelopeRoundTrip(t *testing.T) {
	want := RetryEnvelope{
		OriginalValue: []byte(`{"event_id":"e1"}`), OriginalTopic: "realtime",
		OriginalPartition: 2, OriginalOffset: 17, Attempt: 3,
		AvailableAt: time.Unix(100, 0).UTC(), LastError: "temporary",
	}
	payload, err := MarshalRetry(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalRetry(payload)
	if err != nil {
		t.Fatal(err)
	}
	if got.OriginalTopic != want.OriginalTopic || got.Attempt != want.Attempt || got.OriginalOffset != want.OriginalOffset {
		t.Fatalf("round trip mismatch: %#v", got)
	}
}

func TestMarshalRetryRejectsInvalidRecord(t *testing.T) {
	if _, err := MarshalRetry(RetryEnvelope{OriginalValue: []byte("x"), Attempt: 1}); err == nil {
		t.Fatal("expected missing topic error")
	}
	if _, err := MarshalRetry(RetryEnvelope{OriginalTopic: "x", Attempt: 1}); err == nil {
		t.Fatal("expected missing payload error")
	}
}
