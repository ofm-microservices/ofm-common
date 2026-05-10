package metrics

// Config defines the operational metrics listener for a service.
type Config struct {
	Enabled bool
	Host    string
	Port    int
	Path    string
}
