package v1alpha1

import (
	"encoding/json"
	"fmt"
	"hash"
	"hash/fnv"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	core_vip "github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/vip"
	hostnamegenerator_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/sni"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

// FindPortByName needs to check both name and value at the same time as this is used with BackendRef which can only reference port by value
func (m *MeshServiceResource) FindPortByName(name string) (core.Port, bool) {
	for _, p := range m.Spec.Ports {
		if pointer.Deref(p.Name) == name {
			return p, true
		}
		if fmt.Sprintf("%d", p.Port) == name {
			return p, true
		}
	}
	return Port{}, false
}

func (m *MeshServiceResource) IsLocalMeshService() bool {
	if len(m.GetMeta().GetLabels()) == 0 {
		return true // no labels mean that it's a local resource
	}
	origin, ok := m.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]
	if !ok {
		return true // no origin label mean that it's a local resource
	}
	// Origin label is sufficient to detect if a MeshService is local because
	// MeshServices synced from another zone will have `kuma.io/origin: global`,
	// since origin always reflects the last place the resource was received from.
	return origin == string(mesh_proto.ZoneResourceOrigin)
}

// Hash returns a content-based hash of the MeshService used to gate xDS
// regeneration. Like DataplaneResource.Hash, it excludes meta.GetVersion()
// so that status writes irrelevant to xDS - most notably the
// DataplaneProxies counters refreshed by the status updater's 5s ticker -
// don't force mesh-wide xDS recomputation. Addresses, VIPs, TLS and
// HostnameGenerators are still hashed because outbound and cluster
// generation read them directly from Status.
func (t *MeshServiceResource) Hash() []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMetaIdentity(t))
	core_model.WriteSortedLabels(hasher, t.GetMeta().GetLabels())
	writeJSON(hasher, t.Spec)
	writeJSON(hasher, struct {
		Addresses          []hostnamegenerator_api.Address
		VIPs               []VIP
		TLS                TLS
		HostnameGenerators []hostnamegenerator_api.HostnameGeneratorStatus
	}{
		Addresses:          t.Status.Addresses,
		VIPs:               t.Status.VIPs,
		TLS:                t.Status.TLS,
		HostnameGenerators: t.Status.HostnameGenerators,
	})
	return hasher.Sum(nil)
}

// writeJSON writes a deterministic JSON encoding of v into hasher.
// encoding/json sorts map keys, so this is stable regardless of map
// iteration order.
func writeJSON(hasher hash.Hash, v any) {
	b, err := json.Marshal(v)
	if err == nil {
		_, _ = hasher.Write(b)
	} else {
		// Marshaling should never fail for these plain data structs, but fall
		// back to a value that still changes with content instead of
		// silently treating every MeshService as identical.
		_, _ = fmt.Fprintf(hasher, "%+v", v)
	}
}

var _ core_vip.ResourceHoldingVIPs = &MeshServiceResource{}

func (t *MeshServiceResource) VIPs() []string {
	vips := make([]string, len(t.Status.VIPs))
	for i, vip := range t.Status.VIPs {
		vips[i] = vip.IP
	}
	return vips
}

func (t *MeshServiceResource) AllocateVIP(vip string) {
	t.Status.VIPs = append(t.Status.VIPs, VIP{
		IP: vip,
	})
}

// todo(jakubdyszkiewicz) strongly consider putting this in MeshService object to avoid problems with computation
func (t *MeshServiceResource) SNIName(systemNamespace string) string {
	displayName := t.GetMeta().GetLabels()[mesh_proto.DisplayName]
	namespace := t.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag]
	origin := t.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]
	if origin == string(mesh_proto.GlobalResourceOrigin) {
		// we need to use original name and namespace for services that were synced from another cluster
		sniName := displayName
		if namespace != "" {
			// when we sync resources from universal to kube, when we retrieve it has KubeNamespaceTag as label value
			if systemNamespace == "" || systemNamespace != namespace {
				sniName += "." + namespace
			}
		}
		return sniName
	}
	if systemNamespace == "" && origin == string(mesh_proto.ZoneResourceOrigin) && namespace != "" {
		// when namespace label was added to Universal MeshService to have a copy of Kubernets MeshService
		return t.GetMeta().GetName() + "." + namespace
	}
	return t.GetMeta().GetName()
}

func (t *MeshServiceResource) AsOutbounds() xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, vip := range t.Status.VIPs {
		for _, port := range t.Spec.Ports {
			outbounds = append(outbounds, &xds_types.Outbound{
				Address:  vip.IP,
				Port:     uint32(port.Port),
				Resource: kri.WithSectionName(kri.From(t), port.GetName()),
			})
		}
	}
	return outbounds
}

func (t *MeshServiceResource) Domains() *xds_types.VIPDomains {
	if len(t.Status.VIPs) > 0 {
		var domains []string
		for _, addr := range t.Status.Addresses {
			domains = append(domains, addr.Hostname)
		}
		return &xds_types.VIPDomains{
			Address: t.Status.VIPs[0].IP,
			Domains: domains,
		}
	}
	return nil
}

func (t *MeshServiceResource) GetPorts() []core.Port {
	var ports []core.Port
	for _, port := range t.Spec.Ports {
		ports = append(ports, core.Port(port))
	}
	return ports
}

func (p Port) GetName() string {
	return pointer.DerefOr(p.Name, fmt.Sprintf("%d", p.Port))
}

func (p Port) GetValue() int32 {
	return p.Port
}

func (p Port) GetProtocol() core_meta.Protocol {
	return p.AppProtocol
}

func (s *MeshService) SNIs() []sni.Section {
	if s == nil {
		return nil
	}
	out := make([]sni.Section, 0, len(s.Ports))
	for _, p := range s.Ports {
		out = append(out, sni.Section{Port: p.Port, SectionName: p.GetName()})
	}
	return out
}

func (l *MeshServiceResourceList) GetDestinations() []core.Destination {
	var result []core.Destination
	for _, item := range l.Items {
		result = append(result, item)
	}
	return result
}
