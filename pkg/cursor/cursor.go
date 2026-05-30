package cursor

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"golang.org/x/crypto/chacha20poly1305"
)

type codec struct {
	aead cipher.AEAD
	rand io.Reader
}

// NewCodec constructs an opaque encrypted cursor codec backed by ChaCha20-Poly1305.
func NewCodec(cfg Config) (Codec, error) {
	if strings.TrimSpace(cfg.Secret) == "" {
		return nil, ErrEmptySecret
	}
	sum := sha256.Sum256([]byte(cfg.Secret))
	key := sum[:]
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return &codec{aead: aead, rand: rand.Reader}, nil
}

// Encode serializes and encrypts the supplied payload into an opaque cursor token.
func (c *codec) Encode(value any) (string, error) {
	if c == nil || c.aead == nil {
		return "", ErrEmptySecret
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return "", ErrInvalidToken
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(c.rand, nonce); err != nil {
		return "", ErrInvalidToken
	}
	encoded := c.aead.Seal(nonce, nonce, payload, nil)
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

// Decode decrypts the token into the supplied payload target.
func (c *codec) Decode(token string, value any) error {
	if c == nil || c.aead == nil {
		return ErrEmptySecret
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return ErrInvalidToken
	}
	if len(raw) < c.aead.NonceSize() {
		return ErrInvalidToken
	}
	nonce := raw[:c.aead.NonceSize()]
	ciphertext := raw[c.aead.NonceSize():]
	payload, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return ErrInvalidToken
	}
	if err := json.Unmarshal(payload, value); err != nil {
		return ErrInvalidToken
	}
	return nil
}
