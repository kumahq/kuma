package dp_server

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = &DpServerConfig{}

// Dataplane Server configuration that servers API like Bootstrap/XDS/SDS.
type DpServerConfig struct {
	// Port of the DP Server
	Port int `yaml:"port" envconfig:"kuma_dp_server_port"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert. If empty, autoconfigured from general.tlsCertFile
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_dp_server_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key. If empty, autoconfigured from general.tlsKeyFile
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_dp_server_tls_key_file"`
	// Auth defines an authentication configuration for the DP Server
	Auth DpServerAuthConfig `yaml:"auth"`
}

type DpServerAuthType string

const (
	DpServerAuthServiceAccountToken = "serviceAccountToken"
	DpServerAuthDpToken             = "dpToken"
	DpServerAuthNone                = "none"
)

// Authentication configuration for Dataplane Server
type DpServerAuthConfig struct {
	// Type of authentication. Available values: "serviceAccountToken", "dpToken", "none".
	// If empty, autoconfigured based on the environment - "serviceAccountToken" on Kubernetes, "dpToken" on Universal.
	Type string `yaml:"type" envconfig:"kuma_dp_server_auth_type"`
}

func (a *DpServerAuthConfig) Validate() error {
	if a.Type != "" && a.Type != DpServerAuthNone && a.Type != DpServerAuthDpToken && a.Type != DpServerAuthServiceAccountToken {
		return errors.Errorf("Type is invalid. Available values are: %q, %q, %q", DpServerAuthDpToken, DpServerAuthServiceAccountToken, DpServerAuthNone)
	}
	return nil
}

func (a *DpServerConfig) Sanitize() {
}

func (a *DpServerConfig) Validate() error {
	if a.Port < 0 {
		return errors.New("Port cannot be negative")
	}
	if err := a.Auth.Validate(); err != nil {
		return errors.Wrap(err, "Auth is invalid")
	}
	return nil
}

func DefaultDpServerConfig() *DpServerConfig {
	return &DpServerConfig{
		Port: 5678,
		Auth: DpServerAuthConfig{
			Type: "", // autoconfigured from the environment
		},
	}
}
