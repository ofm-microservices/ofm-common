package cursor

import "errors"

var (
	// ErrEmptySecret indicates that the cursor codec was not configured with a secret key.
	ErrEmptySecret = errors.New("cursor secret is empty")
	// ErrInvalidToken indicates that the cursor token could not be decrypted or decoded.
	ErrInvalidToken = errors.New("invalid cursor token")
)
