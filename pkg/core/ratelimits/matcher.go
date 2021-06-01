package ratelimits

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

type RateLimitsMatcher struct {
	ResourceManager manager.ReadOnlyResourceManager
}

func (m *RateLimitsMatcher) Match(ctx context.Context, dataplane *mesh_core.DataplaneResource, mesh *mesh_core.MeshResource) (core_xds.RateLimitMap, error) {
	permissions := &mesh_core.RateLimitResourceList{}
	if err := m.ResourceManager.List(ctx, permissions, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve traffic permissions")
	}

	return BuildRateLimitMap(dataplane, mesh, permissions.Items)
}

func BuildRateLimitMap(
	dataplane *mesh_core.DataplaneResource,
	mesh *mesh_core.MeshResource,
	rateLimits []*mesh_core.RateLimitResource,
) (core_xds.RateLimitMap, error) {
	policies := make([]policy.ConnectionPolicy, len(rateLimits))
	for i, permission := range rateLimits {
		policies[i] = permission
	}

	additionalInbounds, err := manager_dataplane.AdditionalInbounds(dataplane, mesh)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch additional inbounds")
	}
	inbounds := append(dataplane.Spec.GetNetworking().GetInbound(), additionalInbounds...)
	policyMap := policy.SelectInboundConnectionPolicies(dataplane, inbounds, policies)

	result := core_xds.RateLimitMap{}
	for inbound, connectionPolicy := range policyMap {
		result[inbound] = connectionPolicy.(*mesh_core.RateLimitResource).Spec
	}
	return result, nil
}

func (m *RateLimitsMatcher) MatchExternalServices(ctx context.Context, dataplane *mesh_core.DataplaneResource, externalServices *mesh_core.ExternalServiceResourceList) ([]*mesh_core.ExternalServiceResource, error) {
	permissions := &mesh_core.RateLimitResourceList{}
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

type ExternalServicePermissions map[string]*mesh_core.RateLimitResource

func (m *RateLimitsMatcher) BuildExternalServicesPermissionsMap(externalServices *mesh_core.ExternalServiceResourceList, rateLimits []*mesh_core.RateLimitResource) ExternalServicePermissions {
	policies := make([]policy.ConnectionPolicy, len(rateLimits))
	for i, permission := range rateLimits {
		policies[i] = permission
	}

	result := ExternalServicePermissions{}
	for _, externalService := range externalServices.Items {
		matchedPolicy := policy.SelectInboundConnectionPolicy(externalService.Spec.Tags, policies)
		if matchedPolicy != nil {
			result[externalService.GetMeta().GetName()] = matchedPolicy.(*mesh_core.RateLimitResource)
		}
	}
	return result
}
