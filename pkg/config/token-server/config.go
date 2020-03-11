package token_server

import (
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
)

func DefaultDataplaneTokenServerConfig() *DataplaneTokenServerConfig {
	return &DataplaneTokenServerConfig{
		Enabled: true,
		Local:   DefaultLocalDataplaneTokenServerConfig(),
		Public:  DefaultPublicDataplaneTokenServerConfig(),
	}
}

var _ config.Config = &DataplaneTokenServerConfig{}

// Dataplane Token Server configuration
type DataplaneTokenServerConfig struct {
	// If true then Dataplane Token Server and token verification is enabled
	Enabled bool `yaml:"enabled" envconfig:"kuma_dataplane_token_server_enabled"`
	// Local configuration of server that is available only on localhost
	Local *LocalDataplaneTokenServerConfig `yaml:"local"`
	// Public configuration of server that is available on public interface
	Public *PublicDataplaneTokenServerConfig `yaml:"public"`
}

func (i *DataplaneTokenServerConfig) Sanitize() {
	i.Public.Sanitize()
	i.Local.Sanitize()
}

func (i *DataplaneTokenServerConfig) Validate() error {
	if err := i.Local.Validate(); err != nil {
		return errors.Wrap(err, "Local validation failed")
	}
	if err := i.Public.Validate(); err != nil {
		return errors.Wrap(err, "Public validation failed")
	}
	if !i.Enabled && i.Public.Enabled {
		return errors.New("Public.Enabled cannot be true when server is disabled.")
	}
	return nil
}

// Dataplane Token Server configuration of server that is available only on localhost
type LocalDataplaneTokenServerConfig struct {
	// Port on which the server will be exposed
	Port uint32 `yaml:"port" envconfig:"kuma_dataplane_token_server_local_port"`
}

var _ config.Config = &LocalDataplaneTokenServerConfig{}

func (l *LocalDataplaneTokenServerConfig) Sanitize() {
}

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
	// If true then Dataplane Token Server is exposed on public interface
	Enabled bool `yaml:"enabled" envconfig:"kuma_dataplane_token_server_public_enabled"`
	// Interface on which the server will be exposed
	Interface string `yaml:"interface" envconfig:"kuma_dataplane_token_server_public_interface"`
	// Port on which the server will be exposed. If not specified (0) then port from local configuration will be used
	Port uint32 `yaml:"port" envconfig:"kuma_dataplane_token_server_public_port"`
	// Path to TLS certificate file
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_dataplane_token_server_public_tls_cert_file"`
	// Path to TLS key file
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_dataplane_token_server_public_tls_key_file"`
	// Directory of authorized client certificates
	ClientCertsDir string `yaml:"clientCertsDir" envconfig:"kuma_dataplane_token_server_public_client_certs_dir"`
}

var _ config.Config = &PublicDataplaneTokenServerConfig{}

func DefaultPublicDataplaneTokenServerConfig() *PublicDataplaneTokenServerConfig {
	return &PublicDataplaneTokenServerConfig{
		Enabled:        false,
		Interface:      "",
		Port:           0,
		TlsCertFile:    "",
		TlsKeyFile:     "",
		ClientCertsDir: "",
	}
}

func (p *PublicDataplaneTokenServerConfig) Sanitize() {
}

func (p *PublicDataplaneTokenServerConfig) Validate() error {
	if p.Port > 65535 {
		return errors.New("Port must be in the range [0, 65535]")
	}
	if p.Enabled {
		if p.TlsCertFile == "" {
			return errors.New("TlsCertFile cannot be empty if server is enabled")
		}
		if p.TlsKeyFile == "" {
			return errors.New("TlsKeyFile cannot be empty if server is enabled")
		}
		if p.Interface == "" {
			return errors.New("Interface cannot be empty if server is enabled")
		}
	}
	return nil
}
