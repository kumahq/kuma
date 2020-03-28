package bootstrap

import (
	"net"
	"time"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
)

var _ config.Config = &BootstrapServerConfig{}

type BootstrapServerConfig struct {
	// Port of Server that provides bootstrap configuration for dataplanes
	Port uint32 `yaml:"port" envconfig:"kuma_bootstrap_server_port"`
	// Parameters of bootstrap configuration
	Params *BootstrapParamsConfig `yaml:"params"`
}

func (b *BootstrapServerConfig) Sanitize() {
	b.Params.Sanitize()
}

func (b *BootstrapServerConfig) Validate() error {
	if b.Port > 65535 {
		return errors.New("Port must be in the range [0, 65535]")
	}
	if err := b.Params.Validate(); err != nil {
		return errors.Wrap(err, "Params validation failed")
	}
	return nil
}

func DefaultBootstrapServerConfig() *BootstrapServerConfig {
	return &BootstrapServerConfig{
		Port:   5682,
		Params: DefaultBootstrapParamsConfig(),
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
	// Host of XDS Server. By default it is autoconfigured from KUMA_GENERAL_ADVERTISED_HOSTNAME
	XdsHost string `yaml:"xdsHost" envconfig:"kuma_bootstrap_server_params_xds_host"`
	// Port of XDS Server. By default it is autoconfigured from KUMA_XDS_SERVER_GRPC_PORT
	XdsPort uint32 `yaml:"xdsPort" envconfig:"kuma_bootstrap_server_params_xds_port"`
	// Connection timeout to the XDS Server
	XdsConnectTimeout time.Duration `yaml:"xdsConnectTimeout" envconfig:"kuma_bootstrap_server_params_xds_connect_timeout"`
	// PEM-encoded TLS cert used by envoy
	XdsClientTlsCertFile string `yaml:"xdsClientTlsCertFile" envconfig:"kuma_bootstrap_server_params_xds_client_tls_cert_file"`
	// PEM-encoded TLS key used by envoy
	XdsClientTlsKeyFile string `yaml:"xdsClientTlsKeyFile" envconfig:"kuma_bootstrap_server_params_xds_client_tls_key_file"`
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
	if b.XdsClientTlsCertFile != "" && b.XdsClientTlsKeyFile == "" {
		return errors.New("xdsClientTlsKeyFile cannot be empty if xdsClientTlsCertFile is set")
	}
	if b.XdsClientTlsCertFile == "" && b.XdsClientTlsKeyFile != "" {
		return errors.New("xdsClientTlsCertFile cannot be empty if xdsClientTlsKeyFile is set")
	}
	return nil
}

func DefaultBootstrapParamsConfig() *BootstrapParamsConfig {
	return &BootstrapParamsConfig{
		AdminAddress:         "127.0.0.1", // by default, Envoy Admin interface should listen on loopback address
		AdminPort:            0,           // by default, turn off Admin interface of Envoy
		AdminAccessLogPath:   "/dev/null",
		XdsHost:              "", // by default it is autoconfigured from KUMA_GENERAL_ADVERTISED_HOSTNAME
		XdsPort:              0,  // by default it is autoconfigured from KUMA_XDS_SERVER_GRPC_PORT
		XdsConnectTimeout:    1 * time.Second,
		XdsClientTlsCertFile: "",
		XdsClientTlsKeyFile:  "",
	}
}
