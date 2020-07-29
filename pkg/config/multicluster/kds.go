package multicluster

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
)

type KdsServerConfig struct {
	// Port of a gRPC server that serves Kuma Discovery Service (KDS).
	GrpcPort uint32 `yaml:"grpcPort" envconfig:"kuma_multicluster_global_kds_grpc_port"`
	// Interval for refreshing state of the world
	RefreshInterval time.Duration `yaml:"refreshInterval" envconfig:"kuma_multicluster_global_kds_refresh_interval"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert.
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_multicluster_global_kds_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key.
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_multicluster_global_kds_tls_key_file"`
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
	// Interval for refreshing state of the world
	RefreshInterval time.Duration `yaml:"refreshInterval" envconfig:"kuma_multicluster_remote_kds_refresh_interval"`
	// RootCAFile defines a path to a file with PEM-encoded Root CA. Client will verify server by using it.
	RootCAFile string `yaml:"rootCaFile" envconfig:"kuma_multicluster_remote_kds_root_ca_file"`
}

var _ config.Config = &KdsClientConfig{}

func (k KdsClientConfig) Sanitize() {
}

func (k KdsClientConfig) Validate() error {
	return nil
}
