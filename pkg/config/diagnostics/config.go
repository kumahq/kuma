package diagnostics

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

type DiagnosticsConfig struct {
	// Port of Diagnostic Server for checking health and readiness of the Control Plane
	ServerPort int `yaml:"serverPort" envconfig:"kuma_diagnostics_server_port"`
	// If true, enables https://golang.org/pkg/net/http/pprof/ debug endpoints
	DebugEndpoints bool `yaml:"debugEndpoints" envconfig:"kuma_diagnostics_debug_endpoints"`
}

var _ config.Config = &DiagnosticsConfig{}

func (d *DiagnosticsConfig) Sanitize() {
}

func (d *DiagnosticsConfig) Validate() error {
	if d.ServerPort < 0 {
		return errors.New("DiagnosticPort cannot be negative")
	}
	return nil
}

func DefaultDiagnosticsConfig() *DiagnosticsConfig {
	return &DiagnosticsConfig{
		ServerPort:     5680,
		DebugEndpoints: false,
	}
}
