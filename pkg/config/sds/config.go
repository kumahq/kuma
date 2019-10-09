package sds

import (
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
)

func DefaultSdsServerConfig() *SdsServerConfig {
	return &SdsServerConfig{
		GrpcPort: 5677,
	}
}

// Envoy SDS server configuration
type SdsServerConfig struct {
	// Port of GRPC server that Envoy connects to
	GrpcPort int `yaml:"grpcPort" envconfig:"kuma_sds_server_grpc_port"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert.
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_sds_server_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key.
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_sds_server_tls_key_file"`
}

var _ config.Config = &SdsServerConfig{}

func (c *SdsServerConfig) Validate() error {
	if c.GrpcPort < 0 {
		return errors.New("GrpcPort cannot be negative")
	}
	if c.TlsCertFile == "" && c.TlsKeyFile != "" {
		return errors.New("TlsCertFile cannot be empty if TlsKeyFile has been set")
	}
	if c.TlsKeyFile == "" && c.TlsCertFile != "" {
		return errors.New("TlsKeyFile cannot be empty if TlsCertFile has been set")
	}
	return nil
}

func DefaultInitialTokenServerConfig() *InitialTokenServerConfig {
	return &InitialTokenServerConfig{
		LocalHttpPort: 5679,
	}
}

// Initial Token Server configuration
type InitialTokenServerConfig struct {
	// Port of server listening on HTTP bind to localhost
	LocalHttpPort int `yaml:"localHttpPort" envconfig:"kuma_initial_token_server_local_http_port"`
}

var _ config.Config = &InitialTokenServerConfig{}

func (i *InitialTokenServerConfig) Validate() error {
	if i.LocalHttpPort < 0 {
		return errors.New("LocalHttpPort cannot be negative")
	}
	return nil
}
