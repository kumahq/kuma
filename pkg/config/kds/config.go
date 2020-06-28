package kds

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/Kong/kuma/pkg/config"
)

func DefaultKumaDiscoveryServerConfig() *KumaDiscoveryServerConfig {
	return &KumaDiscoveryServerConfig{
		GrpcPort:        5685,
		RefreshInterval: 1 * time.Second,
	}
}

// Kuma Discovery Service (KDS) server configuration.
type KumaDiscoveryServerConfig struct {
	// Port of a gRPC server that serves Kuma Discovery Service (KDS).
	GrpcPort uint32 `yaml:"grpcPort" envconfig:"kuma_kds_server_grpc_port"`
	// Interval for refreshing state of the world
	RefreshInterval time.Duration `yaml:"refreshInterval" envconfig:"kuma_kds_server_refresh_interval"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert.
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_kds_server_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key.
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_kds_server_tls_key_file"`
}

var _ config.Config = &KumaDiscoveryServerConfig{}

func (c *KumaDiscoveryServerConfig) Sanitize() {
}

func (c *KumaDiscoveryServerConfig) Validate() (errs error) {
	if c.GrpcPort > 65535 {
		errs = multierr.Append(errs, errors.Errorf(".GrpcPort must be in the range [0, 65535]"))
	}
	if c.RefreshInterval <= 0 {
		return errors.New(".RefreshInterval must be positive")
	}
	if c.TlsCertFile == "" && c.TlsKeyFile != "" {
		return errors.New("TlsCertFile cannot be empty if TlsKeyFile has been set")
	}
	if c.TlsKeyFile == "" && c.TlsCertFile != "" {
		return errors.New("TlsKeyFile cannot be empty if TlsCertFile has been set")
	}
	return
}
