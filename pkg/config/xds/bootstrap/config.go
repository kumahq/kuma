package bootstrap

import (
	"net"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ config.Config = &BootstrapServerConfig{}

type BootstrapServerConfig struct {
	// The version of Envoy API (available: "v2", "v3")
	APIVersion envoy_common.APIVersion `yaml:"apiVersion" envconfig:"kuma_bootstrap_server_api_version"`
	// Parameters of bootstrap configuration
	Params *BootstrapParamsConfig `yaml:"params"`
}

func (b *BootstrapServerConfig) Sanitize() {
	b.Params.Sanitize()
}

func (b *BootstrapServerConfig) Validate() error {
	if err := b.Params.Validate(); err != nil {
		return errors.Wrap(err, "Params validation failed")
	}
	switch b.APIVersion {
	case envoy_common.APIV3:
	default:
		return errors.Errorf("APIVersion has invalid value. Available values: %q", envoy_common.APIV3)
	}
	return nil
}

func DefaultBootstrapServerConfig() *BootstrapServerConfig {
	return &BootstrapServerConfig{
		APIVersion: envoy_common.APIV3,
		Params:     DefaultBootstrapParamsConfig(),
	}
}

var _ config.Config = &BootstrapParamsConfig{}

type BootstrapParamsConfig struct {
	// Address of Envoy Admin
	AdminAddress string `yaml:"adminAddress" envconfig:"kuma_bootstrap_server_params_admin_address"`
	// Port of Envoy Admin
	AdminPort uint32 `yaml:"adminPort" envconfig:"kuma_bootstrap_server_params_admin_port"`
	// Path to access log file of Envoy Admin
	AdminAccessLogPath string `yaml:"adminAccessLogPath" envconfig:"kuma_bootstrap_server_params_admin_access_log_path"`
	// Host of XDS Server. By default it is the same host as the one used by kuma-dp to connect to the control plane
	XdsHost string `yaml:"xdsHost" envconfig:"kuma_bootstrap_server_params_xds_host"`
	// Port of XDS Server. By default it is autoconfigured from KUMA_XDS_SERVER_GRPC_PORT
	XdsPort uint32 `yaml:"xdsPort" envconfig:"kuma_bootstrap_server_params_xds_port"`
	// Connection timeout to the XDS Server
	XdsConnectTimeout time.Duration `yaml:"xdsConnectTimeout" envconfig:"kuma_bootstrap_server_params_xds_connect_timeout"`
}

func (b *BootstrapParamsConfig) Sanitize() {
}

func (b *BootstrapParamsConfig) Validate() error {
	if b.AdminAddress == "" {
		return errors.New("AdminAddress cannot be empty")
	}
	if net.ParseIP(b.AdminAddress) == nil {
		return errors.New("AdminAddress should be a valid IP address")
	}
	if b.AdminPort > 65535 {
		return errors.New("AdminPort must be in the range [0, 65535]")
	}
	if b.AdminAccessLogPath == "" {
		return errors.New("AdminAccessLogPath cannot be empty")
	}
	if b.XdsPort > 65535 {
		return errors.New("AdminPort must be in the range [0, 65535]")
	}
	if b.XdsConnectTimeout < 0 {
		return errors.New("XdsConnectTimeout cannot be negative")
	}
	return nil
}

func DefaultBootstrapParamsConfig() *BootstrapParamsConfig {
	return &BootstrapParamsConfig{
		AdminAddress:       "127.0.0.1", // by default, Envoy Admin interface should listen on loopback address
		AdminPort:          0,           // by default, turn off Admin interface of Envoy
		AdminAccessLogPath: "/dev/null",
		XdsHost:            "", // by default it is the same host as the one used by kuma-dp to connect to the control plane
		XdsPort:            0,  // by default it is autoconfigured from KUMA_XDS_SERVER_GRPC_PORT
		XdsConnectTimeout:  1 * time.Second,
	}
}
