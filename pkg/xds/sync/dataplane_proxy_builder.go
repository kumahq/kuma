package sync

import (
	"context"
	"net"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/logs"
	manager_dataplane "github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_resources "github.com/kumahq/kuma/pkg/core/resources/apis/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshextenralservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/ordered"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
	tproxy_dp "github.com/kumahq/kuma/pkg/transparentproxy/config/dataplane"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/template"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type DataplaneProxyBuilder struct {
	Zone              string
	APIVersion        core_xds.APIVersion
	InternalAddresses []core_xds.InternalAddress
	IncludeShadow     bool
}

func (p *DataplaneProxyBuilder) Build(ctx context.Context, key core_model.ResourceKey, meta *core_xds.DataplaneMetadata, meshContext xds_context.MeshContext) (*core_xds.Proxy, error) {
	dp, found := meshContext.DataplanesByName[key.Name]
	if !found {
		return nil, core_store.ErrorResourceNotFound(core_mesh.DataplaneType, key.Name, key.Mesh)
	}

	tpEnabled := tproxy_dp.GetDataplaneConfig(dp, meta).Enabled()
	routing, destinations, outbounds := p.resolveRouting(ctx, meshContext, dp, tpEnabled, meta.HasFeature(xds_types.FeatureBindOutbounds))

	matchedPolicies, err := p.matchPolicies(meshContext, dp, destinations)
	if err != nil {
		return nil, errors.Wrap(err, "could not match policies")
	}

	matchedPolicies.TrafficRoutes = routing.TrafficRoutes

	meshName := meshContext.Resource.GetMeta().GetName()

	allMeshNames := []string{}
	for _, mesh := range meshContext.Resources.Meshes().Items {
		allMeshNames = append(allMeshNames, mesh.GetMeta().GetName())
	}

	secretsTracker := envoy.NewSecretsTracker(meshName, allMeshNames)

	proxy := &core_xds.Proxy{
		Id:                core_xds.FromResourceKey(key),
		APIVersion:        p.APIVersion,
		InternalAddresses: p.InternalAddresses,
		Dataplane:         dp,
		Outbounds:         outbounds,
		Routing:           *routing,
		Policies:          *matchedPolicies,
		SecretsTracker:    secretsTracker,
		Metadata:          meta,
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
	tpEnabled bool,
	bindOutbounds bool,
) (*core_xds.Routing, core_xds.DestinationMap, []*xds_types.Outbound) {
	matchedExternalServices := permissions.MatchExternalServicesTrafficPermissions(dataplane, meshContext.Resources.ExternalServices(), meshContext.Resources.TrafficPermissions())

	outbounds := p.resolveVIPOutbounds(meshContext, dataplane, tpEnabled, bindOutbounds)

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

func (p *DataplaneProxyBuilder) resolveVIPOutbounds(
	meshContext xds_context.MeshContext,
	dataplane *core_mesh.DataplaneResource,
	tpEnabled bool,
	bindOutbounds bool,
) []*xds_types.Outbound {
	if !tpEnabled && !bindOutbounds {
		return asOutbounds(dataplane, meshContext.ResolveResourceIdentifier)
	}
	reachableServices := map[string]bool{}
	var reachableBackends map[kri.Identifier]core_resources.Port
	var onlySelectedBackends bool
	if dataplane.Spec.GetNetworking().GetTransparentProxying() != nil {
		for _, reachableService := range dataplane.Spec.GetNetworking().GetTransparentProxying().GetReachableServices() {
			reachableServices[reachableService] = true
		}
		reachableBackends, onlySelectedBackends = meshContext.BaseMeshContext.DestinationIndex.GetReachableBackends(dataplane)
	}

	// Update the outbound of the dataplane with the generatedVips
	generatedVips := map[string]bool{}
	for _, ob := range meshContext.VIPOutbounds {
		generatedVips[ob.GetAddress()] = true
	}
	dpTagSets := dataplane.Spec.SingleValueTagSets()
	var newOutbounds []*xds_types.Outbound
	var legacyOutbounds []*mesh_proto.Dataplane_Networking_Outbound
	for _, outbound := range meshContext.VIPOutbounds {
		if outbound.LegacyOutbound != nil {
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
			if len(reachableServices) != 0 && !onlySelectedBackends {
				continue
			}

			// we need to skip adding Mesh*Service outbounds when ReachableBackends are not configured and MeshServicesMode is set to ReachableBackends,
			// so we don't send additional clusters and to not impact performance
			if dataplane.Spec.GetNetworking().GetTransparentProxying().GetReachableBackends() == nil &&
				meshContext.Resource.Spec.MeshServicesMode() == mesh_proto.Mesh_MeshServices_ReachableBackends {
				continue
			}

			if onlySelectedBackends {
				// check if there is an entry with specific port or without port
				_, selected := reachableBackends[pointer.Deref(outbound.Resource)]
				_, selectedBySectionName := reachableBackends[kri.NoSectionName(pointer.Deref(outbound.Resource))]
				if !selected && !selectedBySectionName {
					// ignore VIP outbound if reachableServices is defined and not specified
					// Reachable services takes precedence over reachable services graph.
					continue
				}
				// we don't support MeshTrafficPermission for MeshExternalService at the moment
				// TODO: https://github.com/kumahq/kuma/issues/11077
			} else if outbound.Resource.ResourceType != meshextenralservice_api.MeshExternalServiceType {
				// static reachable services takes precedence over the graph
				if !xds_context.CanReachBackendFromAny(meshContext.ReachableServicesGraph, dpTagSets, *outbound.Resource) {
					continue
				}
			}
		}
		if dataplane.UsesInboundInterface(net.ParseIP(outbound.GetAddress()), outbound.GetPort()) {
			// Skip overlapping outbound interface with inbound.
			// This may happen for example with Headless service on Kubernetes (outbound is a PodIP not ClusterIP, so it's the same as inbound).
			continue
		}
		if outbound.LegacyOutbound != nil {
			legacyOutbounds = append(legacyOutbounds, outbound.LegacyOutbound)
		}
		newOutbounds = append(newOutbounds, outbound)
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

func asOutbounds(dataplane *core_mesh.DataplaneResource, resolver resolve.LabelResourceIdentifierResolver) xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, o := range dataplane.Spec.Networking.Outbound {
		if o.BackendRef != nil {
			// convert proto BackendRef to common_api.BackendRef
			backendRef := common_api.BackendRef{
				TargetRef: common_api.TargetRef{
					Kind:   common_api.TargetRefKind(o.BackendRef.Kind),
					Name:   pointer.To(o.BackendRef.Name),
					Labels: pointer.To(o.BackendRef.Labels),
				},
				Port: pointer.To(o.BackendRef.Port),
			}
			ref, ok := resolve.BackendRef(kri.From(dataplane), backendRef, resolver)
			if !ok {
				continue
			}
			if ref.ReferencesRealResource() {
				outbounds = append(outbounds, &xds_types.Outbound{
					Address:  o.Address,
					Port:     o.Port,
					Resource: ref.RealResourceBackendRef().Resource,
				})
			}
		} else {
			outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: o})
		}
	}
	return outbounds
}
