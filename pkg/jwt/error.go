package jwt

import "errors"

var (
	// ErrEmptyToken indicates that no JWT was provided.
	ErrEmptyToken = errors.New("jwt token is empty")
	// ErrInvalidAuthorizationHeader indicates a malformed bearer authorization header.
	ErrInvalidAuthorizationHeader = errors.New("invalid authorization header")
	// ErrInvalidToken indicates that the JWT could not be decoded or verified.
	ErrInvalidToken = errors.New("invalid jwt token")
	// ErrExpiredToken indicates that the JWT has already expired.
	ErrExpiredToken = errors.New("expired jwt token")
	// ErrEmptyVerifierKey indicates that the verifier was not configured with a key.
	ErrEmptyVerifierKey = errors.New("jwt verifier key is empty")
	// ErrEmptySignerKey indicates that the signer was not configured with a key.
	ErrEmptySignerKey = errors.New("jwt signer key is empty")
)
