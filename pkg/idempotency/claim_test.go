package idempotency

import "testing"

func TestDecodeRequiresCanonicalIdentity(t *testing.T) {
	if _, err := Decode([]byte(`{"event_id":"e","event_type":"t","aggregate_id":"a"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := Decode([]byte(`{"event_id":"e"}`)); err == nil {
		t.Fatal("expected incomplete identity error")
	}
}
