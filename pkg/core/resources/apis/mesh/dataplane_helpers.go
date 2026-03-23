package mesh

import (
	"hash/fnv"
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	k8s_metadata "github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	tproxy_config "github.com/kumahq/kuma/v2/pkg/transparentproxy/config"
	tproxy_dp "github.com/kumahq/kuma/v2/pkg/transparentproxy/config/dataplane"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
)

func (r *DataplaneResource) UsesInterface(address net.IP, port uint32) bool {
	return r.UsesInboundInterface(address, port) || r.UsesOutboundInterface(address, port)
}

func (r *DataplaneResource) UsesInboundInterface(address net.IP, port uint32) bool {
	if r == nil {
		return false
	}
	for _, iface := range r.Spec.Networking.GetInboundInterfaces() {
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

func (r *DataplaneResource) UsesOutboundInterface(address net.IP, port uint32) bool {
	if r == nil {
		return false
	}
	for _, oface := range r.Spec.Networking.GetOutboundInterfaces() {
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

func (r *DataplaneResource) GetPrometheusConfig(mesh *MeshResource) (*mesh_proto.PrometheusMetricsBackendConfig, error) {
	if r == nil || mesh == nil || mesh.Meta.GetName() != r.Meta.GetMesh() || !mesh.HasPrometheusMetricsEnabled() {
		return nil, nil
	}
	cfg := mesh_proto.PrometheusMetricsBackendConfig{}
	strCfg := mesh.GetEnabledMetricsBackend().Conf
	if err := util_proto.ToTyped(strCfg, &cfg); err != nil {
		return nil, err
	}

	if r.Spec.GetMetrics().GetType() == mesh_proto.MetricsPrometheusType {
		dpCfg := mesh_proto.PrometheusMetricsBackendConfig{}
		if err := util_proto.ToTyped(r.Spec.Metrics.Conf, &dpCfg); err != nil {
			return nil, err
		}
		r.mergeLists(&cfg, &dpCfg)
		proto.Merge(&cfg, &dpCfg)
	}
	return &cfg, nil
}

// After proto.Merge called two lists are merged and we cannot be sure
// of order of the elements and if the element is from the Mesh or from
// the Dataplane resource.
func (*DataplaneResource) mergeLists(
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

func (r *DataplaneResource) GetIP() string {
	if r == nil {
		return ""
	}
	if r.Spec.Networking.AdvertisedAddress != "" {
		return r.Spec.Networking.AdvertisedAddress
	} else {
		return r.Spec.Networking.Address
	}
}

func (r *DataplaneResource) IsIPv6() bool {
	return r != nil && govalidator.IsIPv6(r.Spec.GetNetworking().GetAddress())
}

func (r *DataplaneResource) GetAddress() string {
	if r == nil || r.Spec == nil {
		return ""
	}

	return r.Spec.GetNetworking().GetAddress()
}

func (r *DataplaneResource) GetTransparentProxy() *tproxy_dp.DataplaneConfig {
	if r == nil {
		return &tproxy_dp.DataplaneConfig{}
	}

	if tp := r.Spec.GetNetworking().GetTransparentProxying(); tp != nil {
		return &tproxy_dp.DataplaneConfig{
			IPFamilyMode: tproxy_config.IPFamilyModeFromStringer(tp.GetIpFamilyMode()),
			Redirect: tproxy_dp.DataplaneRedirect{
				Inbound:  tproxy_dp.DataplaneTrafficFlowFromPortLike(tp.GetRedirectPortInbound()),
				Outbound: tproxy_dp.DataplaneTrafficFlowFromPortLike(tp.GetRedirectPortOutbound()),
			},
		}
	}

	return &tproxy_dp.DataplaneConfig{}
}

func (r *DataplaneResource) AdminAddress(defaultAdminPort uint32) string {
	if r == nil {
		return ""
	}
	ip := r.GetIP()
	adminPort := r.AdminPort(defaultAdminPort)
	return net.JoinHostPort(ip, strconv.FormatUint(uint64(adminPort), 10))
}

func (r *DataplaneResource) AdminPort(defaultAdminPort uint32) uint32 {
	if r == nil {
		return 0
	}
	if adminPort := r.Spec.GetNetworking().GetAdmin().GetPort(); adminPort != 0 {
		return adminPort
	}
	return defaultAdminPort
}

func (r *DataplaneResource) Hash() []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMeta(r))
	_, _ = hasher.Write([]byte(r.Spec.GetNetworking().GetAddress()))
	_, _ = hasher.Write([]byte(r.Spec.GetNetworking().GetAdvertisedAddress()))
	return hasher.Sum(nil)
}

// InboundIdentifyingName returns a dataplane KRI with portName as section name
// when inbound tags are disabled, falling back to IdentifyingName otherwise.
func (r *DataplaneResource) InboundIdentifyingName(inboundTagsDisabled bool, portName string) string {
	if inboundTagsDisabled && portName != "" {
		id := kri.WithSectionName(kri.FromResourceMeta(r.GetMeta(), DataplaneType), portName)
		if !id.IsEmpty() {
			return id.String()
		}
	}
	return r.IdentifyingName(inboundTagsDisabled)
}

// IdentifyingName returns the workload label when inbound tags are disabled,
// falling back to the identifying service name.
func (r *DataplaneResource) IdentifyingName(inboundTagsDisabled bool) string {
	if inboundTagsDisabled {
		if workload := r.GetMeta().GetLabels()[k8s_metadata.KumaWorkload]; workload != "" {
			return workload
		}
	}
	services := r.Spec.TagSet().Values(mesh_proto.ServiceTag)
	if len(services) > 0 {
		return services[0]
	}
	return mesh_proto.ServiceUnknown
}

// SortDataplanes sorts dataplanes by creation time, then by name.
// Used by generators to ensure consistent processing order.
func SortDataplanes(dps []*DataplaneResource) []*DataplaneResource {
	sorted := slices.Clone(dps)
	slices.SortFunc(sorted, func(a, b *DataplaneResource) int {
		if a, b := a.Meta.GetCreationTime(), b.Meta.GetCreationTime(); a.Before(b) {
			return -1
		} else if a.After(b) {
			return 1
		}
		return strings.Compare(a.Meta.GetName(), b.Meta.GetName())
	})
	return sorted
}
