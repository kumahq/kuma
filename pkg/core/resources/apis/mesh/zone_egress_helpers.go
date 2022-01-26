package mesh

import (
	"net"
)

func (r *ZoneEgressResource) UsesInboundInterface(address net.IP, port uint32) bool {
	if r == nil {
		return false
	}

	if port == r.Spec.GetNetworking().GetPort() && overlap(address, net.ParseIP(r.Spec.GetNetworking().GetAddress())) {
		return true
	}

	return false
}

func (r *ZoneEgressResource) IsRemoteEgress(localZone string) bool {
	if r.Spec.GetZone() == "" || r.Spec.GetZone() == localZone {
		return false
	}

	return true
}
