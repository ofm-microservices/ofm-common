package jwt

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"strings"
	"time"
)

// NewSigner constructs an HMAC signer for HS256 access tokens.
func NewSigner(cfg Config) (Signer, error) {
	if strings.TrimSpace(cfg.Secret) == "" {
		return nil, ErrEmptySignerKey
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &hmacSigner{secret: []byte(cfg.Secret), now: cfg.Now}, nil
}

// NewVerifier constructs a token verifier for HS256 or RS256 access tokens.
func NewVerifier(cfg Config) (Verifier, error) {
	if strings.TrimSpace(cfg.Secret) == "" && strings.TrimSpace(cfg.PublicKey) == "" {
		return nil, ErrEmptyVerifierKey
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if strings.TrimSpace(cfg.PublicKey) != "" {
		pub, err := parseRSAPublicKey(cfg.PublicKey)
		if err != nil {
			return nil, err
		}
		return &rsaVerifier{pub: pub, now: cfg.Now}, nil
	}
	return &hmacVerifier{secret: []byte(cfg.Secret), now: cfg.Now}, nil
}

// ParseBearer extracts a bearer token from the Authorization header.
func ParseBearer(header string) (string, error) {
	fields := strings.Fields(header)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return "", ErrInvalidAuthorizationHeader
	}
	if strings.TrimSpace(fields[1]) == "" {
		return "", ErrEmptyToken
	}
	return strings.TrimSpace(fields[1]), nil
}

type hmacVerifier struct {
	secret []byte
	now    func() time.Time
}

func (v *hmacVerifier) Validate(token string) (*Claims, error) {
	headerPart, payloadPart, signaturePart, err := splitJWT(token)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, v.secret)
	if _, err := mac.Write([]byte(headerPart + "." + payloadPart)); err != nil {
		return nil, ErrInvalidToken
	}
	signature, err := base64.RawURLEncoding.DecodeString(signaturePart)
	if err != nil || !hmac.Equal(signature, mac.Sum(nil)) {
		return nil, ErrInvalidToken
	}
	return decodeClaims(payloadPart, v.now())
}

type hmacSigner struct {
	secret []byte
	now    func() time.Time
}

func (s *hmacSigner) Sign(claims Claims) (string, error) {
	if strings.TrimSpace(claims.Subject) == "" {
		return "", ErrInvalidToken
	}
	if claims.IssuedAt == 0 {
		claims.IssuedAt = s.now().UTC().Unix()
	}
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", ErrInvalidToken
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", ErrInvalidToken
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.secret)
	if _, err := mac.Write([]byte(unsigned)); err != nil {
		return "", ErrInvalidToken
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

type rsaVerifier struct {
	pub *rsa.PublicKey
	now func() time.Time
}

func (v *rsaVerifier) Validate(token string) (*Claims, error) {
	headerPart, payloadPart, signaturePart, err := splitJWT(token)
	if err != nil {
		return nil, err
	}
	signed := []byte(headerPart + "." + payloadPart)
	sig, err := base64.RawURLEncoding.DecodeString(signaturePart)
	if err != nil {
		return nil, ErrInvalidToken
	}
	hashed := sha256.Sum256(signed)
	if err := rsa.VerifyPKCS1v15(v.pub, crypto.SHA256, hashed[:], sig); err != nil {
		return nil, ErrInvalidToken
	}
	return decodeClaims(payloadPart, v.now())
}

func decodeClaims(payloadPart string, now time.Time) (*Claims, error) {
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return nil, ErrInvalidToken
	}
	if claims.ExpiresAt > 0 && now.UTC().After(time.Unix(claims.ExpiresAt, 0)) {
		return nil, ErrExpiredToken
	}
	return &claims, nil
}

func splitJWT(token string) (string, string, string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", "", "", ErrInvalidToken
	}
	return parts[0], parts[1], parts[2], nil
}

func parseRSAPublicKey(raw string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, ErrInvalidToken
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, ErrInvalidToken
	}
	key, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, ErrInvalidToken
	}
	return key, nil
}
