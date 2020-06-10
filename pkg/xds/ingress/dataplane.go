package ingress

import (
	"context"
	"reflect"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

type ingressSet []*mesh_proto.Dataplane_Networking_Ingress

func (set ingressSet) getBy(tags map[string]string) *mesh_proto.Dataplane_Networking_Ingress {
	for _, in := range set {
		if reflect.DeepEqual(in.Tags, tags) {
			return in
		}
	}
	return nil
}

func GetIngressByDataplanes(others []*core_mesh.DataplaneResource) []*mesh_proto.Dataplane_Networking_Ingress {
	ingresses := make([]*mesh_proto.Dataplane_Networking_Ingress, 0, len(others))
	for _, dp := range others {
		if dp.Spec.GetNetworking().GetIngress() != nil {
			continue
		}
		for _, dpInbound := range dp.Spec.GetNetworking().GetInbound() {
			if dup := ingressSet(ingresses).getBy(dpInbound.GetTags()); dup != nil {
				continue
			}
			ingresses = append(ingresses, &mesh_proto.Dataplane_Networking_Ingress{
				Service: dpInbound.Tags[mesh_proto.ServiceTag],
				Tags:    mesh_proto.SingleValueTagSet(dpInbound.Tags).Exclude(mesh_proto.ServiceTag),
			})
		}
	}
	return ingresses
}

func GetAllDataplanes(resourceManager manager.ReadOnlyResourceManager) ([]*core_mesh.DataplaneResource, error) {
	ctx := context.Background()
	meshes := &core_mesh.MeshResourceList{}
	if err := resourceManager.List(ctx, meshes); err != nil {
		return nil, err
	}
	dataplanes := make([]*core_mesh.DataplaneResource, 0)
	for _, mesh := range meshes.Items {
		dataplaneList := &core_mesh.DataplaneResourceList{}
		if err := resourceManager.List(ctx, dataplaneList, store.ListByMesh(mesh.Meta.GetName())); err != nil {
			return nil, err
		}
		dataplanes = append(dataplanes, dataplaneList.Items...)
	}
	return dataplanes, nil
}
