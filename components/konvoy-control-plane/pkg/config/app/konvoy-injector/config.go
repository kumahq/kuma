package konvoyinjector

import (
	"net"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func DefaultConfig() Config {
	return Config{
		WebHookServer: WebHookServer{
			Address: "", // listen on all addresses
			Port:    8443,
		},
	}
}

// Config defines cofiguration of the Konvoy Injector.
type Config struct {
	// WebHookServer defines cofiguration of an https server that implements Kubernetes Admission WebHook.
	WebHookServer WebHookServer `yaml:"webHookServer,omitempty"`
}

// WebHookServer defines cofiguration of an https server that implements Kubernetes Admission WebHook.
type WebHookServer struct {
	// Address defines the address the https server should be listening on.
	// By default, https server will be listening on all addresses.
	Address string `yaml:"address,omitempty" envconfig:"konvoy_injector_webhook_server_address"`
	// Port defines the port the https server should be listening on.
	// By default, https server will be listening on port 8443.
	Port uint32 `yaml:"port,omitempty" envconfig:"konvoy_injector_webhook_server_port"`
	// CertDir defines path to a directory with TLS certificate and key for the https server.
	// TLS certificate file must be named `tls.crt`.
	// TLS key file must be named `tls.key`.
	// CertDir has no default value and must always be set explicitly.
	CertDir string `yaml:"certDir,omitempty" envconfig:"konvoy_injector_webhook_server_cert_dir"`
}

var _ config.Config = &Config{}

func (c *Config) Validate() (errs error) {
	if err := c.WebHookServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".WebHookServer is not valid"))
	}
	return
}

var _ config.Config = &WebHookServer{}

func (s *WebHookServer) Validate() (errs error) {
	if s.Address != "" && net.ParseIP(s.Address) == nil {
		errs = multierr.Append(errs, errors.Errorf(".Address must be either empty or a valid IPv4/IPv6 address"))
	}
	if 65535 < s.Port {
		errs = multierr.Append(errs, errors.Errorf(".Port must be in the range [0, 65535]"))
	}
	if s.CertDir == "" {
		errs = multierr.Append(errs, errors.Errorf(".CertDir must be non-empty"))
	}
	return
}
