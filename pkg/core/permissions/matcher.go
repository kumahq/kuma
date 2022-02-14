package permissions

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	manager_dataplane "github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type TrafficPermissionsMatcher struct {
	ResourceManager manager.ReadOnlyResourceManager
}

func (m *TrafficPermissionsMatcher) Match(ctx context.Context, dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) (core_xds.TrafficPermissionMap, error) {
	permissions := &core_mesh.TrafficPermissionResourceList{}
	if err := m.ResourceManager.List(ctx, permissions, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve traffic permissions")
	}

	additionalInbounds, err := manager_dataplane.AdditionalInbounds(dataplane, mesh)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch additional inbounds")
	}
	inbounds := append(dataplane.Spec.GetNetworking().GetInbound(), additionalInbounds...)
	return BuildTrafficPermissionMap(dataplane, inbounds, permissions.Items), nil
}

func BuildTrafficPermissionMap(
	dataplane *core_mesh.DataplaneResource,
	inbounds []*mesh_proto.Dataplane_Networking_Inbound,
	trafficPermissions []*core_mesh.TrafficPermissionResource,
) core_xds.TrafficPermissionMap {
	policies := make([]policy.ConnectionPolicy, len(trafficPermissions))
	for i, permission := range trafficPermissions {
		policies[i] = permission
	}
	policyMap := policy.SelectInboundConnectionPolicies(dataplane, inbounds, policies)

	result := core_xds.TrafficPermissionMap{}
	for inbound, connectionPolicy := range policyMap {
		result[inbound] = connectionPolicy.(*core_mesh.TrafficPermissionResource)
	}
	return result
}

func MatchExternalServices(
	dataplane *core_mesh.DataplaneResource,
	externalServices *core_mesh.ExternalServiceResourceList,
	permissions *core_mesh.TrafficPermissionResourceList,
) ([]*core_mesh.ExternalServiceResource, error) {
	var matchedExternalServices []*core_mesh.ExternalServiceResource

	externalServicePermissions := BuildExternalServicesPermissionsMap(externalServices, permissions.Items)
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

type ExternalServicePermissions map[string]*core_mesh.TrafficPermissionResource

func BuildExternalServicesPermissionsMap(externalServices *core_mesh.ExternalServiceResourceList, trafficPermissions []*core_mesh.TrafficPermissionResource) ExternalServicePermissions {
	policies := make([]policy.ConnectionPolicy, len(trafficPermissions))
	for i, permission := range trafficPermissions {
		policies[i] = permission
	}

	result := ExternalServicePermissions{}
	for _, externalService := range externalServices.Items {
		matchedPolicy := policy.SelectInboundConnectionPolicy(externalService.Spec.Tags, policies)
		if matchedPolicy != nil {
			result[externalService.GetMeta().GetName()] = matchedPolicy.(*core_mesh.TrafficPermissionResource)
		}
	}
	return result
}

// BuildExternalServicesPermissionsMapForZoneEgress is necessary for zone egress
// where we expect to have the permission map with keys equal kuma.io/service tag's
// value.
// Zone Egress currently cannot differentiate different external services with the same
// kuma.io/service tags
func BuildExternalServicesPermissionsMapForZoneEgress(
	externalServices []*core_mesh.ExternalServiceResource,
	trafficPermissions []*core_mesh.TrafficPermissionResource,
) core_xds.ExternalServicePermissionMap {
	policies := make([]policy.ConnectionPolicy, len(trafficPermissions))
	for i, permission := range trafficPermissions {
		policies[i] = permission
	}

	result := core_xds.ExternalServicePermissionMap{}
	for _, externalService := range externalServices {
		tags := externalService.Spec.GetTags()
		serviceName := tags[mesh_proto.ServiceTag]

		matchedPolicy := policy.SelectInboundConnectionPolicy(tags, policies)
		if matchedPolicy != nil {
			result[serviceName] = matchedPolicy.(*core_mesh.TrafficPermissionResource)
		}
	}

	return result
}
