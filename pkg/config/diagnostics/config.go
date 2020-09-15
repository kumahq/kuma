package diagnostics

import "github.com/kumahq/kuma/pkg/config"

type DiagnosticsConfig struct {
	// If true, enables https://golang.org/pkg/net/http/pprof/ debug endpoints
	DebugEndpoints bool `yaml:"debugEndpoints" envconfig:"kuma_diagnostics_debug_endpoints"`
}

var _ config.Config = &DiagnosticsConfig{}

func (d *DiagnosticsConfig) Sanitize() {
}

func (d *DiagnosticsConfig) Validate() error {
	return nil
}

func DefaultDiagnosticsConfig() *DiagnosticsConfig {
	return &DiagnosticsConfig{
		DebugEndpoints: false,
	}
}
