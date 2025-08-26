package vips

import (
	"fmt"
	"sort"
	"strings"
)

// VirtualOutbound the description of a hostname -> address and a list of port/tagSet that identifies each outbound.
type VirtualOutbound struct {
	// This is not default in the legacy case (hostnames won't be complete)
	Address   string          `json:"address,omitempty"`
	Outbounds []OutboundEntry `json:"outbounds,omitempty"`
}

func (vo *VirtualOutbound) Equal(other *VirtualOutbound) bool {
	if vo.Address != other.Address || len(vo.Outbounds) != len(other.Outbounds) {
		return false
	}
	for i := range vo.Outbounds {
		if vo.Outbounds[i].String() != other.Outbounds[i].String() {
			return false
		}
	}
	return true
}

const (
	VirtualOutboundPrefix = "virtual-outbound:"
	HostPrefix            = "external-service:"
	GatewayPrefix         = "mesh-gateway:"
)

var (
	OriginVirtualOutbound = func(name string) string { return VirtualOutboundPrefix + name }
	OriginHost            = func(name string) string { return HostPrefix + name }
	OriginGateway         = func(mesh, name, hostname string) string {
		return fmt.Sprintf("%s%s:%s:%s", GatewayPrefix, mesh, name, hostname)
	}
)

type OutboundEntry struct {
	Port   uint32
	TagSet map[string]string
	// A string to identify where this outbound was defined (usually the name of the outbound policy)
	Origin string
}

func (mo *OutboundEntry) Less(o *OutboundEntry) bool {
	return mo.Port < o.Port
}

func (mo *OutboundEntry) String() string {
	var tags []string
	for k, v := range mo.TagSet {
		tags = append(tags, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(tags)
	return fmt.Sprintf("%s=%d{%s}", mo.Origin, mo.Port, strings.Join(tags, ","))
}
