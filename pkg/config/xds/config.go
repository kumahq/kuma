package xds

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = &XdsServerConfig{}

// Envoy XDS server configuration
type XdsServerConfig struct {
	// Interval for re-genarting configuration for Dataplanes connected to the Control Plane
	DataplaneConfigurationRefreshInterval time.Duration `yaml:"dataplaneConfigurationRefreshInterval" envconfig:"kuma_xds_server_dataplane_configuration_refresh_interval"`
	// Interval for flushing status of Dataplanes connected to the Control Plane
	DataplaneStatusFlushInterval time.Duration `yaml:"dataplaneStatusFlushInterval" envconfig:"kuma_xds_server_dataplane_status_flush_interval"`
	// DataplaneDeregistrationDelay is a delay between proxy terminating a connection and the CP trying to deregister the proxy.
	// It is used only in universal mode when you use direct lifecycle.
	// Setting this setting to 0s disables the delay.
	// Disabling this may cause race conditions that one instance of CP removes proxy object
	// while proxy is connected to another instance of the CP.
	DataplaneDeregistrationDelay time.Duration `yaml:"dataplaneDeregistrationDelay" envconfig:"kuma_xds_dataplane_deregistration_delay"`
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
		DataplaneDeregistrationDelay:          10 * time.Second,
		NACKBackoff:                           5 * time.Second,
	}
}

type Proxy struct {
	// Gateway holds data plane wide configuration for MeshGateway proxies
	Gateway Gateway `yaml:"gateway"`
}

type Gateway struct {
	GlobalDownstreamMaxConnections uint64 `yaml:"globalDownstreamMaxConnections" envconfig:"kuma_proxy_gateway_global_downstream_max_connections"`
}

func DefaultProxyConfig() Proxy {
	return Proxy{
		Gateway: Gateway{
			GlobalDownstreamMaxConnections: 50000,
		},
	}
}
