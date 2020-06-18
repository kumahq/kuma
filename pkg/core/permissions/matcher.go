package permissions

import (
	"context"

	"github.com/pkg/errors"

	manager_dataplane "github.com/Kong/kuma/pkg/core/managers/apis/dataplane"
	"github.com/Kong/kuma/pkg/core/policy"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

type TrafficPermissionsMatcher struct {
	ResourceManager manager.ReadOnlyResourceManager
}

func (m *TrafficPermissionsMatcher) Match(ctx context.Context, dataplane *mesh_core.DataplaneResource, mesh *mesh_core.MeshResource) (core_xds.TrafficPermissionMap, error) {
	permissions := &mesh_core.TrafficPermissionResourceList{}
	if err := m.ResourceManager.List(ctx, permissions, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve traffic permissions")
	}

	policies := make([]policy.ConnectionPolicy, len(permissions.Items))
	for i, permission := range permissions.Items {
		policies[i] = permission
	}

	additionalInbounds, err := manager_dataplane.AdditionalInbounds(dataplane, mesh)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch additional inbounds")
	}
	inbounds := append(dataplane.Spec.GetNetworking().GetInbound(), additionalInbounds...)
	policyMap := policy.SelectInboundConnectionPolicies(dataplane, inbounds, policies)

	result := core_xds.TrafficPermissionMap{}
	for inbound, connectionPolicy := range policyMap {
		result[inbound] = connectionPolicy.(*mesh_core.TrafficPermissionResource)
	}
	return result, nil
}
