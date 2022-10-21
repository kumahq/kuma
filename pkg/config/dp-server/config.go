package dp_server

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

var _ config.Config = &DpServerConfig{}

// Dataplane Server configuration that servers API like Bootstrap/XDS.
type DpServerConfig struct {
	// Port of the DP Server
	Port int `yaml:"port" envconfig:"kuma_dp_server_port"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert. If empty, autoconfigured from general.tlsCertFile
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_dp_server_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key. If empty, autoconfigured from general.tlsKeyFile
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_dp_server_tls_key_file"`
	// TlsMinVersion defines the minimum TLS version to be used
	TlsMinVersion string `yaml:"tlsMinVersion" envconfig:"kuma_dp_server_tls_min_version"`
	// TlsMaxVersion defines the maximum TLS version to be used
	TlsMaxVersion string `yaml:"tlsMaxVersion" envconfig:"kuma_dp_server_tls_max_version"`
	// TlsCipherSuites defines the list of ciphers to use
	TlsCipherSuites []string `yaml:"tlsCipherSuites" envconfig:"kuma_dp_server_tls_cipher_suites"`
	// Auth defines an authentication configuration for the DP Server
	Auth DpServerAuthConfig `yaml:"auth"`
	// Hds defines a Health Discovery Service configuration
	Hds *HdsConfig `yaml:"hds"`
}

type DpServerAuthType string

const (
	DpServerAuthServiceAccountToken = "serviceAccountToken"
	DpServerAuthDpToken             = "dpToken"
	DpServerAuthNone                = "none"
)

// Authentication configuration for Dataplane Server
type DpServerAuthConfig struct {
	// Type of authentication. Available values: "serviceAccountToken", "dpToken", "none".
	// If empty, autoconfigured based on the environment - "serviceAccountToken" on Kubernetes, "dpToken" on Universal.
	Type string `yaml:"type" envconfig:"kuma_dp_server_auth_type"`
	// UseTokenPath define if should use config for ads with path to token that can be reloaded.
	UseTokenPath bool `yaml:"useTokenPath" envconfig:"kuma_dp_server_auth_use_token_path"`
}

func (a *DpServerAuthConfig) Validate() error {
	if a.Type != "" && a.Type != DpServerAuthNone && a.Type != DpServerAuthDpToken && a.Type != DpServerAuthServiceAccountToken {
		return errors.Errorf("Type is invalid. Available values are: %q, %q, %q", DpServerAuthDpToken, DpServerAuthServiceAccountToken, DpServerAuthNone)
	}
	return nil
}

func (a *DpServerConfig) Sanitize() {
}

func (a *DpServerConfig) Validate() (errs error) {
	if a.Port < 0 {
		errs = multierr.Append(errs, errors.New(".Port cannot be negative"))
	}
	if err := a.Auth.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrap(err, ".Auth is invalid"))
	}
	if _, err := config_types.TLSVersion(a.TlsMinVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMinVersion"+err.Error()))
	}
	if _, err := config_types.TLSVersion(a.TlsMaxVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMaxVersion"+err.Error()))
	}
	if _, err := config_types.TLSCiphers(a.TlsCipherSuites); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsCipherSuites"+err.Error()))
	}
	return
}

func DefaultDpServerConfig() *DpServerConfig {
	return &DpServerConfig{
		Port: 5678,
		Auth: DpServerAuthConfig{
			Type:         "", // autoconfigured from the environment
			UseTokenPath: false,
		},
		Hds:             DefaultHdsConfig(),
		TlsMinVersion:   "TLSv1_2",
		TlsCipherSuites: []string{},
	}
}

func DefaultHdsConfig() *HdsConfig {
	return &HdsConfig{
		Enabled:         true,
		Interval:        5 * time.Second,
		RefreshInterval: 10 * time.Second,
		CheckDefaults: &HdsCheck{
			Timeout:            2 * time.Second,
			Interval:           1 * time.Second,
			NoTrafficInterval:  1 * time.Second,
			HealthyThreshold:   1,
			UnhealthyThreshold: 1,
		},
	}
}

type HdsConfig struct {
	// Enabled if true then Envoy will actively check application's ports, but only on Universal.
	// On Kubernetes this feature disabled for now regardless the flag value
	Enabled bool `yaml:"enabled" envconfig:"kuma_dp_server_hds_enabled"`
	// Interval for Envoy to send statuses for HealthChecks
	Interval time.Duration `yaml:"interval" envconfig:"kuma_dp_server_hds_interval"`
	// RefreshInterval is an interval for re-genarting configuration for Dataplanes connected to the Control Plane
	RefreshInterval time.Duration `yaml:"refreshInterval" envconfig:"kuma_dp_server_hds_refresh_interval"`
	// CheckDefaults defines a HealthCheck configuration
	CheckDefaults *HdsCheck `yaml:"checkDefaults"`
}

func (h *HdsConfig) Sanitize() {
}

func (h *HdsConfig) Validate() error {
	if h.Interval <= 0 {
		return errors.New("Interval must be greater than 0s")
	}
	if err := h.CheckDefaults.Validate(); err != nil {
		return errors.Wrap(err, "Check is invalid")
	}
	return nil
}

type HdsCheck struct {
	// Timeout is a time to wait for a health check response. If the timeout is reached the
	// health check attempt will be considered a failure.
	Timeout time.Duration `yaml:"timeout" envconfig:"kuma_dp_server_hds_check_timeout"`
	// Interval between health checks.
	Interval time.Duration `yaml:"interval" envconfig:"kuma_dp_server_hds_check_interval"`
	// NoTrafficInterval is a special health check interval that is used when a cluster has
	// never had traffic routed to it.
	NoTrafficInterval time.Duration `yaml:"noTrafficInterval" envconfig:"kuma_dp_server_hds_check_no_traffic_interval"`
	// HealthyThreshold is a number of healthy health checks required before a host is marked
	// healthy.
	HealthyThreshold uint32 `yaml:"healthyThreshold" envconfig:"kuma_dp_server_hds_check_healthy_threshold"`
	// UnhealthyThreshold is a number of unhealthy health checks required before a host is marked
	// unhealthy.
	UnhealthyThreshold uint32 `yaml:"unhealthyThreshold" envconfig:"kuma_dp_server_hds_check_unhealthy_threshold"`
}

func (h *HdsCheck) Sanitize() {
}

func (h *HdsCheck) Validate() error {
	if h.Timeout <= 0 {
		return errors.New("Timeout must be greater than 0s")
	}
	if h.Interval <= 0 {
		return errors.New("Interval must be greater than 0s")
	}
	if h.NoTrafficInterval <= 0 {
		return errors.New("NoTrafficInterval must be greater than 0s")
	}
	return nil
}
