package dns_server

import (
	"errors"
	"strings"

	"github.com/kumahq/kuma/v3/pkg/config"
)

// Config defines DNS Server configuration
type Config struct {
	config.BaseConfig

	// The domain that the server will resolve the services for
	Domain string `json:"domain" envconfig:"kuma_dns_server_domain"`
	// ServiceVipPort the port to use for virtual IP
	ServiceVipPort uint32 `json:"serviceVipPort" envconfig:"kuma_dns_server_service_vip_port"`
}

func (g *Config) Validate() error {
	if strings.HasPrefix(g.Domain, ".") {
		return errors.New("domain must not start with a dot")
	}
	if g.ServiceVipPort == 0 {
		return errors.New("port can't be 0")
	}
	return nil
}

var _ config.Config = &Config{}

func DefaultDNSServerConfig() *Config {
	return &Config{
		Domain:         "mesh",
		ServiceVipPort: 80,
	}
}
