package jwt

import "time"

// Claims is the shared access-token claim set used by OFM services.
type Claims struct {
	Subject   string   `json:"sub"`
	Issuer    string   `json:"iss,omitempty"`
	ExpiresAt int64    `json:"exp,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	Email     string   `json:"email,omitempty"`
	Username  string   `json:"username,omitempty"`
	Roles     []string `json:"roles,omitempty"`
}

// Verifier validates bearer access tokens.
type Verifier interface {
	Validate(token string) (*Claims, error)
}

// Signer produces bearer access tokens from shared claims.
type Signer interface {
	Sign(claims Claims) (string, error)
}

// Config holds the inputs required to validate a JWT.
type Config struct {
	Secret    string
	PublicKey string
	Now       func() time.Time
}
