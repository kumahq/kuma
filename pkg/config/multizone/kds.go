package multizone

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
)

type KdsServerConfig struct {
	// Port of a gRPC server that serves Kuma Discovery Service (KDS).
	GrpcPort uint32 `yaml:"grpcPort" envconfig:"kuma_multizone_global_kds_grpc_port"`
	// Interval for refreshing state of the world
	RefreshInterval time.Duration `yaml:"refreshInterval" envconfig:"kuma_multizone_global_kds_refresh_interval"`
	// Interval for flushing Zone Insights (stats of multi-zone communication)
	ZoneInsightFlushInterval time.Duration `yaml:"zoneInsightFlushInterval" envconfig:"kuma_multizone_global_kds_zone_insight_flush_interval"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert.
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_multizone_global_kds_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key.
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_multizone_global_kds_tls_key_file"`
	// MaxMsgSize defines a maximum size of the message that is exchanged using KDS.
	// In practice this means a limit on full list of one resource type.
	MaxMsgSize uint32 `yaml:"maxMsgSize" envconfig:"kuma_multizone_global_kds_max_msg_size"`
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
	if c.ZoneInsightFlushInterval <= 0 {
		return errors.New(".ZoneInsightFlushInterval must be positive")
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
	RefreshInterval time.Duration `yaml:"refreshInterval" envconfig:"kuma_multizone_zone_kds_refresh_interval"`
	// RootCAFile defines a path to a file with PEM-encoded Root CA. Client will verify the server by using it.
	RootCAFile string `yaml:"rootCaFile" envconfig:"kuma_multizone_zone_kds_root_ca_file"`
	// MaxMsgSize defines a maximum size of the message that is exchanged using KDS.
	// In practice this means a limit on full list of one resource type.
	MaxMsgSize uint32 `yaml:"maxMsgSize" envconfig:"kuma_multizone_zone_kds_max_msg_size"`
}

var _ config.Config = &KdsClientConfig{}

func (k KdsClientConfig) Sanitize() {
}

func (k KdsClientConfig) Validate() error {
	return nil
}
