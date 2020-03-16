package permissions

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

type TrafficPermissionsMatcher struct {
	ResourceManager manager.ResourceManager
}

type MatchedPermissions map[mesh_proto.InboundInterface]*mesh_core.TrafficPermissionResourceList

func (m MatchedPermissions) Get(inbound mesh_proto.InboundInterface) *mesh_core.TrafficPermissionResourceList {
	matched, ok := m[inbound]
	if ok {
		return matched
	} else {
		return &mesh_core.TrafficPermissionResourceList{}
	}
}

func (m *TrafficPermissionsMatcher) Match(ctx context.Context, dataplane *mesh_core.DataplaneResource) (MatchedPermissions, error) {
	permissions := &mesh_core.TrafficPermissionResourceList{}
	if err := m.ResourceManager.List(ctx, permissions, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, err
	}
	return MatchDataplaneTrafficPermissions(&dataplane.Spec, permissions)
}

func MatchDataplaneTrafficPermissions(dataplane *mesh_proto.Dataplane, permissions *mesh_core.TrafficPermissionResourceList) (MatchedPermissions, error) {
	matchedPermissions := make(MatchedPermissions)
	ifaces, err := dataplane.GetNetworking().GetInboundInterfaces()
	if err != nil {
		return nil, err
	}
	for i, inbound := range dataplane.GetNetworking().GetInbound() {
		matchedPermissions[ifaces[i]] = &mesh_core.TrafficPermissionResourceList{
			Items: matchInbound(inbound, permissions),
		}
	}
	return matchedPermissions, nil
}

func matchInbound(inbound *mesh_proto.Dataplane_Networking_Inbound, trafficPermissions *mesh_core.TrafficPermissionResourceList) []*mesh_core.TrafficPermissionResource {
	matchedPerms := []*mesh_core.TrafficPermissionResource{}
	for _, perm := range trafficPermissions.Items {
		if len(perm.Spec.Sources) == 0 {
			continue
		}
		for _, dest := range perm.Spec.Destinations {
			if inbound.MatchTags(dest.Match) {
				matchedPerms = append(matchedPerms, perm)
				break
			}
		}
	}
	return matchedPerms
}
