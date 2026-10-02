package events

import (
	"encoding/json"
	"testing"
)

func TestNewCreatesUUIDv7Envelope(t *testing.T) {
	envelope, err := New("registration.code_sent", json.RawMessage(`{"session_id":"session-1"}`))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if envelope.EventID == "" || envelope.SchemaVersion != 1 {
		t.Fatalf("unexpected envelope metadata: %+v", envelope)
	}
	if _, err := json.Marshal(envelope); err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
}

func TestEnvelopeRoundTrip(t *testing.T) {
	original, err := New("gig.created", json.RawMessage(`{"gig_id":"gig-1"}`))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	decoded, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.EventID != original.EventID || decoded.EventType != original.EventType {
		t.Fatalf("round-trip metadata mismatch: got %+v want %+v", decoded, original)
	}
}

func TestNewRejectsInvalidPayload(t *testing.T) {
	if _, err := New("gig.created", json.RawMessage(`not-json`)); err == nil {
		t.Fatal("New() accepted invalid payload")
	}
}

func TestWrapAndUnwrap(t *testing.T) {
	wrapper, envelope, err := Wrap("gig.created", []byte(`{"gig_id":"gig-1"}`))
	if err != nil {
		t.Fatalf("Wrap() error = %v", err)
	}
	payload, decoded, err := Unwrap(wrapper)
	if err != nil || decoded == nil {
		t.Fatalf("Unwrap() error = %v decoded=%+v", err, decoded)
	}
	if decoded.EventID != envelope.EventID || string(payload) != `{"gig_id":"gig-1"}` {
		t.Fatalf("unexpected unwrapped event: %+v %s", decoded, payload)
	}
}

func TestWrapDoesNotDoubleWrap(t *testing.T) {
	original, err := New("gig.created", json.RawMessage(`{"gig_id":"gig-1"}`))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	wrapped, decoded, err := Wrap("other.topic", data)
	if err != nil || decoded.EventID != original.EventID || string(wrapped) != string(data) {
		t.Fatalf("event was double-wrapped: err=%v decoded=%+v", err, decoded)
	}
}
