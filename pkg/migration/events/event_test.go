package events

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEnvelopeRoundTrip(t *testing.T) {
	want := Envelope{
		EventID: "event-1", EventType: RegistrationStarted, Operation: "created", SchemaVersion: 1,
		AggregateType: "registration", AggregateID: "session-1", AggregateVersion: 1,
		SourceService: "registration-saga-service", OccurredAt: time.Unix(1, 0).UTC(),
		Payload: json.RawMessage(`{"session_id":"session-1"}`),
	}
	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got Envelope
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.EventType != want.EventType || got.Operation != want.Operation || got.AggregateID != want.AggregateID || string(got.Payload) != string(want.Payload) {
		t.Fatalf("round-trip mismatch: %#v", got)
	}
}
