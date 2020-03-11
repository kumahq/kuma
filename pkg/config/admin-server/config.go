package admin_server

import (
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
)

func DefaultAdminServerConfig() *AdminServerConfig {
	return &AdminServerConfig{
		Apis:   DefaultAdminServerApisConfig(),
		Local:  DefaultLocalAdminServerConfig(),
		Public: DefaultPublicAdminServerConfig(),
	}
}

var _ config.Config = &AdminServerConfig{}

// Admin Server configuration
type AdminServerConfig struct {
	// Admin Server APIs configuration
	Apis *AdminServerApisConfig `yaml:"apis"`
	// Local configuration of server that is available only on localhost
	Local *LocalAdminServerConfig `yaml:"local"`
	// Public configuration of server that is available on public interface
	Public *PublicAdminServerConfig `yaml:"public"`
}

var _ config.Config = &AdminServerApisConfig{}

// Admin Server APIs configuration
type AdminServerApisConfig struct {
	// Dataplane Token API configuration
	DataplaneToken *DataplaneTokenApiConfig `yaml:"dataplaneToken"`
}

func DefaultAdminServerApisConfig() *AdminServerApisConfig {
	return &AdminServerApisConfig{
		DataplaneToken: DefaultDataplaneTokenWsConfig(),
	}
}

func (a *AdminServerApisConfig) Sanitize() {
	a.DataplaneToken.Sanitize()
}

func (a *AdminServerApisConfig) Validate() error {
	if err := a.DataplaneToken.Validate(); err != nil {
		return errors.Wrap(err, "DataplaneToken validation failed")
	}
	return nil
}

var _ config.Config = &DataplaneTokenApiConfig{}

func DefaultDataplaneTokenWsConfig() *DataplaneTokenApiConfig {
	return &DataplaneTokenApiConfig{
		Enabled: true,
	}
}

// Dataplane Token API configuration
type DataplaneTokenApiConfig struct {
	// If true then Dataplane Token WS and token verification is enabled
	Enabled bool `yaml:"enabled" envconfig:"kuma_admin_server_apis_dataplane_token_enabled"`
}

func (d *DataplaneTokenApiConfig) Sanitize() {
}

func (d *DataplaneTokenApiConfig) Validate() error {
	return nil
}

func (i *AdminServerConfig) Sanitize() {
	i.Apis.Sanitize()
	i.Public.Sanitize()
	i.Local.Sanitize()
}

func (i *AdminServerConfig) Validate() error {
	if err := i.Apis.Validate(); err != nil {
		return errors.Wrap(err, "Apis validation failed")
	}
	if err := i.Local.Validate(); err != nil {
		return errors.Wrap(err, "Local validation failed")
	}
	if err := i.Public.Validate(); err != nil {
		return errors.Wrap(err, "Public validation failed")
	}
	return nil
}

// Admin Server configuration of server that is available only on localhost
type LocalAdminServerConfig struct {
	// Port on which the server will be exposed
	Port uint32 `yaml:"port" envconfig:"kuma_admin_server_local_port"`
}

var _ config.Config = &LocalAdminServerConfig{}

func (l *LocalAdminServerConfig) Sanitize() {
}

func (l *LocalAdminServerConfig) Validate() error {
	if l.Port > 65535 {
		return errors.New("Port must be in the range [0, 65535]")
	}
	return nil
}

func DefaultLocalAdminServerConfig() *LocalAdminServerConfig {
	return &LocalAdminServerConfig{
		Port: 5679,
	}
}

// Admin Server configuration of server that is available on public interface
type PublicAdminServerConfig struct {
	// If true then Admin Server is exposed on public interface
	Enabled bool `yaml:"enabled" envconfig:"kuma_admin_server_public_enabled"`
	// Interface on which the server will be exposed
	Interface string `yaml:"interface" envconfig:"kuma_admin_server_public_interface"`
	// Port on which the server will be exposed. If not specified (0) then port from local configuration will be used
	Port uint32 `yaml:"port" envconfig:"kuma_admin_server_public_port"`
	// Path to TLS certificate file
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_admin_server_public_tls_cert_file"`
	// Path to TLS key file
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_admin_server_public_tls_key_file"`
	// Directory of authorized client certificates
	ClientCertsDir string `yaml:"clientCertsDir" envconfig:"kuma_admin_server_public_client_certs_dir"`
}

var _ config.Config = &PublicAdminServerConfig{}

func DefaultPublicAdminServerConfig() *PublicAdminServerConfig {
	return &PublicAdminServerConfig{
		Enabled:        false,
		Interface:      "",
		Port:           0,
		TlsCertFile:    "",
		TlsKeyFile:     "",
		ClientCertsDir: "",
	}
}

func (p *PublicAdminServerConfig) Sanitize() {
}

func (p *PublicAdminServerConfig) Validate() error {
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
