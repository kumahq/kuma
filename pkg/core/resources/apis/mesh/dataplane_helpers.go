package mesh

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"net"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	k8s_metadata "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	tproxy_config "github.com/kumahq/kuma/v3/pkg/transparentproxy/config"
	tproxy_dp "github.com/kumahq/kuma/v3/pkg/transparentproxy/config/dataplane"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
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
	return d != nil && govalidator.IsIPv6(d.Spec.GetNetworking().GetAddress())
}

func (d *DataplaneResource) GetAddress() string {
	if d == nil || d.Spec == nil {
		return ""
	}

	return d.Spec.GetNetworking().GetAddress()
}

func (d *DataplaneResource) GetTransparentProxy() *tproxy_dp.DataplaneConfig {
	if d == nil {
		return &tproxy_dp.DataplaneConfig{}
	}

	if tp := d.Spec.GetNetworking().GetTransparentProxying(); tp != nil {
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

// Hash returns a content-based hash of the Dataplane used to gate xDS
// regeneration. It intentionally excludes meta.GetVersion() (the Kubernetes
// resourceVersion) so that writes irrelevant to xDS generation - status,
// annotations, managedFields - don't invalidate the mesh-wide xDS context.
// Labels are included because they affect policy matching.
func (d *DataplaneResource) Hash() []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMetaIdentity(d))
	writeSortedLabels(hasher, d.GetMeta().GetLabels())
	specBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(d.Spec)
	if err == nil {
		_, _ = hasher.Write(specBytes)
	} else {
		// Deterministic marshaling should never fail for a well-formed Dataplane
		// spec, but fall back to a value that still changes with the spec
		// instead of silently treating every Dataplane as identical.
		_, _ = hasher.Write([]byte(d.Spec.String()))
	}
	return hasher.Sum(nil)
}

// writeSortedLabels writes labels into hasher in a deterministic,
// unambiguous order regardless of map iteration order. Keys and values are
// length-prefixed so that e.g. {"a":"bc"} and {"ab":"c"} don't collide.
func writeSortedLabels(hasher hash.Hash, labels map[string]string) {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lenBuf [8]byte
	writeLenPrefixed := func(s string) {
		binary.BigEndian.PutUint64(lenBuf[:], uint64(len(s)))
		_, _ = hasher.Write(lenBuf[:])
		_, _ = hasher.Write([]byte(s))
	}

	for _, k := range keys {
		writeLenPrefixed(k)
		writeLenPrefixed(labels[k])
	}
}

// InboundIdentifyingName returns a dataplane KRI with portName as section name
// when inbound tags are disabled, falling back to IdentifyingName otherwise.
func (d *DataplaneResource) InboundIdentifyingName(inboundTagsDisabled bool, portName string) string {
	if inboundTagsDisabled && portName != "" {
		id := kri.WithSectionName(kri.FromResourceMeta(d.GetMeta(), DataplaneType), portName)
		if !id.IsEmpty() {
			return id.String()
		}
	}
	return d.IdentifyingName(inboundTagsDisabled)
}

// IdentifyingName returns the workload label when inbound tags are disabled,
// falling back to the identifying service name.
func (d *DataplaneResource) IdentifyingName(inboundTagsDisabled bool) string {
	if inboundTagsDisabled {
		if workload := d.GetMeta().GetLabels()[k8s_metadata.KumaWorkload]; workload != "" {
			return workload
		}
	}
	services := d.Spec.TagSet().Values(mesh_proto.ServiceTag)
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
