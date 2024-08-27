package sync

import (
	"context"
	"net"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	manager_dataplane "github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/ordered"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/template"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type DataplaneProxyBuilder struct {
	Zone          string
	APIVersion    core_xds.APIVersion
	IncludeShadow bool
}

func (p *DataplaneProxyBuilder) Build(ctx context.Context, key core_model.ResourceKey, meshContext xds_context.MeshContext) (*core_xds.Proxy, error) {
	dp, found := meshContext.DataplanesByName[key.Name]
	if !found {
		return nil, core_store.ErrorResourceNotFound(core_mesh.DataplaneType, key.Name, key.Mesh)
	}

	routing, destinations, outbounds := p.resolveRouting(ctx, meshContext, dp)

	matchedPolicies, err := p.matchPolicies(meshContext, dp, destinations)
	if err != nil {
		return nil, errors.Wrap(err, "could not match policies")
	}

	matchedPolicies.TrafficRoutes = routing.TrafficRoutes

	meshName := meshContext.Resource.GetMeta().GetName()

	allMeshNames := []string{meshName}
	for _, mesh := range meshContext.Resources.OtherMeshes().Items {
		allMeshNames = append(allMeshNames, mesh.GetMeta().GetName())
	}

	secretsTracker := envoy.NewSecretsTracker(meshName, allMeshNames)

	proxy := &core_xds.Proxy{
		Id:                core_xds.FromResourceKey(key),
		APIVersion:        p.APIVersion,
		Dataplane:         dp,
		Outbounds:         outbounds,
		Routing:           *routing,
		Policies:          *matchedPolicies,
		SecretsTracker:    secretsTracker,
		Metadata:          &core_xds.DataplaneMetadata{},
		Zone:              p.Zone,
		RuntimeExtensions: map[string]interface{}{},
	}
	for k, pl := range core_plugins.Plugins().ProxyPlugins() {
		err := pl.Apply(ctx, meshContext, proxy)
		if err != nil {
			return nil, errors.Wrapf(err, "failed applying proxy plugin: %s", k)
		}
	}
	return proxy, nil
}

func (p *DataplaneProxyBuilder) resolveRouting(
	ctx context.Context,
	meshContext xds_context.MeshContext,
	dataplane *core_mesh.DataplaneResource,
) (*core_xds.Routing, core_xds.DestinationMap, []*core_xds.Outbound) {
	matchedExternalServices := permissions.MatchExternalServicesTrafficPermissions(dataplane, meshContext.Resources.ExternalServices(), meshContext.Resources.TrafficPermissions())

	outbounds := p.resolveVIPOutbounds(meshContext, dataplane)

	// pick a single the most specific route for each outbound interface
	routes := xds_topology.BuildRouteMap(dataplane, meshContext.Resources.TrafficRoutes().Items)

	// create a map of selectors to match other dataplanes reachable via given routes
	destinations := xds_topology.BuildDestinationMap(dataplane, routes)

	endpointMap := xds_topology.BuildExternalServicesEndpointMap(
		ctx,
		meshContext.Resource,
		matchedExternalServices,
		meshContext.DataSourceLoader,
		p.Zone,
	)
	routing := &core_xds.Routing{
		TrafficRoutes:                  routes,
		OutboundTargets:                meshContext.EndpointMap,
		ExternalServiceOutboundTargets: endpointMap,
	}
	return routing, destinations, outbounds
}

func (p *DataplaneProxyBuilder) resolveVIPOutbounds(meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource) []*core_xds.Outbound {
	if dataplane.Spec.Networking.GetTransparentProxying() == nil {
		newOutbounds := []*core_xds.Outbound{}
		for _, o := range dataplane.Spec.Networking.Outbound {
			newOutbounds = append(newOutbounds, &core_xds.Outbound{LegacyOutbound: o})
		}
		return newOutbounds
	}
	reachableServices := map[string]bool{}
	for _, reachableService := range dataplane.Spec.Networking.TransparentProxying.ReachableServices {
		reachableServices[reachableService] = true
	}
	reachableBackends := meshContext.GetReachableBackends(dataplane)

	// Update the outbound of the dataplane with the generatedVips
	generatedVips := map[string]bool{}
	for _, ob := range meshContext.VIPOutbounds {
		generatedVips[ob.LegacyOutbound.Address] = true
	}
	dpTagSets := dataplane.Spec.SingleValueTagSets()
	var newOutbounds []*core_xds.Outbound
	var legacyOutbounds []*mesh_proto.Dataplane_Networking_Outbound
	for _, outbound := range meshContext.VIPOutbounds {
		if outbound.LegacyOutbound.BackendRef == nil {
			if reachableBackends != nil && len(reachableServices) == 0 {
				continue
			}
			service := outbound.LegacyOutbound.GetService()
			if len(reachableServices) != 0 {
				if !reachableServices[service] {
					// ignore VIP outbound if reachableServices is defined and not specified
					// Reachable services takes precedence over reachable services graph.
					continue
				}
			} else {
				// static reachable services takes precedence over the graph
				if !xds_context.CanReachFromAny(meshContext.ReachableServicesGraph, dpTagSets, outbound.LegacyOutbound.Tags) {
					continue
				}
			}
		} else {
			// we need to verify if the user has already reachableServices defined, and to don't send additional clusters and ruin the performance
			// of the dataplane
			if len(reachableServices) != 0 && reachableBackends == nil {
				continue
			}
			if reachableBackends != nil {
				backendKey := xds_context.BackendKey{
					Kind: outbound.LegacyOutbound.BackendRef.Kind,
					Name: outbound.LegacyOutbound.BackendRef.Name,
					Port: outbound.LegacyOutbound.BackendRef.Port,
				}
				// check if there is an entry with specific port or without port
				if !pointer.Deref(reachableBackends)[backendKey] && !pointer.Deref(reachableBackends)[xds_context.BackendKey{Kind: outbound.LegacyOutbound.BackendRef.Kind, Name: outbound.LegacyOutbound.BackendRef.Name}] {
					// ignore VIP outbound if reachableServices is defined and not specified
					// Reachable services takes precedence over reachable services graph.
					continue
				}
				// we don't support MeshTrafficPermission for MeshExternalService at the moment
				// TODO: https://github.com/kumahq/kuma/issues/11077
			} else if outbound.LegacyOutbound.BackendRef.Kind != string(common_api.MeshExternalService) {
				// static reachable services takes precedence over the graph
				if !xds_context.CanReachBackendFromAny(meshContext.ReachableServicesGraph, dpTagSets, outbound.LegacyOutbound.BackendRef) {
					continue
				}
			}
		}
		if dataplane.UsesInboundInterface(net.ParseIP(outbound.LegacyOutbound.Address), outbound.LegacyOutbound.Port) {
			// Skip overlapping outbound interface with inbound.
			// This may happen for example with Headless service on Kubernetes (outbound is a PodIP not ClusterIP, so it's the same as inbound).
			continue
		}
		legacyOutbounds = append(legacyOutbounds, outbound.LegacyOutbound)
		newOutbounds = append(newOutbounds, &core_xds.Outbound{LegacyOutbound: outbound.LegacyOutbound})
	}
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		if generatedVips[outbound.Address] { // Useful while we still have resources with computed vip outbounds
			continue
		}
		legacyOutbounds = append(legacyOutbounds, outbound)
		newOutbounds = append(newOutbounds, &core_xds.Outbound{LegacyOutbound: outbound})
	}
	// we still set legacy outbounds for the dataplane to not break old policies that rely on this field
	dataplane.Spec.Networking.Outbound = legacyOutbounds
	return newOutbounds
}

func (p *DataplaneProxyBuilder) matchPolicies(meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource, outboundSelectors core_xds.DestinationMap) (*core_xds.MatchedPolicies, error) {
	additionalInbounds, err := manager_dataplane.AdditionalInbounds(dataplane, meshContext.Resource)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch additional inbounds")
	}
	inbounds := append(dataplane.Spec.GetNetworking().GetInbound(), additionalInbounds...)

	resources := meshContext.Resources
	ratelimits := ratelimits.BuildRateLimitMap(dataplane, inbounds, resources.RateLimits().Items)
	matchedPolicies := &core_xds.MatchedPolicies{
		TrafficPermissions: permissions.BuildTrafficPermissionMap(dataplane, inbounds, resources.TrafficPermissions().Items),
		TrafficLogs:        logs.BuildTrafficLogMap(dataplane, resources.TrafficLogs().Items),
		HealthChecks:       xds_topology.BuildHealthCheckMap(dataplane, outboundSelectors, resources.HealthChecks().Items),
		CircuitBreakers:    xds_topology.BuildCircuitBreakerMap(dataplane, outboundSelectors, resources.CircuitBreakers().Items),
		TrafficTrace:       xds_topology.SelectTrafficTrace(dataplane, resources.TrafficTraces().Items),
		FaultInjections:    faultinjections.BuildFaultInjectionMap(dataplane, inbounds, resources.FaultInjections().Items),
		Retries:            xds_topology.BuildRetryMap(dataplane, resources.Retries().Items, outboundSelectors),
		Timeouts:           xds_topology.BuildTimeoutMap(dataplane, resources.Timeouts().Items),
		RateLimitsInbound:  ratelimits.Inbound,
		RateLimitsOutbound: ratelimits.Outbound,
		ProxyTemplate:      template.SelectProxyTemplate(dataplane, resources.ProxyTemplates().Items),
		Dynamic:            core_xds.PluginOriginatedPolicies{},
	}
	opts := []core_plugins.MatchedPoliciesOption{}
	if p.IncludeShadow {
		opts = append(opts, core_plugins.IncludeShadow())
	}
	for _, p := range core_plugins.Plugins().PolicyPlugins(ordered.Policies) {
		res, err := p.Plugin.MatchedPolicies(dataplane, resources, opts...)
		if err != nil {
			return nil, errors.Wrapf(err, "could not apply policy plugin %s", p.Name)
		}
		if res.Type == "" {
			return nil, errors.Wrapf(err, "matched policy didn't set type for policy plugin %s", p.Name)
		}
		matchedPolicies.Dynamic[res.Type] = res
	}
	return matchedPolicies, nil
}
