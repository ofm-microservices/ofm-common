package jwt

import (
	"strings"
	"testing"
	"time"
)

func TestSignerAndVerifierRoundTrip(t *testing.T) {
	signer, err := NewSigner(Config{Secret: "test-secret", Now: func() time.Time {
		return time.Unix(1700000000, 0).UTC()
	}})
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}

	token, err := signer.Sign(Claims{
		Subject:   "user-1",
		Email:     "user@example.com",
		Username:  "alex",
		Roles:     []string{"admin", "freelancer"},
		ExpiresAt: time.Unix(1700000600, 0).UTC().Unix(),
	})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	verifier, err := NewVerifier(Config{Secret: "test-secret", Now: func() time.Time {
		return time.Unix(1700000000, 0).UTC()
	}})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	claims, err := verifier.Validate(token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}

	if claims.Subject != "user-1" {
		t.Fatalf("subject = %q", claims.Subject)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("email = %q", claims.Email)
	}
	if claims.Username != "alex" {
		t.Fatalf("username = %q", claims.Username)
	}
	if len(claims.Roles) != 2 || claims.Roles[0] != "admin" || claims.Roles[1] != "freelancer" {
		t.Fatalf("roles = %#v", claims.Roles)
	}
	if claims.ExpiresAt == 0 {
		t.Fatalf("expected exp to be set")
	}
}

func TestParseBearer(t *testing.T) {
	token, err := ParseBearer("Bearer abc.def.ghi")
	if err != nil {
		t.Fatalf("parse bearer: %v", err)
	}
	if token != "abc.def.ghi" {
		t.Fatalf("token = %q", token)
	}

	if _, err := ParseBearer("bad"); err == nil || !strings.Contains(err.Error(), "invalid authorization header") {
		t.Fatalf("expected invalid authorization header error, got %v", err)
	}
}

func TestSignerRejectsEmptySecret(t *testing.T) {
	if _, err := NewSigner(Config{}); err == nil || !strings.Contains(err.Error(), "jwt signer key is empty") {
		t.Fatalf("expected empty signer key error, got %v", err)
	}
}
