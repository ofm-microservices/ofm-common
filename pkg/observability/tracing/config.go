package tracing

// Config controls how the shared OpenTelemetry tracer exporter behaves.
type Config struct {
	Enabled        bool
	ServiceName    string
	ServiceEnv     string
	ServiceVersion string
	Endpoint       string
	Protocol       string
	SampleRatio    float64
}

// DefaultConfig returns a conservative local tracing configuration.
func DefaultConfig(service, env string) Config {
	return Config{
		Enabled:        true,
		ServiceName:    service,
		ServiceEnv:     env,
		ServiceVersion: "dev",
		Endpoint:       "http://127.0.0.1:9099",
		Protocol:       "http/protobuf",
		SampleRatio:    1.0,
	}
}
