package cursor

// Codec encrypts and decrypts opaque cursor payloads.
type Codec interface {
	Encode(value any) (string, error)
	Decode(token string, value any) error
}

// Config holds the inputs required to construct an opaque cursor codec.
type Config struct {
	Secret string
}
