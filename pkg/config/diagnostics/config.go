package diagnostics

import (
	"github.com/kumahq/kuma/pkg/config"
)

type DiagnosticsConfig struct {
	// Port of Diagnostic Server for checking health and readiness of the Control Plane
	ServerPort uint32 `json:"serverPort" envconfig:"kuma_diagnostics_server_port"`
	// If true, enables https://golang.org/pkg/net/http/pprof/ debug endpoints
	DebugEndpoints bool `json:"debugEndpoints" envconfig:"kuma_diagnostics_debug_endpoints"`
}

var _ config.Config = &DiagnosticsConfig{}

func (d *DiagnosticsConfig) Sanitize() {
}

func (d *DiagnosticsConfig) Validate() error {
	return nil
}

func DefaultDiagnosticsConfig() *DiagnosticsConfig {
	return &DiagnosticsConfig{
		ServerPort:     5680,
		DebugEndpoints: false,
	}
}
