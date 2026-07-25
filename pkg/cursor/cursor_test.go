package cursor

import "testing"

type payload struct {
	Scope     string `json:"scope"`
	Window    int    `json:"window"`
	LastScore int64  `json:"last_score"`
	LastID    string `json:"last_id"`
	Limit     int    `json:"limit"`
	Page      int    `json:"page"`
}

func TestCodecRoundTrip(t *testing.T) {
	secret := "12345678901234567890123456789012"
	codec, err := NewCodec(Config{Secret: secret})
	if err != nil {
		t.Fatalf("new codec: %v", err)
	}
	want := payload{Scope: "gig", Window: 1, LastScore: 123, LastID: "abc", Limit: 10, Page: 2}
	token, err := codec.Encode(want)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var got payload
	if err := codec.Decode(token, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got != want {
		t.Fatalf("unexpected payload: %+v", got)
	}
}

func TestCodecRejectsTampering(t *testing.T) {
	secret := "12345678901234567890123456789012"
	codec, err := NewCodec(Config{Secret: secret})
	if err != nil {
		t.Fatalf("new codec: %v", err)
	}
	token, err := codec.Encode(payload{Scope: "gig", Window: 1})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	bad := token[:len(token)-1] + "A"
	var got payload
	if err := codec.Decode(bad, &got); err == nil {
		t.Fatal("expected tampering to fail")
	}
}
