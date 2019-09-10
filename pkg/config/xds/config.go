package xds

import (
	"time"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
)

var _ config.Config = &XdsServerConfig{}

// Envoy XDS server configuration
type XdsServerConfig struct {
	// Port of GRPC server that Envoy connects to
	GrpcPort int `yaml:"grpcPort" envconfig:"kuma_xds_server_grpc_port"`
	// Port of Diagnostic Server for checking health and readiness of the Control Plane
	DiagnosticsPort int `yaml:"diagnosticsPort" envconfig:"kuma_xds_server_diagnostics_port"`

	// Interval for re-genarting configuration for Dataplanes connected to the Control Plane
	DataplaneConfigurationRefreshInterval time.Duration `yaml:"dataplaneConfigurationRefreshInterval" envconfig:"kuma_xds_server_dataplane_configuration_refresh_interval"`
	// Interval for flushing status of Dataplanes connected to the Control Plane
	DataplaneStatusFlushInterval time.Duration `yaml:"dataplaneStatusFlushInterval" envconfig:"kuma_xds_server_dataplane_status_flush_interval"`
}

func (x *XdsServerConfig) Validate() error {
	if x.GrpcPort < 0 {
		return errors.New("GrpcPort cannot be negative")
	}
	if x.DiagnosticsPort < 0 {
		return errors.New("DiagnosticPort cannot be negative")
	}
	if x.DataplaneConfigurationRefreshInterval <= 0 {
		return errors.New("DataplaneConfigurationRefreshInterval must be positive")
	}
	if x.DataplaneStatusFlushInterval <= 0 {
		return errors.New("DataplaneStatusFlushInterval must be positive")
	}
	return nil
}

func DefaultXdsServerConfig() *XdsServerConfig {
	return &XdsServerConfig{
		GrpcPort:                              5678,
		DiagnosticsPort:                       5680,
		DataplaneConfigurationRefreshInterval: 1 * time.Second,
		DataplaneStatusFlushInterval:          1 * time.Second,
	}
}

type BootstrapServerConfig struct {
	// Port of Server that provides bootstrap configuration for dataplanes
	Port int `yaml:"port" envconfig:"kuma_bootstrap_server_port"`
	// Parameters of bootstrap configuration
	Params *BootstrapParamsConfig `yaml:"params"`
}

func (b *BootstrapServerConfig) Validate() error {
	if b.Port < 0 {
		return errors.New("Port cannot be negative")
	}
	if err := b.Params.Validate(); err != nil {
		return errors.Wrap(err, "Params validation failed")
	}
	return nil
}

func DefaultBootstrapServerConfig() *BootstrapServerConfig {
	return &BootstrapServerConfig{
		Port:   5682,
		Params: DefaultBootstrapParamsConfig(),
	}
}

type BootstrapParamsConfig struct {
	// Port of Envoy Admin
	AdminPort uint32 `yaml:"adminPort" envconfig:"kuma_bootstrap_server_params_admin_port"`
	// Host of XDS Server
	XdsHost string `yaml:"xdsHost" envconfig:"kuma_bootstrap_server_params_xds_host"`
	// Port of XDS Server
	XdsPort uint32 `yaml:"xdsPort" envconfig:"kuma_bootstrap_server_params_xds_port"`
}

func (b *BootstrapParamsConfig) Validate() error {
	if b.AdminPort < 0 {
		return errors.New("AdminPort cannot be negative")
	}
	if b.XdsHost == "" {
		return errors.New("XdsHost cannot be empty")
	}
	if b.XdsPort < 0 {
		return errors.New("XdsPort cannot be negative")
	}
	return nil
}

func DefaultBootstrapParamsConfig() *BootstrapParamsConfig {
	return &BootstrapParamsConfig{
		AdminPort: 0, // by default, turn off Admin interface of Envoy
		XdsHost:   "127.0.0.1",
		XdsPort:   5678,
	}
}
