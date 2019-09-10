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

func (m *TrafficPermissionsMatcher) Match(ctx context.Context, dataplane *mesh_core.DataplaneResource) (*mesh_core.TrafficPermissionResourceList, error) {
	permissions := &mesh_core.TrafficPermissionResourceList{}
	if err := m.ResourceManager.List(ctx, permissions, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, err
	}
	return MatchDataplaneTrafficPermissions(&dataplane.Spec, permissions), nil
}

func MatchDataplaneTrafficPermissions(dataplane *mesh_proto.Dataplane, trafficPermissions *mesh_core.TrafficPermissionResourceList) *mesh_core.TrafficPermissionResourceList {
	matchedPerms := []*mesh_core.TrafficPermissionResource{}

	for _, perm := range trafficPermissions.Items {
		matchedRules := []*mesh_proto.TrafficPermission_Rule{}

		for _, rule := range perm.Spec.Rules {
			for _, dest := range rule.Destinations {
				if len(rule.Sources) > 0 && dataplane.MatchTags(dest.Match) {
					matchedRules = append(matchedRules, &mesh_proto.TrafficPermission_Rule{
						Sources:      rule.Sources,
						Destinations: rule.Destinations,
					})
				}
			}
		}

		if len(matchedRules) > 0 {
			matchedPerms = append(matchedPerms, &mesh_core.TrafficPermissionResource{
				Meta: perm.Meta,
				Spec: mesh_proto.TrafficPermission{
					Rules: matchedRules,
				},
			})
		}
	}

	return &mesh_core.TrafficPermissionResourceList{
		Items: matchedPerms,
	}
}
