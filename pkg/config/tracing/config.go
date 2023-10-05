package tracing

type Config struct {
	OpenTelemetry OpenTelemetry `json:"openTelemetry,omitempty"`
}

func (c Config) Validate() error {
	if err := c.OpenTelemetry.Validate(); err != nil {
		return err
	}

	return nil
}

type OpenTelemetry struct {
	Enabled bool `json:"enabled,omitempty" envconfig:"kuma_tracing_opentelemetry_enabled"`
	// This field is deprecated. Use generic OTEL_EXPORTER_OTLP_* variables
	// instead.
	// Address of OpenTelemetry collector.
	// E.g. otel-collector:4317
	Endpoint string `json:"endpoint,omitempty" envconfig:"kuma_tracing_opentelemetry_endpoint"`
}

func (c OpenTelemetry) Validate() error {
	return nil
}
