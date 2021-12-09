package api_server

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = &ApiServerConfig{}

// API Server configuration
type ApiServerConfig struct {
	// If true, then API Server will operate in read only mode (serving GET requests)
	ReadOnly bool `yaml:"readOnly" envconfig:"kuma_api_server_read_only"`
	// Allowed domains for Cross-Origin Resource Sharing. The value can be either domain or regexp
	CorsAllowedDomains []string `yaml:"corsAllowedDomains" envconfig:"kuma_api_server_cors_allowed_domains"`
	// HTTP configuration of the API Server
	HTTP ApiServerHTTPConfig `yaml:"http"`
	// HTTPS configuration of the API Server
	HTTPS ApiServerHTTPSConfig `yaml:"https"`
	// Authentication configuration for administrative endpoints like Dataplane Token or managing Secrets
	Auth ApiServerAuth `yaml:"auth"`
	// Authentication configuration for API Server
	Authn ApiServerAuthn `yaml:"authn"`
}

// API Server HTTP configuration
type ApiServerHTTPConfig struct {
	// If true then API Server will be served on HTTP
	Enabled bool `yaml:"enabled" envconfig:"kuma_api_server_http_enabled"`
	// Network interface on which HTTP API Server will be exposed
	Interface string `yaml:"interface" envconfig:"kuma_api_server_http_interface"`
	// Port of the HTTP API Server
	Port uint32 `yaml:"port" envconfig:"kuma_api_server_http_port"`
}

func (a *ApiServerHTTPConfig) Validate() error {
	if a.Interface == "" {
		return errors.New("Interface cannot be empty")
	}
	if a.Port > 65535 {
		return errors.New("Port must be in range [0, 65535]")
	}
	return nil
}

// API Server HTTPS configuration
type ApiServerHTTPSConfig struct {
	// If true then API Server will be served on HTTPS
	Enabled bool `yaml:"enabled" envconfig:"kuma_api_server_https_enabled"`
	// Network interface on which HTTPS API Server will be exposed
	Interface string `yaml:"interface" envconfig:"kuma_api_server_https_interface"`
	// Port of the HTTPS API Server
	Port uint32 `yaml:"port" envconfig:"kuma_api_server_https_port"`
	// Path to TLS certificate file. Autoconfigured from KUMA_GENERAL_TLS_CERT_FILE if empty
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_api_server_https_tls_cert_file"`
	// Path to TLS key file. Autoconfigured from KUMA_GENERAL_TLS_KEY_FILE if empty
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_api_server_https_tls_key_file"`
}

func (a *ApiServerHTTPSConfig) Validate() error {
	if a.Interface == "" {
		return errors.New("Interface cannot be empty")
	}
	if a.Port > 65535 {
		return errors.New("Port must be in range [0, 65535]")
	}
	if (a.TlsKeyFile == "" && a.TlsCertFile != "") || (a.TlsKeyFile != "" && a.TlsCertFile == "") {
		return errors.New("Both TlsCertFile and TlsKeyFile has to be specified")
	}
	return nil
}

// API Server authentication configuration
type ApiServerAuth struct {
	// Directory of authorized client certificates (only validate in HTTPS)
	ClientCertsDir string `yaml:"clientCertsDir" envconfig:"kuma_api_server_auth_client_certs_dir"`
}

// Api Server Authentication configuration
type ApiServerAuthn struct {
	// Type of authentication mechanism (available values: "clientCerts")
	Type string `yaml:"type" envconfig:"kuma_api_server_authn_type"`
	// Localhost is authenticated as a user admin of group admin
	LocalhostIsAdmin bool `yaml:"localhostIsAdmin" envconfig:"kuma_api_server_authn_localhost_is_admin"`
	// Configuration for tokens authentication
	Tokens ApiServerAuthnTokens `yaml:"tokens"`
}

type ApiServerAuthnTokens struct {
	// If true then User Token with name admin and group admin will be created and placed as admin-user-token Kuma Global Secret
	BootstrapAdminToken bool `yaml:"bootstrapAdminToken" envconfig:"kuma_api_server_authn_tokens_bootstrap_admin_token"`
}

func (a *ApiServerConfig) Sanitize() {
}

func (a *ApiServerConfig) Validate() error {
	if err := a.HTTP.Validate(); err != nil {
		return errors.Wrap(err, ".HTTP not valid")
	}
	if err := a.HTTPS.Validate(); err != nil {
		return errors.Wrap(err, ".HTTP not valid")
	}
	return nil
}

func DefaultApiServerConfig() *ApiServerConfig {
	return &ApiServerConfig{
		ReadOnly:           false,
		CorsAllowedDomains: []string{".*"},
		HTTP: ApiServerHTTPConfig{
			Enabled:   true,
			Interface: "0.0.0.0",
			Port:      5681,
		},
		HTTPS: ApiServerHTTPSConfig{
			Enabled:     true,
			Interface:   "0.0.0.0",
			Port:        5682,
			TlsCertFile: "", // autoconfigured
			TlsKeyFile:  "", // autoconfigured
		},
		Auth: ApiServerAuth{
			ClientCertsDir: "",
		},
		Authn: ApiServerAuthn{
			Type:             "tokens",
			LocalhostIsAdmin: true,
			Tokens: ApiServerAuthnTokens{
				BootstrapAdminToken: true,
			},
		},
	}
}
