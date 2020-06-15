package dns_server

import (
	"errors"
	"net"

	"github.com/Kong/kuma/pkg/config"
)

// DNS Server configuration
type DNSServerConfig struct {
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
		Domain: "mesh",
		Port:   5653,
		CIDR:   "240.0.0.0/4",
	}
}
