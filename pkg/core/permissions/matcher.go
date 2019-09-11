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

type MatchedPermissions map[string]*mesh_core.TrafficPermissionResourceList

func (m MatchedPermissions) Get(inbound string) *mesh_core.TrafficPermissionResourceList {
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
	return MatchDataplaneTrafficPermissions(&dataplane.Spec, permissions), nil
}

func MatchDataplaneTrafficPermissions(dataplane *mesh_proto.Dataplane, permissions *mesh_core.TrafficPermissionResourceList) MatchedPermissions {
	matchedPermissions := make(MatchedPermissions)
	for _, inbound := range dataplane.GetNetworking().GetInbound() {
		matchedPermissions[inbound.Interface] = &mesh_core.TrafficPermissionResourceList{
			Items: matchInbound(inbound, permissions),
		}
	}
	return matchedPermissions
}

func matchInbound(inbound *mesh_proto.Dataplane_Networking_Inbound, trafficPermissions *mesh_core.TrafficPermissionResourceList) []*mesh_core.TrafficPermissionResource {
	matchedPerms := []*mesh_core.TrafficPermissionResource{}
	for _, perm := range trafficPermissions.Items {
		matchedRules := []*mesh_proto.TrafficPermission_Rule{}

		for _, rule := range perm.Spec.Rules {
			if len(rule.Sources) == 0 {
				// todo(jakubdyszkiewicz) there shouldn't be any rule with 0 sources. Move to validation logic in a manager
				continue
			}
			for _, dest := range rule.Destinations {
				if inbound.MatchTags(dest.Match) {
					matchedRules = append(matchedRules, rule)
				}
			}
		}

		if len(matchedRules) > 0 {
			// construct copy of the permission resource but only with matched rules
			matchedPerms = append(matchedPerms, &mesh_core.TrafficPermissionResource{
				Meta: perm.Meta,
				Spec: mesh_proto.TrafficPermission{
					Rules: matchedRules,
				},
			})
		}
	}
	return matchedPerms
}
