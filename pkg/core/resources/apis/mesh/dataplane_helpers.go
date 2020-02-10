package mesh

import (
	"net"
	"strings"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

// Protocol identifies a protocol supported by a service.
type Protocol string

const (
	ProtocolUnknown = "<unknown>"
	ProtocolTCP     = "tcp"
	ProtocolHTTP    = "http"
)

func ParseProtocol(tag string) Protocol {
	switch strings.ToLower(tag) {
	case ProtocolHTTP:
		return ProtocolHTTP
	case ProtocolTCP:
		return ProtocolTCP
	default:
		return ProtocolUnknown
	}
}

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
		// compare against port and IP address of the dataplane
		if port == iface.DataplanePort && overlap(address, net.ParseIP(iface.DataplaneIP)) {
			return true
		}
		// compare against port and IP address of the application
		if port == iface.WorkloadPort && overlap(address, ipv4loopback) {
			return true
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
		// compare against port and IP address of the dataplane
		if port == oface.DataplanePort && overlap(address, net.ParseIP(oface.DataplaneIP)) {
			return true
		}
	}
	return false
}

func overlap(address1 net.IP, address2 net.IP) bool {
	if address1.IsUnspecified() || address2.IsUnspecified() {
		// wildcard match (either IPv4 address "0.0.0.0" or the IPv6 address "::")
		return true
	}
	// exact match
	return address1.Equal(address2)
}

func (d *DataplaneResource) GetPrometheusEndpoint(mesh *MeshResource) *mesh_proto.Metrics_Prometheus {
	if d == nil || mesh == nil || mesh.Meta.GetName() != d.Meta.GetMesh() || !mesh.HasPrometheusMetricsEnabled() {
		return nil
	}
	result := &mesh_proto.Metrics_Prometheus{}
	proto.Merge(result, mesh.Spec.GetMetrics().GetPrometheus())
	proto.Merge(result, d.Spec.GetMetrics().GetPrometheus())
	return result
}

func (d *DataplaneResource) GetIP() string {
	if d == nil {
		return ""
	}
	ifaces, err := d.Spec.Networking.GetInboundInterfaces()
	if err != nil {
		return ""
	}
	if len(ifaces) == 0 {
		return ""
	}
	return ifaces[0].DataplaneIP
}

func (d *DataplaneResource) GetProtocol(idx int) Protocol {
	if d == nil {
		return ProtocolTCP
	}
	if idx < 0 || idx > len(d.Spec.Networking.GetInbound())-1 {
		return ProtocolTCP
	}
	iface := d.Spec.Networking.Inbound[idx]
	return ParseProtocol(iface.Tags[mesh_proto.ProtocolTag])
}
