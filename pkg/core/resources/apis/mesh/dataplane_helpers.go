package mesh

import (
	"net"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

var ipv4loopback = net.IPv4(127, 0, 0, 1)

func (d *DataplaneResource) UsesInterface(address net.IP, port uint32) bool {
	return d.UsesInboundInterface(address, port) || d.UsesOutboundInterface(address, port)
}

func (d *DataplaneResource) UsesInboundInterface(address net.IP, port uint32) bool {
	if d == nil {
		return false
	}
	for _, inbound := range d.Spec.Networking.GetInbound() {
		iface, err := mesh_proto.ParseInboundInterface(inbound.Interface)
		if err != nil {
			continue
		}
		if iface.DataplanePort != port && iface.WorkloadPort != port {
			// no need to compare IP addresses
			continue
		}
		if iface.DataplanePort == port {
			// compare against IP address of dataplane
			inboundAddress := net.ParseIP(iface.DataplaneIP)
			if inboundAddress.IsUnspecified() || address.IsUnspecified() {
				// wildcard match (either IPv4 address "0.0.0.0" or the IPv6 address "::")
				return true
			}
			if inboundAddress.Equal(address) {
				// exact match
				return true
			}
		}
		if iface.WorkloadPort == port {
			// compare against IP address of application
			applicationAddress := ipv4loopback
			if address.IsUnspecified() {
				// wildcard match (either IPv4 address "0.0.0.0" or the IPv6 address "::")
				return true
			}
			if applicationAddress.Equal(address) {
				// exact match
				return true
			}
		}
	}
	return false
}

func (d *DataplaneResource) UsesOutboundInterface(address net.IP, port uint32) bool {
	if d == nil {
		return false
	}
	for _, outbound := range d.Spec.Networking.GetOutbound() {
		oface, err := mesh_proto.ParseOutboundInterface(outbound.Interface)
		if err != nil {
			continue
		}
		if oface.DataplanePort != port {
			// no need to compare IP addresses
			continue
		}
		outboundAddress := net.ParseIP(oface.DataplaneIP)
		if outboundAddress.IsUnspecified() || address.IsUnspecified() {
			// wildcard match (either IPv4 address "0.0.0.0" or the IPv6 address "::")
			return true
		}
		if outboundAddress.Equal(address) {
			// exact match
			return true
		}
	}
	return false
}
