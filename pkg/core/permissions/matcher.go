package permissions

import (
	"context"

	"github.com/pkg/errors"

	manager_dataplane "github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	"github.com/kumahq/kuma/pkg/core/policy"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type TrafficPermissionsMatcher struct {
	ResourceManager manager.ReadOnlyResourceManager
}

func (m *TrafficPermissionsMatcher) Match(ctx context.Context, dataplane *mesh_core.DataplaneResource, mesh *mesh_core.MeshResource) (core_xds.TrafficPermissionMap, error) {
	permissions := &mesh_core.TrafficPermissionResourceList{}
	if err := m.ResourceManager.List(ctx, permissions, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve traffic permissions")
	}

	return BuildTrafficPermissionMap(dataplane, mesh, permissions.Items)
}

func BuildTrafficPermissionMap(
	dataplane *mesh_core.DataplaneResource,
	mesh *mesh_core.MeshResource,
	trafficPermissions []*mesh_core.TrafficPermissionResource,
) (core_xds.TrafficPermissionMap, error) {
	policies := make([]policy.ConnectionPolicy, len(trafficPermissions))
	for i, permission := range trafficPermissions {
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

func (m *TrafficPermissionsMatcher) MatchExternalServices(ctx context.Context, dataplane *mesh_core.DataplaneResource, externalServices *mesh_core.ExternalServiceResourceList) ([]*mesh_core.ExternalServiceResource, error) {
	permissions := &mesh_core.TrafficPermissionResourceList{}
	if err := m.ResourceManager.List(ctx, permissions, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve traffic permissions")
	}

	var matchedExternalServices []*mesh_core.ExternalServiceResource

	externalServicePermissions := m.BuildExternalServicesPermissionsMap(externalServices, permissions.Items)
	for _, externalService := range externalServices.Items {
		permission := externalServicePermissions[externalService.GetMeta().GetName()]
		if permission == nil {
			continue
		}
		matched := false
		for _, selector := range permission.Spec.Sources {
			if dataplane.Spec.MatchTags(selector.Match) {
				matched = true
			}
		}
		if matched {
			matchedExternalServices = append(matchedExternalServices, externalService)
		}
	}
	return matchedExternalServices, nil
}

type ExternalServicePermissions map[string]*mesh_core.TrafficPermissionResource

func (m *TrafficPermissionsMatcher) BuildExternalServicesPermissionsMap(externalServices *mesh_core.ExternalServiceResourceList, trafficPermissions []*mesh_core.TrafficPermissionResource) ExternalServicePermissions {
	policies := make([]policy.ConnectionPolicy, len(trafficPermissions))
	for i, permission := range trafficPermissions {
		policies[i] = permission
	}

	result := ExternalServicePermissions{}
	for _, externalService := range externalServices.Items {
		matchedPolicy := policy.SelectInboundConnectionPolicy(externalService.Spec.Tags, policies)
		if matchedPolicy != nil {
			result[externalService.GetMeta().GetName()] = matchedPolicy.(*mesh_core.TrafficPermissionResource)
		}
	}
	return result
}
