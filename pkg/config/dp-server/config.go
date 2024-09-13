package dp_server

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

var _ config.Config = &DpServerConfig{}

// DpServerConfig defines the data plane Server configuration that serves API
// like Bootstrap/XDS.
type DpServerConfig struct {
	config.BaseConfig

	// Port of the DP Server
	Port int `json:"port" envconfig:"kuma_dp_server_port"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert. If empty, autoconfigured from general.tlsCertFile
	TlsCertFile string `json:"tlsCertFile" envconfig:"kuma_dp_server_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key. If empty, autoconfigured from general.tlsKeyFile
	TlsKeyFile string `json:"tlsKeyFile" envconfig:"kuma_dp_server_tls_key_file"`
	// TlsMinVersion defines the minimum TLS version to be used
	TlsMinVersion string `json:"tlsMinVersion" envconfig:"kuma_dp_server_tls_min_version"`
	// TlsMaxVersion defines the maximum TLS version to be used
	TlsMaxVersion string `json:"tlsMaxVersion" envconfig:"kuma_dp_server_tls_max_version"`
	// TlsCipherSuites defines the list of ciphers to use
	TlsCipherSuites []string `json:"tlsCipherSuites" envconfig:"kuma_dp_server_tls_cipher_suites"`
	// ReadHeaderTimeout defines the amount of time DP server will be
	// allowed to read request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what is considered
	// too slow for the body. If ReadHeaderTimeout is zero there is no timeout.
	//
	// The timeout is configurable as in rare cases, when Kuma CP was restarting,
	// 1s which is explicitly set in other servers was insufficient and DPs
	// were failing to reconnect (we observed this in Projected Service Account
	// Tokens e2e tests, which started flaking a lot after introducing explicit
	// 1s timeout)
	ReadHeaderTimeout config_types.Duration `json:"readHeaderTimeout" envconfig:"kuma_dp_server_read_header_timeout"`
	// Auth defines an authentication configuration for the DP Server
	// Deprecated: use "authn" section.
	Auth DpServerAuthConfig `json:"auth"`
	// Authn defines authentication configuration for the DP Server.
	Authn DpServerAuthnConfig `json:"authn"`
	// Hds defines a Health Discovery Service configuration
	Hds *HdsConfig `json:"hds"`
}

type DpServerAuthType string

const (
	DpServerAuthServiceAccountToken = "serviceAccountToken"
	DpServerAuthDpToken             = "dpToken"
	DpServerAuthZoneToken           = "zoneToken"
	DpServerAuthNone                = "none"
)

// Authentication configuration for Dataplane Server
type DpServerAuthConfig struct {
	config.BaseConfig

	// Type of authentication. Available values: "serviceAccountToken", "dpToken", "none".
	// If empty, autoconfigured based on the environment - "serviceAccountToken" on Kubernetes, "dpToken" on Universal.
	Type string `json:"type" envconfig:"kuma_dp_server_auth_type"`
	// UseTokenPath define if should use config for ads with path to token that can be reloaded.
	UseTokenPath bool `json:"useTokenPath" envconfig:"kuma_dp_server_auth_use_token_path"`
}

func (a *DpServerAuthConfig) Validate() error {
	if a.Type != "" && a.Type != DpServerAuthNone && a.Type != DpServerAuthDpToken && a.Type != DpServerAuthServiceAccountToken {
		return errors.Errorf("Type is invalid. Available values are: %q, %q, %q", DpServerAuthDpToken, DpServerAuthServiceAccountToken, DpServerAuthNone)
	}
	return nil
}

func (a *DpServerConfig) PostProcess() error {
	return multierr.Combine(a.Hds.PostProcess())
}

func (a *DpServerConfig) Validate() error {
	var errs error
	if a.Port < 0 {
		errs = multierr.Append(errs, errors.New(".Port cannot be negative"))
	}
	if err := a.Auth.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrap(err, ".Auth is invalid"))
	}
	if err := a.Authn.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrap(err, ".Auth is invalid"))
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
	if a.ReadHeaderTimeout.Duration < 0 {
		return errors.New("ReadHeaderTimeout must be greater or equal 0s")
	}
	return errs
}

type DpServerAuthnConfig struct {
	// Configuration for data plane proxy authentication.
	DpProxy DpProxyAuthnConfig `json:"dpProxy"`
	// Configuration for zone proxy authentication.
	ZoneProxy ZoneProxyAuthnConfig `json:"zoneProxy"`
	// If true then Envoy uses Google gRPC instead of Envoy gRPC which lets a proxy reload the auth data (service account token, dp token etc.) from path without proxy restart.
	// This is enabled on Kubernetes.
	EnableReloadableTokens bool `json:"enableReloadableTokens" envconfig:"kuma_dp_server_authn_enable_reloadable_tokens"`
}

func (d DpServerAuthnConfig) Validate() error {
	if err := d.DpProxy.Validate(); err != nil {
		return errors.Wrap(err, ".DpProxy is not valid")
	}
	if err := d.ZoneProxy.Validate(); err != nil {
		return errors.Wrap(err, ".ZoneProxy is not valid")
	}
	return nil
}

type ZoneTokenAuthnConfig struct {
	// If true the control plane token issuer is enabled. It's recommended to set it to false when all the tokens are issued offline.
	EnableIssuer bool `json:"enableIssuer" envconfig:"kuma_dp_server_authn_zone_proxy_zone_token_enable_issuer"`
	// Zone Token validator configuration
	Validator ZoneTokenValidatorConfig `json:"validator"`
}

func (c ZoneTokenAuthnConfig) Validate() error {
	if err := c.Validator.Validate(); err != nil {
		return errors.Wrap(err, ".Validator is not valida")
	}
	return nil
}

type ZoneProxyAuthnConfig struct {
	// Type of authentication. Available values: "serviceAccountToken", "zoneToken", "none".
	// If empty, autoconfigured based on the environment - "serviceAccountToken" on Kubernetes, "zoneToken" on Universal.
	Type string `json:"type" envconfig:"kuma_dp_server_authn_zone_proxy_type"`
	// Configuration for zoneToken authentication method.
	ZoneToken ZoneTokenAuthnConfig `json:"zoneToken"`
}

func (c ZoneProxyAuthnConfig) Validate() error {
	if c.Type == DpServerAuthZoneToken {
		if err := c.ZoneToken.Validate(); err != nil {
			return errors.Wrap(err, ".ZoneToken is not valid")
		}
	}
	return nil
}

type ZoneTokenValidatorConfig struct {
	// If true then Kuma secrets with prefix "zone-token-signing-key" are considered as signing keys.
	UseSecrets bool `json:"useSecrets" envconfig:"kuma_dp_server_authn_zone_proxy_zone_token_validator_use_secrets"`
	// List of public keys used to validate the token
	PublicKeys []config_types.PublicKey `json:"publicKeys"`
}

func (z ZoneTokenValidatorConfig) Validate() error {
	for i, key := range z.PublicKeys {
		if err := key.Validate(); err != nil {
			return errors.Wrapf(err, ".PublicKeys[%d] is not valid", i)
		}
	}
	return nil
}

type DpProxyAuthnConfig struct {
	// Type of authentication. Available values: "serviceAccountToken", "dpToken", "none".
	// If empty, autoconfigured based on the environment - "serviceAccountToken" on Kubernetes, "dpToken" on Universal.
	Type string `json:"type" envconfig:"kuma_dp_server_authn_dp_proxy_type"`
	// Configuration of dpToken authentication method
	DpToken DpTokenAuthnConfig `json:"dpToken"`
}

func (d DpProxyAuthnConfig) Validate() error {
	if d.Type == DpServerAuthDpToken {
		if err := d.DpToken.Validate(); err != nil {
			return errors.Wrap(err, ".DpToken is not valid")
		}
	}
	return nil
}

type DpTokenAuthnConfig struct {
	// If true the control plane token issuer is enabled. It's recommended to set it to false when all the tokens are issued offline.
	EnableIssuer bool `json:"enableIssuer" envconfig:"kuma_dp_server_authn_dp_proxy_dp_token_enable_issuer"`
	// DP Token validator configuration
	Validator DpTokenValidatorConfig `json:"validator"`
}

func (d DpTokenAuthnConfig) Validate() error {
	if err := d.Validator.Validate(); err != nil {
		return errors.Wrap(err, ".Validator is not valid")
	}
	return nil
}

type DpTokenValidatorConfig struct {
	// If true then Kuma secrets with prefix "dataplane-token-signing-key-{mesh}" are considered as signing keys.
	UseSecrets bool `json:"useSecrets" envconfig:"kuma_dp_server_authn_dp_proxy_dp_token_validator_use_secrets"`
	// List of public keys used to validate the token
	PublicKeys []config_types.MeshedPublicKey `json:"publicKeys"`
}

func (d DpTokenValidatorConfig) Validate() error {
	for i, key := range d.PublicKeys {
		if err := key.Validate(); err != nil {
			return errors.Wrapf(err, ".PublicKeys[%d] is not valid", i)
		}
	}
	return nil
}

func DefaultDpServerConfig() *DpServerConfig {
	return &DpServerConfig{
		Port: 5678,
		Auth: DpServerAuthConfig{
			Type:         "", // autoconfigured from the environment
			UseTokenPath: false,
		},
		Authn: DpServerAuthnConfig{
			DpProxy: DpProxyAuthnConfig{
				Type: "", // autoconfigured from the environment
				DpToken: DpTokenAuthnConfig{
					EnableIssuer: true,
					Validator: DpTokenValidatorConfig{
						UseSecrets: true,
						PublicKeys: []config_types.MeshedPublicKey{},
					},
				},
			},
			ZoneProxy: ZoneProxyAuthnConfig{
				Type: "", // autoconfigured from the environment
				ZoneToken: ZoneTokenAuthnConfig{
					EnableIssuer: true,
					Validator: ZoneTokenValidatorConfig{
						UseSecrets: true,
						PublicKeys: []config_types.PublicKey{},
					},
				},
			},
		},
		Hds:             DefaultHdsConfig(),
		TlsMinVersion:   "TLSv1_2",
		TlsCipherSuites: []string{},
		// Pay attention that the default value is set to 5s and not 1s
		// like in the other parts of the code. In rare cases, when Kuma CP
		// was restarting, 1s was insufficient and DPs were failing to reconnect
		// (we observed this in Projected Service Account Tokens e2e tests,
		// which started flaking a lot after introducing this 1s timeout)
		ReadHeaderTimeout: config_types.Duration{Duration: 5 * time.Second},
	}
}

func DefaultHdsConfig() *HdsConfig {
	return &HdsConfig{
		Enabled:         true,
		Interval:        config_types.Duration{Duration: 5 * time.Second},
		RefreshInterval: config_types.Duration{Duration: 10 * time.Second},
		CheckDefaults: &HdsCheck{
			Timeout:            config_types.Duration{Duration: 2 * time.Second},
			Interval:           config_types.Duration{Duration: 1 * time.Second},
			NoTrafficInterval:  config_types.Duration{Duration: 1 * time.Second},
			HealthyThreshold:   1,
			UnhealthyThreshold: 1,
		},
	}
}

type HdsConfig struct {
	config.BaseConfig

	// Enabled if true then Envoy will actively check application's ports, but only on Universal.
	// On Kubernetes this feature disabled for now regardless the flag value
	Enabled bool `json:"enabled" envconfig:"kuma_dp_server_hds_enabled"`
	// Interval for Envoy to send statuses for HealthChecks
	Interval config_types.Duration `json:"interval" envconfig:"kuma_dp_server_hds_interval"`
	// RefreshInterval is an interval for re-genarting configuration for Dataplanes connected to the Control Plane
	RefreshInterval config_types.Duration `json:"refreshInterval" envconfig:"kuma_dp_server_hds_refresh_interval"`
	// CheckDefaults defines a HealthCheck configuration
	CheckDefaults *HdsCheck `json:"checkDefaults"`
}

func (h *HdsConfig) PostProcess() error {
	return multierr.Combine(h.CheckDefaults.PostProcess())
}

func (h *HdsConfig) Validate() error {
	if h.Interval.Duration <= 0 {
		return errors.New("Interval must be greater than 0s")
	}
	if err := h.CheckDefaults.Validate(); err != nil {
		return errors.Wrap(err, "Check is invalid")
	}
	return nil
}

var _ config.Config = &HdsCheck{}

type HdsCheck struct {
	config.BaseConfig

	// Timeout is a time to wait for a health check response. If the timeout is reached the
	// health check attempt will be considered a failure.
	Timeout config_types.Duration `json:"timeout" envconfig:"kuma_dp_server_hds_check_timeout"`
	// Interval between health checks.
	Interval config_types.Duration `json:"interval" envconfig:"kuma_dp_server_hds_check_interval"`
	// NoTrafficInterval is a special health check interval that is used when a cluster has
	// never had traffic routed to it.
	NoTrafficInterval config_types.Duration `json:"noTrafficInterval" envconfig:"kuma_dp_server_hds_check_no_traffic_interval"`
	// HealthyThreshold is a number of healthy health checks required before a host is marked
	// healthy.
	HealthyThreshold uint32 `json:"healthyThreshold" envconfig:"kuma_dp_server_hds_check_healthy_threshold"`
	// UnhealthyThreshold is a number of unhealthy health checks required before a host is marked
	// unhealthy.
	UnhealthyThreshold uint32 `json:"unhealthyThreshold" envconfig:"kuma_dp_server_hds_check_unhealthy_threshold"`
}

func (h *HdsCheck) Validate() error {
	if h.Timeout.Duration <= 0 {
		return errors.New("Timeout must be greater than 0s")
	}
	if h.Interval.Duration <= 0 {
		return errors.New("Interval must be greater than 0s")
	}
	if h.NoTrafficInterval.Duration <= 0 {
		return errors.New("NoTrafficInterval must be greater than 0s")
	}
	return nil
}
