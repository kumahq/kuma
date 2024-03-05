package dns_server

import (
	"errors"
	"net"

	"github.com/kumahq/kuma/pkg/config"
)

// Config defines DNS Server configuration
type Config struct {
	config.BaseConfig

	// The domain that the server will resolve the services for
	Domain string `json:"domain" envconfig:"kuma_dns_server_domain"`
	// CIDR used to allocate virtual IPs from
	CIDR string `json:"CIDR" envconfig:"kuma_dns_server_cidr"`
	// ServiceVipEnabled will create a service "<kuma.io/service>.mesh" dns entry for every service.
	ServiceVipEnabled bool `json:"serviceVipEnabled" envconfig:"kuma_dns_server_service_vip_enabled"`
	// ServiceVipPort the port to use for virtual IP
	ServiceVipPort uint32 `json:"serviceVipPort" envconfig:"kuma_dns_server_service_vip_port"`
}

func (g *Config) Validate() error {
	_, _, err := net.ParseCIDR(g.CIDR)
	if err != nil {
		return errors.New("CIDR must be valid")
	}
	if g.ServiceVipPort == 0 {
		return errors.New("port can't be 0")
	}
	return nil
}

var _ config.Config = &Config{}

func DefaultDNSServerConfig() *Config {
	return &Config{
		ServiceVipEnabled: true,
		Domain:            "mesh",
		CIDR:              "240.0.0.0/4",
		ServiceVipPort:    80,
	}
}
