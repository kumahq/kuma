package token_server

import (
	"github.com/Kong/kuma/pkg/config"
	"github.com/pkg/errors"
)

func DefaultDataplaneTokenServerConfig() *DataplaneTokenServerConfig {
	return &DataplaneTokenServerConfig{
		Local:  DefaultLocalDataplaneTokenServerConfig(),
		Public: DefaultPublicDataplaneTokenServerConfig(),
	}
}

var _ config.Config = &DataplaneTokenServerConfig{}

// Dataplane Token Server configuration
type DataplaneTokenServerConfig struct {
	// Local configuration of server that is available only on localhost
	Local *LocalDataplaneTokenServerConfig `yaml:"local"`
	// Public configuration of server that is available on public interface
	Public *PublicDataplaneTokenServerConfig `yaml:"public"`
}

func (i *DataplaneTokenServerConfig) Validate() error {
	if err := i.Local.Validate(); err != nil {
		return errors.Wrap(err, "Local validation failed")
	}
	if err := i.Public.Validate(); err != nil {
		return errors.Wrap(err, "Public validation failed")
	}
	return nil
}

// Dataplane Token Server configuration of server that is available only on localhost
type LocalDataplaneTokenServerConfig struct {
	// Port on which the server will be exposed
	Port uint32 `yaml:"port" envconfig:"kuma_dataplane_token_server_local_port"`
}

var _ config.Config = &LocalDataplaneTokenServerConfig{}

func (l *LocalDataplaneTokenServerConfig) Validate() error {
	if l.Port > 65535 {
		return errors.New("Port must be in the range [0, 65535]")
	}
	return nil
}

func DefaultLocalDataplaneTokenServerConfig() *LocalDataplaneTokenServerConfig {
	return &LocalDataplaneTokenServerConfig{
		Port: 5679,
	}
}

// Dataplane Token Server configuration of server that is available on public interface
type PublicDataplaneTokenServerConfig struct {
	// Interface on which the server will be exposed
	Interface string `yaml:"interface" envconfig:"kuma_dataplane_token_server_public_interface"`
	// Port on which the server will be exposed. If not specified (0) then port from local configuration will be used
	Port uint32 `yaml:"port" envconfig:"kuma_dataplane_token_server_public_port"`
	// Path to TLS certificate file
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_dataplane_token_server_public_tls_cert_file"`
	// Path to TLS key file
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_dataplane_token_server_public_tls_key_file"`
	// Paths to authorized client certificates
	ClientCertFiles []string `yaml:"clientCertFiles" envconfig:"kuma_dataplane_token_server_public_client_cert_files"`
}

var _ config.Config = &PublicDataplaneTokenServerConfig{}

func DefaultPublicDataplaneTokenServerConfig() *PublicDataplaneTokenServerConfig {
	return &PublicDataplaneTokenServerConfig{
		Interface:       "",
		Port:            0,
		TlsCertFile:     "",
		TlsKeyFile:      "",
		ClientCertFiles: nil,
	}
}

func (p *PublicDataplaneTokenServerConfig) Validate() error {
	if p.Port > 65535 {
		return errors.New("Port must be in the range [0, 65535]")
	}
	if p.TlsCertFile == "" && p.TlsKeyFile != "" {
		return errors.New("TlsCertFile cannot be empty if TlsKeyFile has been set")
	}
	if p.TlsKeyFile == "" && p.TlsCertFile != "" {
		return errors.New("TlsKeyFile cannot be empty if TlsCertFile has been set")
	}
	if p.Interface != "" && p.TlsCertFile == "" {
		return errors.New("TlsCertFile and TlsKeyFile have to be set when PublicInterface is specified")
	}
	return nil
}

func (i *DataplaneTokenServerConfig) TlsEnabled() bool {
	return i.Public.Interface != ""
}
