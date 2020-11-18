package sds

import (
	"time"

	"github.com/kumahq/kuma/pkg/config"
)

func DefaultSdsServerConfig() *SdsServerConfig {
	return &SdsServerConfig{
		DataplaneConfigurationRefreshInterval: 1 * time.Second,
	}
}

// Envoy SDS server configuration
type SdsServerConfig struct {
	// Interval for re-genarting configuration for Dataplanes connected to the Control Plane
	DataplaneConfigurationRefreshInterval time.Duration `yaml:"dataplaneConfigurationRefreshInterval" envconfig:"kuma_sds_server_dataplane_configuration_refresh_interval"`
}

var _ config.Config = &SdsServerConfig{}

func (c *SdsServerConfig) Sanitize() {
}

func (c *SdsServerConfig) Validate() error {
	return nil
}
