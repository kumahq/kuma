package mesh

import (
	"hash/fnv"
	"net"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

// Protocol identifies a protocol supported by a service.
type Protocol string

const (
	ProtocolUnknown = "<unknown>"
	ProtocolTCP     = "tcp"
	ProtocolTLS     = "tls"
	ProtocolHTTP    = "http"
	ProtocolHTTP2   = "http2"
	ProtocolGRPC    = "grpc"
	ProtocolKafka   = "kafka"
)

func ParseProtocol(tag string) Protocol {
	switch strings.ToLower(tag) {
	case ProtocolHTTP:
		return ProtocolHTTP
	case ProtocolHTTP2:
		return ProtocolHTTP2
	case ProtocolTCP:
		return ProtocolTCP
	case ProtocolTLS:
		return ProtocolTLS
	case ProtocolGRPC:
		return ProtocolGRPC
	case ProtocolKafka:
		return ProtocolKafka
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
	ProtocolKafka,
	ProtocolTCP,
}

// Service that indicates L4 pass through cluster
const PassThroughService = "pass_through"

var (
	IPv4Loopback = net.IPv4(127, 0, 0, 1)
	IPv6Loopback = net.IPv6loopback
)

func (d *DataplaneResource) UsesInterface(address net.IP, port uint32) bool {
	return d.UsesInboundInterface(address, port) || d.UsesOutboundInterface(address, port)
}

func (d *DataplaneResource) UsesInboundInterface(address net.IP, port uint32) bool {
	if d == nil {
		return false
	}
	for _, iface := range d.Spec.Networking.GetInboundInterfaces() {
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
	for _, oface := range d.Spec.Networking.GetOutboundInterfaces() {
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

func (d *DataplaneResource) GetPrometheusConfig(mesh *MeshResource) (*mesh_proto.PrometheusMetricsBackendConfig, error) {
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
		d.mergeLists(&cfg, &dpCfg)
		proto.Merge(&cfg, &dpCfg)
	}
	return &cfg, nil
}

// After proto.Merge called two lists are merged and we cannot be sure
// of order of the elements and if the element is from the Mesh or from
// the Dataplane resource.
func (d *DataplaneResource) mergeLists(
	meshCfg *mesh_proto.PrometheusMetricsBackendConfig,
	dpCfg *mesh_proto.PrometheusMetricsBackendConfig,
) {
	aggregate := make(map[string]*mesh_proto.PrometheusAggregateMetricsConfig)
	for _, conf := range meshCfg.Aggregate {
		aggregate[conf.Name] = conf
	}
	// override Mesh aggregate configuration with Dataplane
	for _, conf := range dpCfg.Aggregate {
		aggregate[conf.Name] = conf
	}
	// contains all the elements for Dataplane configuration
	var unduplicatedConfig []*mesh_proto.PrometheusAggregateMetricsConfig
	for _, value := range aggregate {
		unduplicatedConfig = append(unduplicatedConfig, value)
	}
	// we cannot set the same values because they are going to be appended
	meshCfg.Aggregate = []*mesh_proto.PrometheusAggregateMetricsConfig{}
	dpCfg.Aggregate = unduplicatedConfig
}

func (d *DataplaneResource) GetIP() string {
	if d == nil {
		return ""
	}
	if d.Spec.Networking.AdvertisedAddress != "" {
		return d.Spec.Networking.AdvertisedAddress
	} else {
		return d.Spec.Networking.Address
	}
}

func (d *DataplaneResource) IsIPv6() bool {
	if d == nil {
		return false
	}

	ip := net.ParseIP(d.Spec.Networking.Address)
	if ip == nil {
		return false
	}

	return ip.To4() == nil
}

func (d *DataplaneResource) IsUsingTransparentProxy() bool {
	if d == nil {
		return false
	}

	tproxy := d.Spec.GetNetworking().GetTransparentProxying()

	switch {
	case tproxy == nil, tproxy.GetRedirectPortInbound() == 0, tproxy.GetRedirectPortOutbound() == 0:
		return false
	case d.IsIPv6():
		return tproxy.GetIpFamilyMode() != mesh_proto.Dataplane_Networking_TransparentProxying_IPv4
	default:
		return true
	}
}

func (d *DataplaneResource) AdminAddress(defaultAdminPort uint32) string {
	if d == nil {
		return ""
	}
	ip := d.GetIP()
	adminPort := d.AdminPort(defaultAdminPort)
	return net.JoinHostPort(ip, strconv.FormatUint(uint64(adminPort), 10))
}

func (d *DataplaneResource) AdminPort(defaultAdminPort uint32) uint32 {
	if d == nil {
		return 0
	}
	if adminPort := d.Spec.GetNetworking().GetAdmin().GetPort(); adminPort != 0 {
		return adminPort
	}
	return defaultAdminPort
}

func (d *DataplaneResource) Hash() []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMeta(d))
	_, _ = hasher.Write([]byte(d.Spec.GetNetworking().GetAddress()))
	_, _ = hasher.Write([]byte(d.Spec.GetNetworking().GetAdvertisedAddress()))
	return hasher.Sum(nil)
}

func (d *DataplaneResource) AsOutbounds(resolver core_model.LabelResourceIdentifierResolver) xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, o := range d.Spec.Networking.Outbound {
		if o.BackendRef != nil {
			// convert proto BackendRef to common_api.BackendRef
			backendRef := common_api.BackendRef{
				TargetRef: common_api.TargetRef{
					Kind:   common_api.TargetRefKind(o.BackendRef.Kind),
					Name:   o.BackendRef.Name,
					Labels: o.BackendRef.Labels,
				},
				Port: pointer.To(o.BackendRef.Port),
			}
			outbounds = append(outbounds, &xds_types.Outbound{
				Address:  o.Address,
				Port:     o.Port,
				Resource: core_model.ResolveBackendRef(d.GetMeta(), backendRef, resolver).Resource,
			})
		} else {
			outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: o})
		}
	}
	return outbounds
}
