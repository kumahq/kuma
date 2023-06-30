package tracing

type Config struct {
	OpenTelemetry OpenTelemetry `json:"openTelemetry"`
}

type OpenTelemetry struct {
	// Address of OpenTelemetry collector.
	// E.g. otel-collector:4317
	Endpoint string `json:"endpoint"`
}
