package dataplane

import (
	"github.com/asaskevich/govalidator"
	"golang.org/x/exp/constraints"

	core_config "github.com/kumahq/kuma/v3/pkg/config"
	tproxy_config "github.com/kumahq/kuma/v3/pkg/transparentproxy/config"
)

type PortLike interface {
	constraints.Integer | constraints.Float | tproxy_config.Port
}

type DataplaneResourcer interface {
	GetAddress() string
}

type DataplaneMetadater interface {
	GetTransparentProxy() *DataplaneConfig
	GetDNSPort() uint32
}

type DatalpaneTrafficFlow struct {
	Enabled bool               `json:"enabled"`
	Port    tproxy_config.Port `json:"port"`
}

func NewDataplaneTrafficFlow[T PortLike](enabled bool, port T) DatalpaneTrafficFlow {
	return DatalpaneTrafficFlow{
		Enabled: enabled,
		Port:    tproxy_config.Port(port),
	}
}

func DataplaneTrafficFlowFromPortLike[T PortLike](port T) DatalpaneTrafficFlow {
	return NewDataplaneTrafficFlow(port > 0, port)
}

// DataplaneVNet is a serialization-only subset of config.VNet. It is
// intentionally separate to decouple the dataplane bootstrap payload from
// the full transparent proxy config (which carries initialization logic
// and additional fields not needed by kuma-dp).
type DataplaneVNet struct {
	Networks []string `json:"networks,omitempty"`
}

type DataplaneRedirect struct {
	Inbound  DatalpaneTrafficFlow `json:"inbound"`
	Outbound DatalpaneTrafficFlow `json:"outbound"`
	DNS      DatalpaneTrafficFlow `json:"dns"`
	VNet     DataplaneVNet        `json:"vnet"`
}

type DataplaneConfig struct {
	core_config.BaseConfig `json:"-"`

	IPFamilyMode tproxy_config.IPFamilyMode `json:"ipFamilyMode"`
	Redirect     DataplaneRedirect          `json:"redirect"`

	address string
}

func (c *DataplaneConfig) withAddress(address string) *DataplaneConfig {
	if c == nil {
		return nil
	}
	c.address = address
	return c
}

func (c *DataplaneConfig) withDNSPort(port uint32) *DataplaneConfig {
	if c == nil {
		return nil
	}
	c.Redirect.DNS = DataplaneTrafficFlowFromPortLike(port)
	return c
}

func (c *DataplaneConfig) Enabled() bool {
	if c == nil || !c.Redirect.Inbound.Enabled && !c.Redirect.Outbound.Enabled {
		return false
	}
	return c.IPFamilyMode != tproxy_config.IPFamilyModeIPv4 || govalidator.IsIPv4(c.address)
}

func (c *DataplaneConfig) HasVNet() bool {
	if c == nil {
		return false
	}
	return len(c.Redirect.VNet.Networks) > 0
}

func (c *DataplaneConfig) EnabledIPv6() bool {
	if c == nil {
		return false
	}
	// IPv4 addresses can always be represented in IPv6 format (as ::ffff:a.b.c.d),
	// so there's no need to verify the format of the address itself.
	// It's enough to check that the configured IP family mode is not IPv4.
	return c.IPFamilyMode != tproxy_config.IPFamilyModeIPv4
}

func getDNSPort(meta DataplaneMetadater) uint32 {
	if meta == nil {
		return 0
	}
	return meta.GetDNSPort()
}

func getAddress(dp DataplaneResourcer) string {
	if dp == nil {
		return ""
	}
	return dp.GetAddress()
}

func GetDataplaneConfig(dp DataplaneResourcer, meta DataplaneMetadater) *DataplaneConfig {
	cfg := &DataplaneConfig{}
	if meta != nil {
		if tp := meta.GetTransparentProxy(); tp != nil {
			cfg = tp
		}
	}

	return cfg.
		withDNSPort(getDNSPort(meta)).
		withAddress(getAddress(dp))
}

func DefaultDataplaneConfig() DataplaneConfig {
	cfg := tproxy_config.DefaultConfig()

	return DataplaneConfig{
		IPFamilyMode: cfg.IPFamilyMode,
		Redirect: DataplaneRedirect{
			Inbound: NewDataplaneTrafficFlow(
				cfg.Redirect.Inbound.Enabled,
				cfg.Redirect.Inbound.Port,
			),
			Outbound: NewDataplaneTrafficFlow(
				cfg.Redirect.Outbound.Enabled,
				cfg.Redirect.Outbound.Port,
			),
			DNS: NewDataplaneTrafficFlow(
				cfg.Redirect.DNS.Enabled,
				cfg.Redirect.DNS.Port,
			),
		},
	}
}
