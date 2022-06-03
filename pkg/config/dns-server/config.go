package dns_server

import (
	"errors"
	"net"

	"github.com/kumahq/kuma/pkg/config"
)

// DNS Server configuration
type Config struct {
	// The domain that the server will resolve the services for
	Domain string `yaml:"domain" envconfig:"kuma_dns_server_domain"`
	// CIDR used to allocate virtual IPs from
	CIDR string `yaml:"CIDR" envconfig:"kuma_dns_server_cidr"`
	// ServiceVipEnabled will create a service "<kuma.io/service>.mesh" dns entry for every service.
	ServiceVipEnabled bool `yaml:"serviceVipEnabled" envconfig:"kuma_dns_server_service_vip_enabled"`
}

func (g *Config) Sanitize() {
}

func (g *Config) Validate() error {
	_, _, err := net.ParseCIDR(g.CIDR)
	if err != nil {
		return errors.New("Must provide a valid CIDR")
	}
	return nil
}

var _ config.Config = &Config{}

func DefaultDNSServerConfig() *Config {
	return &Config{
		ServiceVipEnabled: true,
		Domain:            "mesh",
		CIDR:              "240.0.0.0/4",
	}
}
