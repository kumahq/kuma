package ingress

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/kds/util"
)

type availableServicesByMesh map[string][]*mesh_proto.Dataplane_Networking_Ingress_AvailableService

// SplitIngressesByMeshAndFlatten takes a list of dataplanes, applies 'splitByMesh' for each Ingress
// and appends resulting ingresses to the return-value list
func SplitIngressesByMeshAndFlatten(dataplanes *core_mesh.DataplaneResourceList) *core_mesh.DataplaneResourceList {
	rv := &core_mesh.DataplaneResourceList{}
	for _, d := range dataplanes.Items {
		if d.Spec.IsIngress() {
			for _, ingress := range splitByMesh(d) {
				_ = rv.AddItem(ingress) // err is ignores because we control the type
			}
		} else {
			_ = rv.AddItem(d)
		}
	}
	return rv
}

// splitByMesh takes Ingress that has AvailableServices from all meshes and
// returns list of Ingresses each of them has AvailableServices only for single mesh.
func splitByMesh(ingress *core_mesh.DataplaneResource) []*core_mesh.DataplaneResource {
	as := availableServicesByMesh{}
	for _, service := range ingress.Spec.Networking.Ingress.AvailableServices {
		mesh, ok := service.Tags["mesh"]
		if !ok {
			mesh = "default"
		}
		as[mesh] = append(as[mesh], service)
	}
	rv := []*core_mesh.DataplaneResource{}
	for mesh, services := range as {
		rv = append(rv, &core_mesh.DataplaneResource{
			Meta: util.ResourceKeyToMeta(ingress.GetMeta().GetName(), mesh),
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Ingress: &mesh_proto.Dataplane_Networking_Ingress{
						AvailableServices: services,
					},
					Address: ingress.Spec.Networking.Address,
					Inbound: ingress.Spec.Networking.Inbound,
				},
			},
		})
	}
	return rv
}
