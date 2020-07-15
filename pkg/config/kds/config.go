package kds

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
)

func DefaultKdsConfig() *KdsConfig {
	return &KdsConfig{
		Server: &KdsServerConfig{
			GrpcPort:        5685,
			RefreshInterval: 1 * time.Second,
		},
		Client: &KdsClientConfig{},
	}
}

// Kuma Discovery Service (KDS) configuration.
type KdsConfig struct {
	// Server stores configuration for the KDS server part.
	Server *KdsServerConfig `yaml:"server"`
	// Client stores configuration for the KDS client part.
	Client *KdsClientConfig `yaml:"client"`
}

var _ config.Config = &KdsConfig{}

func (c *KdsConfig) Sanitize() {
}

func (c *KdsConfig) Validate() (errs error) {
	if err := c.Server.Validate(); err != nil {
		return errors.Wrap(err, "Server validation failed")
	}
	if err := c.Client.Validate(); err != nil {
		return errors.Wrap(err, "Client validation failed")
	}
	return nil
}

type KdsServerConfig struct {
	// Port of a gRPC server that serves Kuma Discovery Service (KDS).
	GrpcPort uint32 `yaml:"grpcPort" envconfig:"kuma_kds_server_grpc_port"`
	// Interval for refreshing state of the world
	RefreshInterval time.Duration `yaml:"refreshInterval" envconfig:"kuma_kds_server_refresh_interval"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert.
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_kds_server_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key.
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_kds_server_tls_key_file"`
}

var _ config.Config = &KdsServerConfig{}

func (c *KdsServerConfig) Sanitize() {
}

func (c *KdsServerConfig) Validate() (errs error) {
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

type KdsClientConfig struct {
	// RootCAFile defines a path to a file with PEM-encoded Root CA. Client will verify server by using it.
	RootCAFile string `yaml:"rootCaFile" envconfig:"kuma_kds_client_root_ca_file"`
}

var _ config.Config = &KdsClientConfig{}

func (k KdsClientConfig) Sanitize() {
}

func (k KdsClientConfig) Validate() error {
	return nil
}
