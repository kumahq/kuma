package sds

import (
	"github.com/pkg/errors"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
)

func DefaultSdsServerConfig() *SdsServerConfig {
	return &SdsServerConfig{
		GrpcPort: 5677,
	}
}

// Envoy SDS server configuration
type SdsServerConfig struct {
	// Port of GRPC server that Envoy connects to
	GrpcPort int `yaml:"grpcPort" envconfig:"konvoy_sds_server_grpc_port"`
}

var _ config.Config = &SdsServerConfig{}

func (x *SdsServerConfig) Validate() error {
	if x.GrpcPort < 0 {
		return errors.New("GrpcPort cannot be negative")
	}
	return nil
}
