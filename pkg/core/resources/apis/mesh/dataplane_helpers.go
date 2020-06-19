package mesh

import (
	"net"
	"strings"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
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

// ProtocolList represents a list of Protocols.
type ProtocolList []Protocol

func (l ProtocolList) Strings() []string {
	values := make([]string, len(l))
	for i := range l {
		values[i] = string(l[i])
	}
	return values
}

// SupportedProtocols is a list of supported protocols that will be communicated to a user.
var SupportedProtocols = ProtocolList{
	ProtocolHTTP,
	ProtocolTCP,
}

// Service that indicates L4 pass through cluster
const PassThroughService = "pass_through"

var IPv4Loopback = net.IPv4(127, 0, 0, 1)

func (d *DataplaneResource) UsesInterface(address net.IP, port uint32) bool {
	return d.UsesInboundInterface(address, port) || d.UsesOutboundInterface(address, port)
}

func (d *DataplaneResource) UsesInboundInterface(address net.IP, port uint32) bool {
	if d == nil {
		return false
	}
	ifaces, err := d.Spec.Networking.GetInboundInterfaces()
	if err != nil {
		return false
	}
	for _, iface := range ifaces {
		// compare against port and IP address of the dataplane
		if port == iface.DataplanePort && overlap(address, net.ParseIP(iface.DataplaneIP)) {
			return true
		}
		// compare against port and IP address of the application
		if port == iface.WorkloadPort && overlap(address, IPv4Loopback) {
			return true
		}
	}
	return false
}

func (d *DataplaneResource) UsesOutboundInterface(address net.IP, port uint32) bool {
	if d == nil {
		return false
	}
	ofaces, err := d.Spec.Networking.GetOutboundInterfaces()
	if err != nil {
		return false
	}
	for _, oface := range ofaces {
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

func (d *DataplaneResource) GetPrometheusEndpoint(mesh *MeshResource) (*mesh_proto.PrometheusMetricsBackendConfig, error) {
	if d == nil || mesh == nil || mesh.Meta.GetName() != d.Meta.GetMesh() || !mesh.HasPrometheusMetricsEnabled() {
		return nil, nil
	}
	cfg := mesh_proto.PrometheusMetricsBackendConfig{}
	strCfg := mesh.GetEnabledMetricsBackend().Conf
	if err := util_proto.ToTyped(strCfg, &cfg); err != nil {
		return nil, err
	}

	if d.Spec.GetMetrics().GetType() == mesh_proto.MetricsPrometheusType {
		dpCfg := mesh_proto.PrometheusMetricsBackendConfig{}
		if err := util_proto.ToTyped(d.Spec.Metrics.Conf, &dpCfg); err != nil {
			return nil, err
		}
		proto.Merge(&cfg, &dpCfg)
	}
	return &cfg, nil
}

func (d *DataplaneResource) GetIP() string {
	if d == nil {
		return ""
	}
	return d.Spec.Networking.Address
}
