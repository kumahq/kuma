package ingress

import (
	"context"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"reflect"
)

var lastUsedPort uint32 = 10000

func getNewPort() uint32 {
	lastUsedPort++
	return lastUsedPort
}

type inboundSet []*mesh_proto.Dataplane_Networking_Inbound

func (set inboundSet) getBy(tags map[string]string) *mesh_proto.Dataplane_Networking_Inbound {
	for _, in := range set {
		if reflect.DeepEqual(in.Tags, tags) {
			return in
		}
	}
	return nil
}

func GetInbounds(others []*core_mesh.DataplaneResource, old []*mesh_proto.Dataplane_Networking_Inbound) []*mesh_proto.Dataplane_Networking_Inbound {
	inbounds := make([]*mesh_proto.Dataplane_Networking_Inbound, 0, len(others))
	for _, dp := range others {
		if dp.Spec.GetNetworking().GetIngress() != nil {
			continue
		}
		for _, dpInbound := range dp.Spec.GetNetworking().GetInbound() {
			if dup := inboundSet(inbounds).getBy(dpInbound.GetTags()); dup != nil {
				continue
			}
			var port uint32
			if prev := inboundSet(old).getBy(dpInbound.GetTags()); prev != nil {
				port = prev.Port
			} else {
				port = getNewPort()
			}

			inbounds = append(inbounds, &mesh_proto.Dataplane_Networking_Inbound{
				Port: port, //  picked automatically
				Tags: dpInbound.GetTags(),
			})
		}
	}
	return inbounds
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
