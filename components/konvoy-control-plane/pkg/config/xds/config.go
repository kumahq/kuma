package xds

import (
	"errors"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
)

var _ config.Config = &XdsServerConfig{}

// Envoy XDS server configuration
type XdsServerConfig struct {
	// Port of GRPC server that Envoy connects to
	GrpcPort int `yaml:"grpcPort" envconfig:"konvoy_xds_server_grpc_port"`
	// Port of HTTP Server for retrieving xDS data in REST way
	HttpPort int `yaml:"httpPort" envconfig:"konvoy_xds_server_http_port"`
	// Port of Diagnostic Server for checking health and readiness of the Control Plane
	DiagnosticsPort int `yaml:"diagnosticsPort" envconfig:"konvoy_xds_server_diagnostics_port"`
}

func (x *XdsServerConfig) Validate() error {
	if x.GrpcPort < 0 {
		return errors.New("GrpcPort cannot be negative")
	}
	if x.HttpPort < 0 {
		return errors.New("HttpPort cannot be negative")
	}
	if x.DiagnosticsPort < 0 {
		return errors.New("DiagnosticPort cannot be negative")
	}
	return nil
}

func DefaultXdsServerConfig() *XdsServerConfig {
	return &XdsServerConfig{
		GrpcPort:        5678,
		HttpPort:        5679,
		DiagnosticsPort: 5680,
	}
}
