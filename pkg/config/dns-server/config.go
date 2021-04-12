package dns_server

import (
	"errors"
	"net"

	"github.com/kumahq/kuma/pkg/config"
)

// DNS Server configuration
type DNSServerConfig struct {
	// Enabled if true then Kuma CP will start a DNS Server on <Port> and will serve DNS requests for '.<Domain>'
	Enabled bool `yaml:"enabled" envconfig:"kuma_dns_server_enabled"`
	// The domain that the server will resolve the services for
	Domain string `yaml:"domain" envconfig:"kuma_dns_server_domain"`
	// Port on which the server is exposed
	Port uint32 `yaml:"port" envconfig:"kuma_dns_server_port"`
	// CIDR used to allocate virtual IPs from
	CIDR string `yaml:"CIDR" envconfig:"kuma_dns_server_cidr"`
}

func (g *DNSServerConfig) Sanitize() {
}

func (g *DNSServerConfig) Validate() error {
	if g.Port > 65535 {
		return errors.New("Port must be in the range [0, 65535]")
	}
	_, _, err := net.ParseCIDR(g.CIDR)
	if err != nil {
		return errors.New("Must provide a valid CIDR")
	}
	return nil
}

var _ config.Config = &DNSServerConfig{}

func DefaultDNSServerConfig() *DNSServerConfig {
	return &DNSServerConfig{
		Enabled: false,
		Domain:  "mesh",
		Port:    5653,
		CIDR:    "240.0.0.0/4",
	}
}
