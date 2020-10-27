package api_server

import (
	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/api-server/catalog"
)

var _ config.Config = &ApiServerConfig{}

// API Server configuration
type ApiServerConfig struct {
	// If true, then API Server will operate in read only mode (serving GET requests)
	ReadOnly bool `yaml:"readOnly" envconfig:"kuma_api_server_read_only"`
	// API Catalog
	Catalog *catalog.CatalogConfig `yaml:"catalog"`
	// Allowed domains for Cross-Origin Resource Sharing. The value can be either domain or regexp
	CorsAllowedDomains []string             `yaml:"corsAllowedDomains" envconfig:"kuma_api_server_cors_allowed_domains"`
	HTTP               ApiServerHTTPConfig  `yaml:"http"`
	HTTPS              ApiServerHTTPSConfig `yaml:"https"`
}

type ApiServerHTTPConfig struct {
	Enabled   bool   `yaml:"enabled" envconfig:"kuma_api_server_http_enabled"`
	Interface string `yaml:"interface" envconfig:"kuma_api_server_http_interface"`
	// Port of the API Server
	Port uint32 `yaml:"port" envconfig:"kuma_api_server_http_port"`
}

type ApiServerHTTPSConfig struct { // todo comments
	Enabled   bool   `yaml:"enabled" envconfig:"kuma_api_server_https_enabled"`
	Interface string `yaml:"interface" envconfig:"kuma_api_server_https_interface"`
	Port      uint32 `yaml:"port" envconfig:"kuma_admin_server_public_port"`
	// Path to TLS certificate file
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_api_server_https_tls_cert_file"`
	// Path to TLS key file
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_api_server_https_tls_key_file"`
	// Directory of authorized client certificates
	ClientCertsDir string `yaml:"clientCertsDir" envconfig:"kuma_api_server_https_client_certs_dir"`
}

func (a *ApiServerConfig) Sanitize() {
}

func (a *ApiServerConfig) Validate() error {
	return nil
}

func DefaultApiServerConfig() *ApiServerConfig {
	return &ApiServerConfig{
		ReadOnly:           false,
		Catalog:            &catalog.CatalogConfig{},
		CorsAllowedDomains: []string{".*"},
		HTTP: ApiServerHTTPConfig{
			Enabled:   true,
			Interface: "0.0.0.0",
			Port:      5681,
		},
		HTTPS: ApiServerHTTPSConfig{
			Enabled:        true,
			Interface:      "0.0.0.0",
			Port:           5679,
			TlsCertFile:    "", // autoconfigured
			TlsKeyFile:     "", // autoconfigured
			ClientCertsDir: "",
		},
	}
}
