package otelenv

import (
	"os"
	"time"

	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

type Config struct {
	PipeEnabled bool
	Shared      Layer
	Traces      Layer
	Logs        Layer
	Metrics     Layer
	Inventory   core_xds.OtelBootstrapInventory
}

type Layer struct {
	Signal            core_xds.OtelSignal
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

type SignalRuntime struct {
	Enabled        bool
	BlockedReasons []string
	HTTPPath       string
	Transport      ExporterTransport
}

// HasHardBlock returns true if any blocked reason prevents actual signal usage.
// Soft blocks (EnvDisabledByPolicy, SignalOverridesBlocked) are informational -
// the signal still works via explicit configuration.
func (r SignalRuntime) HasHardBlock() bool {
	for _, reason := range r.BlockedReasons {
		switch reason {
		case core_xds.OtelBlockedReasonEnvDisabledByPolicy,
			core_xds.OtelBlockedReasonSignalOverridesBlocked:
			continue
		default:
			return true
		}
	}
	return false
}

type BackendRuntime struct {
	Traces  SignalRuntime
	Logs    SignalRuntime
	Metrics SignalRuntime
}

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

// EnvReader abstracts environment variable lookup for testability.
type EnvReader interface {
	Lookup(string) (string, bool)
}

type OSEnvReader struct{}

func (OSEnvReader) Lookup(name string) (string, bool) {
	return os.LookupEnv(name)
}

type MapEnvReader map[string]string

func (m MapEnvReader) Lookup(name string) (string, bool) {
	value, ok := m[name]
	return value, ok
}
