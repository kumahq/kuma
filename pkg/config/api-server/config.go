package api_server

import (
	"net/url"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

var _ config.Config = &ApiServerConfig{}

// ApiServerConfig defines API Server configuration
type ApiServerConfig struct {
	config.BaseConfig

	// If true, then API Server will operate in read only mode (serving GET requests)
	ReadOnly bool `json:"readOnly" envconfig:"kuma_api_server_read_only"`
	// Allowed domains for Cross-Origin Resource Sharing. The value can be either domain or regexp
	CorsAllowedDomains []string `json:"corsAllowedDomains" envconfig:"kuma_api_server_cors_allowed_domains"`
	// HTTP configuration of the API Server
	HTTP ApiServerHTTPConfig `json:"http"`
	// HTTPS configuration of the API Server
	HTTPS ApiServerHTTPSConfig `json:"https"`
	// Authentication configuration for administrative endpoints like Dataplane Token or managing Secrets
	Auth ApiServerAuth `json:"auth"`
	// Authentication configuration for API Server
	Authn ApiServerAuthn `json:"authn"`
	// BasePath the path to serve the API from
	BasePath string `json:"basePath" envconfig:"kuma_api_server_base_path"`
	// RootUrl can be used if you use a reverse proxy
	RootUrl string `json:"rootUrl" envconfig:"kuma_api_server_root_url"`
	// GUI configuration specific to the GUI
	GUI ApiServerGUI `json:"gui,omitempty"`
}

type ApiServerGUI struct {
	// Enabled whether to serve to gui (if mode=zone this has no effect)
	Enabled bool `json:"enabled" envconfig:"kuma_api_server_gui_enabled"`
	// RootUrl can be used if you set a reverse proxy or want to serve the gui from a different path
	RootUrl string `json:"rootUrl" envconfig:"kuma_api_server_gui_root_url"`
	// BasePath the path to serve the GUI from
	BasePath string `json:"basePath" envconfig:"kuma_api_server_gui_base_path"`
}

func (a *ApiServerGUI) Validate() error {
	var errs error
	if a.RootUrl != "" {
		_, err := url.Parse(a.RootUrl)
		if err != nil {
			errs = multierr.Append(errs, errors.New("RootUrl is not a valid url"))
		}
	}
	if a.BasePath != "" {
		_, err := url.Parse(a.BasePath)
		if err != nil {
			errs = multierr.Append(errs, errors.New("BaseGuiPath is not a valid url"))
		}
	}
	return errs
}

// ApiServerHTTPConfig defines API Server HTTP configuration
type ApiServerHTTPConfig struct {
	// If true then API Server will be served on HTTP
	Enabled bool `json:"enabled" envconfig:"kuma_api_server_http_enabled"`
	// Network interface on which HTTP API Server will be exposed
	Interface string `json:"interface" envconfig:"kuma_api_server_http_interface"`
	// Port of the HTTP API Server
	Port uint32 `json:"port" envconfig:"kuma_api_server_http_port"`
}

func (a *ApiServerHTTPConfig) Validate() error {
	var errs error
	if a.Interface == "" {
		errs = multierr.Append(errs, errors.New("Interface cannot be empty"))
	}
	if a.Port > 65535 {
		errs = multierr.Append(errs, errors.New("Port must be in range [0, 65535]"))
	}
	return errs
}

// ApiServerHTTPSConfig defines API Server HTTPS configuration
type ApiServerHTTPSConfig struct {
	// If true then API Server will be served on HTTPS
	Enabled bool `json:"enabled" envconfig:"kuma_api_server_https_enabled"`
	// Network interface on which HTTPS API Server will be exposed
	Interface string `json:"interface" envconfig:"kuma_api_server_https_interface"`
	// Port of the HTTPS API Server
	Port uint32 `json:"port" envconfig:"kuma_api_server_https_port"`
	// Path to TLS certificate file. Autoconfigured from KUMA_GENERAL_TLS_CERT_FILE if empty
	TlsCertFile string `json:"tlsCertFile" envconfig:"kuma_api_server_https_tls_cert_file"`
	// Path to TLS key file. Autoconfigured from KUMA_GENERAL_TLS_KEY_FILE if empty
	TlsKeyFile string `json:"tlsKeyFile" envconfig:"kuma_api_server_https_tls_key_file"`
	// TlsMinVersion defines the minimum TLS version to be used
	TlsMinVersion string `json:"tlsMinVersion" envconfig:"kuma_api_server_https_tls_min_version"`
	// TlsMaxVersion defines the maximum TLS version to be used
	TlsMaxVersion string `json:"tlsMaxVersion" envconfig:"kuma_api_server_https_tls_max_version"`
	// TlsCipherSuites defines the list of ciphers to use
	TlsCipherSuites []string `json:"tlsCipherSuites" envconfig:"kuma_api_server_https_tls_cipher_suites"`
	// If true, then HTTPS connection will require client cert.
	RequireClientCert bool `json:"requireClientCert" envconfig:"kuma_api_server_https_require_client_cert"`
	// Path to the CA certificate which is used to sign client certificates. It is used only for verifying client certificates.
	TlsCaFile string `json:"tlsCaFile" envconfig:"kuma_api_server_https_tls_ca_file"`
}

func (a *ApiServerHTTPSConfig) Validate() error {
	var errs error
	if a.Interface == "" {
		errs = multierr.Append(errs, errors.New(".Interface cannot be empty"))
	}
	if a.Port > 65535 {
		return errors.New("Port must be in range [0, 65535]")
	}
	if (a.TlsKeyFile == "" && a.TlsCertFile != "") || (a.TlsKeyFile != "" && a.TlsCertFile == "") {
		errs = multierr.Append(errs, errors.New("Both TlsCertFile and TlsKeyFile has to be specified"))
	}
	if _, err := config_types.TLSVersion(a.TlsMinVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMinVersion"+err.Error()))
	}
	if _, err := config_types.TLSVersion(a.TlsMaxVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMaxVersion"+err.Error()))
	}
	if _, err := config_types.TLSCiphers(a.TlsCipherSuites); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsCipherSuites"+err.Error()))
	}
	return errs
}

// ApiServerAuth defines API Server authentication configuration
type ApiServerAuth struct {
	// Directory of authorized client certificates (only valid in HTTPS)
	ClientCertsDir string `json:"clientCertsDir" envconfig:"kuma_api_server_auth_client_certs_dir"`
}

// ApiServerAuthn defines Api Server Authentication configuration
type ApiServerAuthn struct {
	// Type of authentication mechanism (available values: "clientCerts", "tokens")
	Type string `json:"type" envconfig:"kuma_api_server_authn_type"`
	// Localhost is authenticated as a user admin of group admin
	LocalhostIsAdmin bool `json:"localhostIsAdmin" envconfig:"kuma_api_server_authn_localhost_is_admin"`
	// Configuration for tokens authentication
	Tokens ApiServerAuthnTokens `json:"tokens"`
}

func (a ApiServerAuthn) Validate() error {
	if a.Type == "tokens" {
		if err := a.Tokens.Validate(); err != nil {
			return errors.Wrap(err, ".Tokens is not valid")
		}
	}
	return nil
}

type ApiServerAuthnTokens struct {
	// If true then User Token with name admin and group admin will be created and placed as admin-user-token Kuma Global Secret
	BootstrapAdminToken bool `json:"bootstrapAdminToken" envconfig:"kuma_api_server_authn_tokens_bootstrap_admin_token"`
	// If true the control plane token issuer is enabled. It's recommended to set it to false when all the tokens are issued offline.
	EnableIssuer bool `json:"enableIssuer" envconfig:"kuma_api_server_authn_tokens_enable_issuer"`
	// Token validator configuration
	Validator TokensValidator `json:"validator"`
}

func (a ApiServerAuthnTokens) Validate() error {
	if err := a.Validator.Validate(); err != nil {
		return errors.Wrap(err, ".Validator is not valid")
	}
	return nil
}

var _ config.Config = TokensValidator{}

type TokensValidator struct {
	config.BaseConfig

	// If true then Kuma secrets with prefix "user-token-signing-key" are considered as signing keys.
	UseSecrets bool `json:"useSecrets" envconfig:"kuma_api_server_authn_tokens_validator_use_secrets"`
	// List of public keys used to validate the token.
	PublicKeys []config_types.PublicKey `json:"publicKeys"`
}

func (t TokensValidator) Validate() error {
	for i, key := range t.PublicKeys {
		if err := key.Validate(); err != nil {
			return errors.Wrapf(err, ".PublicKeys[%d] is not valid", i)
		}
	}
	return nil
}

func (a *ApiServerConfig) Validate() error {
	var errs error
	if err := a.HTTP.Validate(); err != nil {
		errs = multierr.Append(err, errors.Wrap(err, ".HTTP not valid"))
	}
	if err := a.HTTPS.Validate(); err != nil {
		errs = multierr.Append(err, errors.Wrap(err, ".HTTPS not valid"))
	}
	if err := a.GUI.Validate(); err != nil {
		errs = multierr.Append(err, errors.Wrap(err, ".GUI not valid"))
	}
	if a.RootUrl != "" {
		if _, err := url.Parse(a.RootUrl); err != nil {
			errs = multierr.Append(err, errors.New("RootUrl is not a valid URL"))
		}
	}
	if a.BasePath != "" {
		_, err := url.Parse(a.BasePath)
		if err != nil {
			errs = multierr.Append(errs, errors.New("BaseGuiPath is not a valid url"))
		}
	}
	if err := a.Authn.Validate(); err != nil {
		errs = multierr.Append(err, errors.Wrap(err, ".Authn is not valid"))
	}
	return errs
}

func DefaultApiServerConfig() *ApiServerConfig {
	return &ApiServerConfig{
		ReadOnly:           false,
		CorsAllowedDomains: []string{".*"},
		BasePath:           "/",
		HTTP: ApiServerHTTPConfig{
			Enabled:   true,
			Interface: "0.0.0.0",
			Port:      5681,
		},
		HTTPS: ApiServerHTTPSConfig{
			Enabled:         true,
			Interface:       "0.0.0.0",
			Port:            5682,
			TlsCertFile:     "", // autoconfigured
			TlsKeyFile:      "", // autoconfigured
			TlsMinVersion:   "TLSv1_2",
			TlsCipherSuites: []string{},
		},
		Auth: ApiServerAuth{
			ClientCertsDir: "",
		},
		Authn: ApiServerAuthn{
			Type:             "tokens",
			LocalhostIsAdmin: true,
			Tokens: ApiServerAuthnTokens{
				BootstrapAdminToken: true,
				EnableIssuer:        true,
				Validator: TokensValidator{
					UseSecrets: true,
					PublicKeys: []config_types.PublicKey{},
				},
			},
		},
		GUI: ApiServerGUI{
			Enabled:  true,
			BasePath: "/gui",
		},
	}
}
