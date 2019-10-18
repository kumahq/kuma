package token_server

import (
	"github.com/Kong/kuma/pkg/config"
	"github.com/pkg/errors"
)

func DefaultDataplaneTokenServerConfig() *DataplaneTokenServerConfig {
	return &DataplaneTokenServerConfig{
		Port:            5679,
		PublicInterface: "",
		PublicPort:      0,
		TlsCertFile:     "",
		TlsKeyFile:      "",
		ClientCertFiles: nil,
	}
}

// Dataplane Token Server configuration
type DataplaneTokenServerConfig struct {
	// Port of the server
	Port uint32 `yaml:"port" envconfig:"kuma_dataplane_token_server_port"`
	// Public interface on which the SSL server will be exposed
	PublicInterface string `yaml:"publicInterface" envconfig:"kuma_dataplane_token_server_public_interface"`
	// Public port. If not specified (0) then Port will be used
	PublicPort uint32 `yaml:"publicPort" envconfig:"kuma_dataplane_token_server_public_port"`
	// Path to TLS certificate file
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_dataplane_token_server_tls_cert_file"`
	// Path to TLS key file
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_dataplane_token_server_tls_key_file"`
	// Paths to authorized client certificates
	ClientCertFiles []string `yaml:"clientCertFiles" envconfig:"kuma_dataplane_token_server_client_cert_files"`
}

var _ config.Config = &DataplaneTokenServerConfig{}

func (i *DataplaneTokenServerConfig) Validate() error {
	if i.Port > 65535 {
		return errors.New("Port must be in the range [0, 65535]")
	}
	if i.PublicPort > 65535 {
		return errors.New("PublicPort must be in the range [0, 65535]")
	}
	if i.TlsCertFile == "" && i.TlsKeyFile != "" {
		return errors.New("TlsCertFile cannot be empty if TlsKeyFile has been set")
	}
	if i.TlsKeyFile == "" && i.TlsCertFile != "" {
		return errors.New("TlsKeyFile cannot be empty if TlsCertFile has been set")
	}
	if i.PublicInterface != "" && i.TlsCertFile == "" {
		return errors.New("TlsCertFile and TlsKeyFile have to be set when PublicInterface is specified")
	}
	return nil
}

func (i *DataplaneTokenServerConfig) TlsEnabled() bool {
	return i.PublicInterface != ""
}
