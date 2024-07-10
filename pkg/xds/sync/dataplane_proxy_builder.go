package sync

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
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
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/template"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type DataplaneProxyBuilder struct {
	Zone           string
	APIVersion     core_xds.APIVersion
	IncludeShadow  bool
	UseMeshService bool
}

func (p *DataplaneProxyBuilder) Build(ctx context.Context, key core_model.ResourceKey, meshContext xds_context.MeshContext) (*core_xds.Proxy, error) {
	dp, found := meshContext.DataplanesByName[key.Name]
	if !found {
		return nil, core_store.ErrorResourceNotFound(core_mesh.DataplaneType, key.Name, key.Mesh)
	}

	routing, destinations := p.resolveRouting(ctx, meshContext, dp)

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
) (*core_xds.Routing, core_xds.DestinationMap) {
	matchedExternalServices := permissions.MatchExternalServicesTrafficPermissions(dataplane, meshContext.Resources.ExternalServices(), meshContext.Resources.TrafficPermissions())

	p.resolveVIPOutbounds(meshContext, dataplane)

	// pick a single the most specific route for each outbound interface
	routes := xds_topology.BuildRouteMap(dataplane, meshContext.Resources.TrafficRoutes().Items)

	// create a map of selectors to match other dataplanes reachable via given routes
	destinations := xds_topology.BuildDestinationMap(dataplane, routes)

	endpointMap := xds_topology.BuildExternalServicesEndpointMap(
		ctx,
		meshContext.Resource,
		matchedExternalServices,
		meshContext.Resources.MeshExternalServices().Items,
		meshContext.DataSourceLoader,
		p.Zone,
	)
	routing := &core_xds.Routing{
		TrafficRoutes:                  routes,
		OutboundTargets:                meshContext.EndpointMap,
		ExternalServiceOutboundTargets: endpointMap,
	}
	return routing, destinations
}

func (p *DataplaneProxyBuilder) resolveVIPOutbounds(meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource) {
	if dataplane.Spec.Networking.GetTransparentProxying() == nil {
		return
	}
	reachableServices := map[string]bool{}
	for _, reachableService := range dataplane.Spec.Networking.TransparentProxying.ReachableServices {
		reachableServices[reachableService] = true
	}

	reachableBackends := GetReachableBackends(meshContext, dataplane)
	for key, val := range reachableBackends {
		core.Log.Info("TEST_LOG_REACHABLE", "key", key, "val", val)
	}
	
	// Update the outbound of the dataplane with the generatedVips
	generatedVips := map[string]bool{}
	for _, ob := range meshContext.VIPOutbounds {
		generatedVips[ob.Address] = true
	}
	dpTagSets := dataplane.Spec.SingleValueTagSets()
	var outbounds []*mesh_proto.Dataplane_Networking_Outbound
	core.Log.Info("TEST vIPS", "meshContext.VIPOutbounds", meshContext.VIPOutbounds)
	for _, outbound := range meshContext.VIPOutbounds {
		if outbound.BackendRef == nil { // reachable services does not work with backend ref yet.
			if p.UseMeshService {
				continue
			}
			service := outbound.GetService()
			if len(reachableServices) != 0 {
				if !reachableServices[service] {
					// ignore VIP outbound if reachableServices is defined and not specified
					// Reachable services takes precedence over reachable services graph.
					continue
				}
			} else {
				// static reachable services takes precedence over the graph
				if !xds_context.CanReachFromAny(meshContext.ReachableServicesGraph, dpTagSets, outbound.Tags) {
					continue
				}
			}
			if dataplane.UsesInboundInterface(net.ParseIP(outbound.Address), outbound.Port) {
				// Skip overlapping outbound interface with inbound.
				// This may happen for example with Headless service on Kubernetes (outbound is a PodIP not ClusterIP, so it's the same as inbound).
				continue
			}
		} else {
			/// add port check
			if len(reachableBackends) != 0 {
				backendKey := BackendKey{
					Kind: outbound.BackendRef.Kind,
					Name: outbound.BackendRef.Name,
					Port: outbound.BackendRef.Port,
				}
				core.Log.Info("TEST FIND something", "backendKey", backendKey, "outbound", outbound)
				core.Log.Info("TEST FIND", "!reachableBackends[backendKey]", !reachableBackends[backendKey], "!reachableBackends[BackendKey{Kind: outbound.BackendRef.Kind, Name: outbound.BackendRef.Name}", !reachableBackends[BackendKey{Kind: outbound.BackendRef.Kind, Name: outbound.BackendRef.Name}])
				// check if there is an entry with specific port or without port
				if !reachableBackends[backendKey] && !reachableBackends[BackendKey{Kind: outbound.BackendRef.Kind, Name: outbound.BackendRef.Name}] {
					// ignore VIP outbound if reachableServices is defined and not specified
					// Reachable services takes precedence over reachable services graph.
					continue
				}
			} else if outbound.BackendRef.Kind != "MeshExternalService" {
				// static reachable services takes precedence over the graph
				if !xds_context.CanReachBackendFromAny(meshContext.ReachableServicesGraph, dpTagSets, outbound.BackendRef) {
					continue
				}
			}
			if dataplane.UsesInboundInterface(net.ParseIP(outbound.Address), outbound.Port) {
				// Skip overlapping outbound interface with inbound.
				// This may happen for example with Headless service on Kubernetes (outbound is a PodIP not ClusterIP, so it's the same as inbound).
				continue
			}
		}
		outbounds = append(outbounds, outbound)
	}
	core.Log.Info("TEST OUTBOUND", "outbounds", outbounds, "generatedVips", generatedVips)
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		if generatedVips[outbound.Address] { // Useful while we still have resources with computed vip outbounds
			continue
		}
		outbounds = append(outbounds, outbound)
	}
	core.Log.Info("TEST OUTBOUND", "outbounds", outbounds)
	dataplane.Spec.Networking.Outbound = outbounds
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

type BackendKey struct {
	Kind string
	Name string
	Port uint32
}

type ReachableBackends map[BackendKey]bool

func GetReachableBackends(meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource) ReachableBackends {
	reachableBackends := ReachableBackends{}
	for _, reachableBackend := range dataplane.Spec.Networking.TransparentProxying.ReachableBackendRefs {
		key := BackendKey{Kind: reachableBackend.Kind}
		name := ""
		if reachableBackend.Name != "" {
			name = reachableBackend.Name
		}
		if reachableBackend.Namespace != "" {
			name += fmt.Sprintf(".%s", reachableBackend.Namespace)
		}
		key.Name = name
		if reachableBackend.Port != nil {
			key.Port = reachableBackend.Port.GetValue()
		}
		resourcesLabels := meshContext.MeshServiceNamesByLabels
		if reachableBackend.Kind == "MeshExternalService" {
			resourcesLabels = meshContext.MeshExternalServiceNamesByLabels
		}
		core.Log.Info("TEST_LOG_REACHABLE 2", "key", key)
		if len(reachableBackend.Labels) > 0 {
			reachable := GetResourceNamesForLabels(resourcesLabels, reachableBackend.Labels)
			for name, count := range reachable {
				if count == len(reachableBackend.Labels) {
					reachableBackends[BackendKey{
						Kind: reachableBackend.Kind,
						Name: name,
					}] = true
				}
			}
		}
		if name != "" {
			reachableBackends[key] = true
		}
	}
	return reachableBackends
}

func GetResourceNamesForLabels(resourcesLabels map[string]map[string][]string, labels map[string]string) map[string]int {
	reachable := map[string]int{}
	for key, value := range labels {
		if _, ok := resourcesLabels[key]; ok {
			if _, ok := resourcesLabels[key][value]; ok {
				for _, name := range resourcesLabels[key][value] {
					reachable[name]++
				}
			}
		}
	}
	core.Log.Info("GetResourceNamesForLabels", "resourcesLabels", resourcesLabels, "labels", labels)
	return reachable
}
