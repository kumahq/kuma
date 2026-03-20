package otelenv

import (
	"os"
	"time"

	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

// Config represents the OpenTelemetry environment configuration
type Config struct {
	PipeEnabled bool
	Shared      Layer
	Traces      Layer
	Logs        Layer
	Metrics     Layer
	Inventory   *core_xds.OtelBootstrapInventory
}

// Layer represents a single layer of OpenTelemetry configuration
type Layer struct {
	Endpoint          *string
	Protocol          *string
	Headers           *string
	Timeout           *string
	Compression       *string
	Insecure          *string
	Certificate       *string
	ClientCertificate *string
	ClientKey         *string
}

// SignalRuntime represents the runtime configuration for a specific signal
type SignalRuntime struct {
	Enabled        bool
	BlockedReasons []string
	HTTPPath       string
	Transport      ExporterTransport
}

// BackendRuntime represents the complete runtime configuration for all signals
type BackendRuntime struct {
	Traces  SignalRuntime
	Logs    SignalRuntime
	Metrics SignalRuntime
}

// ExporterTransport represents the transport configuration for exporting telemetry
type ExporterTransport struct {
	Protocol          core_xds.OtelProtocol
	Endpoint          string
	UseTLS            *bool
	Headers           map[string]string
	Compression       string
	Timeout           time.Duration
	Certificate       string
	ClientCertificate string
	ClientKey         string // #nosec G117 -- OTLP mTLS config field, not a hardcoded key
}

// exporterOverride represents overrides that can be applied to an exporter configuration
type exporterOverride struct {
	Endpoint          *string
	UseTLS            *bool
	HTTPPath          *string
	Headers           map[string]string
	HeadersPresent    bool
	Compression       *string
	Timeout           *time.Duration
	Certificate       *string
	ClientCertificate *string
	ClientKey         *string // #nosec G117 -- OTLP mTLS config field, not a hardcoded key
}

// EnvReader defines the interface for reading environment variables
type EnvReader interface {
	Lookup(string) (string, bool)
}

// OSEnvReader implements EnvReader using os.LookupEnv
type OSEnvReader struct{}

func (OSEnvReader) Lookup(name string) (string, bool) {
	return os.LookupEnv(name)
}

// MapEnvReader implements EnvReader using a map
type MapEnvReader map[string]string

func (m MapEnvReader) Lookup(name string) (string, bool) {
	value, ok := m[name]
	return value, ok
}
