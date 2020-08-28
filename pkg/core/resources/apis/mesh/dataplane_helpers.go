package mesh

import (
	"net"
	"strings"
	"time"

	"github.com/ghodss/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

// Protocol identifies a protocol supported by a service.
type Protocol string

const (
	ProtocolUnknown = "<unknown>"
	ProtocolTCP     = "tcp"
	ProtocolHTTP    = "http"
	ProtocolHTTP2   = "http2"
	ProtocolGRPC    = "grpc"
)

func ParseProtocol(tag string) Protocol {
	switch strings.ToLower(tag) {
	case ProtocolHTTP:
		return ProtocolHTTP
	case ProtocolHTTP2:
		return ProtocolHTTP2
	case ProtocolTCP:
		return ProtocolTCP
	case ProtocolGRPC:
		return ProtocolGRPC
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
	ProtocolGRPC,
	ProtocolHTTP,
	ProtocolHTTP2,
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
		if port == iface.WorkloadPort && overlap(address, net.ParseIP(iface.WorkloadIP)) {
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

func ParseDataplaneYAML(bytes []byte) (*DataplaneResource, error) {
	resMeta := rest.ResourceMeta{}
	if err := yaml.Unmarshal(bytes, &resMeta); err != nil {
		return nil, err
	}
	dp := &DataplaneResource{}
	if err := util_proto.FromYAML(bytes, dp.GetSpec()); err != nil {
		return nil, err
	}
	dp.SetMeta(meta{
		Name: resMeta.Name,
		Mesh: resMeta.Mesh,
	})
	return dp, nil
}

var _ model.ResourceMeta = &meta{}

// meta is private implementation of ResourceMeta exclusively for parsing Dataplanes
type meta struct {
	Name string
	Mesh string
}

func (m meta) GetName() string {
	return m.Name
}

func (m meta) GetNameExtensions() model.ResourceNameExtensions {
	return model.ResourceNameExtensionsUnsupported
}

func (m meta) GetVersion() string {
	return ""
}

func (m meta) GetMesh() string {
	return m.Mesh
}

func (m meta) GetCreationTime() time.Time {
	return time.Unix(0, 0).UTC() // the date doesn't matter since it is set on server side anyways
}

func (m meta) GetModificationTime() time.Time {
	return time.Unix(0, 0).UTC() // the date doesn't matter since it is set on server side anyways
}
