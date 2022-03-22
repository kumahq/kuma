package faultinjections

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	manager_dataplane "github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type FaultInjectionMatcher struct {
	ResourceManager manager.ReadOnlyResourceManager
}

func (f *FaultInjectionMatcher) Match(ctx context.Context, dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) (core_xds.FaultInjectionMap, error) {
	faultInjections := &core_mesh.FaultInjectionResourceList{}
	if err := f.ResourceManager.List(ctx, faultInjections, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve fault injections")
	}
	additionalInbounds, err := manager_dataplane.AdditionalInbounds(dataplane, mesh)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch additional inbounds")
	}
	inbounds := append(dataplane.Spec.GetNetworking().GetInbound(), additionalInbounds...)
	return BuildFaultInjectionMap(dataplane, inbounds, faultInjections.Items), nil
}

func BuildFaultInjectionMap(
	dataplane *core_mesh.DataplaneResource,
	inbounds []*mesh_proto.Dataplane_Networking_Inbound,
	faultInjections []*core_mesh.FaultInjectionResource,
) core_xds.FaultInjectionMap {
	policies := make([]policy.ConnectionPolicy, len(faultInjections))
	for i, faultInjection := range faultInjections {
		policies[i] = faultInjection
	}

	policyMap := policy.SelectInboundConnectionMatchingPolicies(dataplane, inbounds, policies)

	result := core_xds.FaultInjectionMap{}
	for inbound, connectionPolicies := range policyMap {
		for _, connectionPolicy := range connectionPolicies {
			result[inbound] = append(result[inbound], connectionPolicy.(*core_mesh.FaultInjectionResource))
		}
	}
	return result
}

// BuildExternalServiceFaultInjectionMapForZoneEgress creates mapping between Service's name and the list of FaultInjections.
// todo(lobkovilya): that's not really correct way to build a policy map for External Services. Policies such as
// Fault Injections, Rate Limit and Traffic Permissions support arbitrary tags in Destination, but we lose this
// information if putting policies in map just by Service's name. See https://github.com/kumahq/kuma/issues/3999
func BuildExternalServiceFaultInjectionMapForZoneEgress(
	externalServices []*core_mesh.ExternalServiceResource,
	faultInjections []*core_mesh.FaultInjectionResource,
) core_xds.ExternalServiceFaultInjectionMap {
	policies := make([]policy.ConnectionPolicy, len(faultInjections))
	for i, faultInjection := range faultInjections {
		policies[i] = faultInjection
	}

	result := core_xds.ExternalServiceFaultInjectionMap{}
	for _, externalService := range externalServices {
		tags := externalService.Spec.GetTags()
		serviceName := tags[mesh_proto.ServiceTag]

		matchedPolicies := policy.SelectInboundConnectionAllPolicies(tags, policies)
		for _, matchedPolicy := range matchedPolicies {
			result[serviceName] = append(result[serviceName], matchedPolicy.(*core_mesh.FaultInjectionResource))
		}
	}

	for service, resources := range result {
		result[service] = dedup(resources)
	}

	return result
}

func dedup(policies []*core_mesh.FaultInjectionResource) []*core_mesh.FaultInjectionResource {
	seen := map[core_model.ResourceKey]struct{}{}
	result := []*core_mesh.FaultInjectionResource{}
	for _, p := range policies {
		if _, ok := seen[core_model.MetaToResourceKey(p.GetMeta())]; ok {
			continue
		}
		seen[core_model.MetaToResourceKey(p.GetMeta())] = struct{}{}
		result = append(result, p)
	}
	return result
}
