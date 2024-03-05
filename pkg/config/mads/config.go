package mads

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/mads"
)

func DefaultMonitoringAssignmentServerConfig() *MonitoringAssignmentServerConfig {
	return &MonitoringAssignmentServerConfig{
		Port:                      5676,
		DefaultFetchTimeout:       config_types.Duration{Duration: 30 * time.Second},
		ApiVersions:               []mads.ApiVersion{mads.API_V1},
		AssignmentRefreshInterval: config_types.Duration{Duration: 1 * time.Second},
		TlsMinVersion:             "TLSv1_2",
		TlsCipherSuites:           []string{},
	}
}

// MonitoringAssignmentServerConfig contains Monitoring Assignment Discovery Service (MADS)
// server configuration.
type MonitoringAssignmentServerConfig struct {
	config.BaseConfig

	// Port of the server that serves Monitoring Assignment Discovery Service (MADS)
	// over both grpc and http.
	Port uint32 `json:"port" envconfig:"kuma_monitoring_assignment_server_port"`
	// The default timeout for a single fetch-based discovery request, if not specified.
	DefaultFetchTimeout config_types.Duration `json:"defaultFetchTimeout" envconfig:"kuma_monitoring_assignment_server_default_fetch_timeout"`
	// Which observability apiVersions to serve
	ApiVersions []string `json:"apiVersions" envconfig:"kuma_monitoring_assignment_server_api_versions"`
	// Interval for re-generating monitoring assignments for clients connected to the Control Plane.
	AssignmentRefreshInterval config_types.Duration `json:"assignmentRefreshInterval" envconfig:"kuma_monitoring_assignment_server_assignment_refresh_interval"`
	// TlsEnabled whether tls is enabled or not
	TlsEnabled bool `json:"tlsEnabled" envconfig:"kuma_monitoring_assignment_server_tls_enabled"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert. If empty, autoconfigured from general.tlsCertFile
	TlsCertFile string `json:"tlsCertFile" envconfig:"kuma_monitoring_assignment_server_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key. If empty, autoconfigured from general.tlsKeyFile
	TlsKeyFile string `json:"tlsKeyFile" envconfig:"kuma_monitoring_assignment_server_tls_key_file"`
	// TlsMinVersion defines the minimum TLS version to be used
	TlsMinVersion string `json:"tlsMinVersion" envconfig:"kuma_monitoring_assignment_server_tls_min_version"`
	// TlsMaxVersion defines the maximum TLS version to be used
	TlsMaxVersion string `json:"tlsMaxVersion" envconfig:"kuma_monitoring_assignment_server_tls_max_version"`
	// TlsCipherSuites defines the list of ciphers to use
	TlsCipherSuites []string `json:"tlsCipherSuites" envconfig:"kuma_monitoring_assignment_server_tls_cipher_suites"`
}

var _ config.Config = &MonitoringAssignmentServerConfig{}

func (c *MonitoringAssignmentServerConfig) Validate() error {
	var errs error
	if 65535 < c.Port {
		errs = multierr.Append(errs, errors.Errorf(".Port must be in the range [0, 65535]"))
	}

	if len(c.ApiVersions) == 0 {
		errs = multierr.Append(errs, errors.Errorf(".ApiVersions must contain at least one version"))
	}

	for _, apiVersion := range c.ApiVersions {
		if apiVersion != mads.API_V1 {
			errs = multierr.Append(errs, errors.Errorf(".ApiVersions contains invalid version %s", apiVersion))
		}
	}

	if c.AssignmentRefreshInterval.Duration <= 0 {
		return errors.New(".AssignmentRefreshInterval must be positive")
	}
	if c.TlsCertFile == "" && c.TlsKeyFile != "" {
		errs = multierr.Append(errs, errors.New(".TlsCertFile cannot be empty if TlsKeyFile has been set"))
	}
	if c.TlsKeyFile == "" && c.TlsCertFile != "" {
		errs = multierr.Append(errs, errors.New(".TlsKeyFile cannot be empty if TlsCertFile has been set"))
	}
	if _, err := config_types.TLSVersion(c.TlsMinVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMinVersion"+err.Error()))
	}
	if _, err := config_types.TLSVersion(c.TlsMaxVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMaxVersion"+err.Error()))
	}
	if _, err := config_types.TLSCiphers(c.TlsCipherSuites); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsCipherSuites"+err.Error()))
	}
	return errs
}

// VersionIsEnabled checks whether a MADS version has been enabled and should be served.
func (c *MonitoringAssignmentServerConfig) VersionIsEnabled(apiVersion mads.ApiVersion) bool {
	for _, version := range c.ApiVersions {
		if apiVersion == version {
			return true
		}
	}
	return false
}
