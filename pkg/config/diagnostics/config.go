package diagnostics

import (
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

type DiagnosticsConfig struct {
	config.BaseConfig

	// Port of Diagnostic Server for checking health and readiness of the Control Plane
	ServerPort uint32 `json:"serverPort" envconfig:"kuma_diagnostics_server_port"`
	// If true, enables https://golang.org/pkg/net/http/pprof/ debug endpoints
	DebugEndpoints bool `json:"debugEndpoints" envconfig:"kuma_diagnostics_debug_endpoints"`
	// TlsEnabled whether tls is enabled or not
	TlsEnabled bool `json:"tlsEnabled" envconfig:"kuma_diagnostics_tls_enabled"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert. If empty, autoconfigured from general.tlsCertFile
	TlsCertFile string `json:"tlsCertFile" envconfig:"kuma_diagnostics_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key. If empty, autoconfigured from general.tlsKeyFile
	TlsKeyFile string `json:"tlsKeyFile" envconfig:"kuma_diagnostics_tls_key_file"`
	// TlsMinVersion defines the minimum TLS version to be used
	TlsMinVersion string `json:"tlsMinVersion" envconfig:"kuma_diagnostics_tls_min_version"`
	// TlsMaxVersion defines the maximum TLS version to be used
	TlsMaxVersion string `json:"tlsMaxVersion" envconfig:"kuma_diagnostics_tls_max_version"`
	// TlsCipherSuites defines the list of ciphers to use
	TlsCipherSuites []string `json:"tlsCipherSuites" envconfig:"kuma_diagnostics_tls_cipher_suites"`
}

var _ config.Config = &DiagnosticsConfig{}

func (d *DiagnosticsConfig) Validate() error {
	var errs error
	if d.TlsCertFile == "" && d.TlsKeyFile != "" {
		errs = multierr.Append(errs, errors.New(".TlsCertFile cannot be empty if TlsKeyFile has been set"))
	}
	if d.TlsKeyFile == "" && d.TlsCertFile != "" {
		errs = multierr.Append(errs, errors.New(".TlsKeyFile cannot be empty if TlsCertFile has been set"))
	}
	if _, err := config_types.TLSVersion(d.TlsMinVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMinVersion"+err.Error()))
	}
	if _, err := config_types.TLSVersion(d.TlsMaxVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMaxVersion"+err.Error()))
	}
	if _, err := config_types.TLSCiphers(d.TlsCipherSuites); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsCipherSuites"+err.Error()))
	}
	return errs
}

func DefaultDiagnosticsConfig() *DiagnosticsConfig {
	return &DiagnosticsConfig{
		ServerPort:      5680,
		DebugEndpoints:  false,
		TlsEnabled:      false,
		TlsMinVersion:   "TLSv1_2",
		TlsCipherSuites: []string{},
	}
}
