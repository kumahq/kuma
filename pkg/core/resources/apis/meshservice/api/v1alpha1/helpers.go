package v1alpha1

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/core"
	core_vip "github.com/kumahq/kuma/v2/pkg/core/resources/apis/core/vip"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

type Destination struct {
	MeshServiceResource
}

func ToDst(ms *MeshServiceResource) *Destination {
	return &Destination{*ms}
}

// FindPortByName needs to check both name and value at the same time as this is used with BackendRef which can only reference port by value
func (d *Destination) FindPortByName(name string) (core.Port, bool) {
	for _, p := range d.Spec.Ports {
		if pointer.Deref(p.Name) == name {
			return p, true
		}
		if fmt.Sprintf("%d", p.Port) == name {
			return p, true
		}
	}
	return Port{}, false
}

func IsLocalMeshService(ms *MeshServiceResource) bool {
	if len(ms.GetMeta().GetLabels()) == 0 {
		return true // no labels mean that it's a local resource
	}
	origin, ok := ms.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]
	if !ok {
		return true // no zone label mean that it's a local resource
	}
	return origin == string(mesh_proto.ZoneResourceOrigin)
}

var _ core_vip.ResourceHoldingVIPs = &Destination{}

func (d *Destination) VIPs() []string {
	vips := make([]string, len(d.Status.VIPs))
	for i, vip := range d.Status.VIPs {
		vips[i] = vip.IP
	}
	return vips
}

func (d *Destination) AllocateVIP(vip string) {
	d.Status.VIPs = append(d.Status.VIPs, VIP{
		IP: vip,
	})
}

// todo(jakubdyszkiewicz) strongly consider putting this in MeshService object to avoid problems with computation
func (d *Destination) SNIName(systemNamespace string) string {
	displayName := d.GetMeta().GetLabels()[mesh_proto.DisplayName]
	namespace := d.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag]
	origin := d.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]
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
		return d.GetMeta().GetName() + "." + namespace
	}
	return d.GetMeta().GetName()
}

func (d *Destination) AsOutbounds() xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, vip := range d.Status.VIPs {
		for _, port := range d.Spec.Ports {
			outbounds = append(outbounds, &xds_types.Outbound{
				Address:  vip.IP,
				Port:     uint32(port.Port),
				Resource: kri.WithSectionName(kri.From(d), port.GetName()),
			})
		}
	}
	return outbounds
}

func (d *Destination) Domains() *xds_types.VIPDomains {
	if len(d.Status.VIPs) > 0 {
		var domains []string
		for _, addr := range d.Status.Addresses {
			domains = append(domains, addr.Hostname)
		}
		return &xds_types.VIPDomains{
			Address: d.Status.VIPs[0].IP,
			Domains: domains,
		}
	}
	return nil
}

func (d *Destination) GetPorts() []core.Port {
	var ports []core.Port
	for _, port := range d.Spec.Ports {
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

type DestinationList struct {
	MeshServiceResourceList
}

func ToDstList(msList *MeshServiceResourceList) *DestinationList {
	return &DestinationList{*msList}
}

func (l *DestinationList) GetDestinations() []core.Destination {
	var result []core.Destination
	for _, item := range l.Items {
		result = append(result, &Destination{*item})
	}
	return result
}
