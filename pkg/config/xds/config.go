package xds

import (
	"errors"
	"time"

	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = &XdsServerConfig{}

// Envoy XDS server configuration
type XdsServerConfig struct {
	// Interval for re-genarting configuration for Dataplanes connected to the Control Plane
	DataplaneConfigurationRefreshInterval time.Duration `yaml:"dataplaneConfigurationRefreshInterval" envconfig:"kuma_xds_server_dataplane_configuration_refresh_interval"`
	// Interval for flushing status of Dataplanes connected to the Control Plane
	DataplaneStatusFlushInterval time.Duration `yaml:"dataplaneStatusFlushInterval" envconfig:"kuma_xds_server_dataplane_status_flush_interval"`
	// Backoff that is executed when Control Plane is sending the response that was previously rejected by Dataplane
	NACKBackoff time.Duration `yaml:"nackBackoff" envconfig:"kuma_xds_server_nack_backoff"`
}

func (x *XdsServerConfig) Sanitize() {
}

func (x *XdsServerConfig) Validate() error {
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
		DataplaneConfigurationRefreshInterval: 1 * time.Second,
		DataplaneStatusFlushInterval:          10 * time.Second,
		NACKBackoff:                           5 * time.Second,
	}
}
