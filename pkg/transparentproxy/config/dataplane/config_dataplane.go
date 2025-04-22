package dataplane

import (
	"github.com/asaskevich/govalidator"
	"golang.org/x/exp/constraints"

	core_config "github.com/kumahq/kuma/pkg/config"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	tproxy_config "github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type PortLike interface {
	constraints.Integer | constraints.Float | tproxy_config.Port
}

type DataplaneConfigGetter interface {
	GetTransparentProxy() *DataplaneConfig
}

type DataplaneResourcer interface {
	DataplaneConfigGetter
	GetAddress() string
}

type DataplaneMetadater interface {
	DataplaneConfigGetter
	HasFeature(string) bool
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

type DataplaneRedirect struct {
	Inbound  DatalpaneTrafficFlow `json:"inbound"`
	Outbound DatalpaneTrafficFlow `json:"outbound"`
	DNS      DatalpaneTrafficFlow `json:"dns"`
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
	if c == nil || !c.Redirect.Inbound.Enabled || !c.Redirect.Outbound.Enabled {
		return false
	}
	return c.IPFamilyMode != tproxy_config.IPFamilyModeIPv4 || govalidator.IsIPv4(c.address)
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

func hasFeature(meta DataplaneMetadater, feature string) bool {
	if meta == nil {
		return false
	}
	return meta.HasFeature(feature)
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

func getConfig(cg DataplaneConfigGetter, fallback DataplaneConfig) *DataplaneConfig {
	if cg == nil {
		return pointer.To(fallback)
	}

	if tp := cg.GetTransparentProxy(); tp != nil {
		return tp
	}

	return pointer.To(fallback)
}

func GetDataplaneConfig(dp DataplaneResourcer, meta DataplaneMetadater) *DataplaneConfig {
	dnsPort := getDNSPort(meta)
	address := getAddress(dp)

	if hasFeature(meta, xds_types.FeatureTransparentProxyInDataplaneMetadata) {
		return getConfig(meta, DefaultDataplaneConfig()).
			withDNSPort(dnsPort).
			withAddress(address)
	}

	return getConfig(dp, DataplaneConfig{}).
		withDNSPort(dnsPort).
		withAddress(address)
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
