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

func TestDecodeOrFingerprintUsesRecoveryCommandIDWithoutAggregateID(t *testing.T) {
	event := DecodeOrFingerprint("migration.recovery.commands.gig", []byte(`{"command_id":"00000000-0000-0000-0000-000000000001","event_type":"recovery.gig.post","aggregate_type":"gig"}`))
	if event.EventID != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("expected command identity, got %q", event.EventID)
	}
}
