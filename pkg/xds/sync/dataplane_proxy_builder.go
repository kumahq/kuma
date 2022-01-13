package sync

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	manager_dataplane "github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	"github.com/kumahq/kuma/pkg/core/permissions"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var syncLog = core.Log.WithName("sync")

type DataplaneProxyBuilder struct {
	DataSourceLoader datasource.Loader
	MetadataTracker  DataplaneMetadataTracker

	Zone       string
	APIVersion envoy.APIVersion
}

func (p *DataplaneProxyBuilder) Build(key core_model.ResourceKey, envoyContext *xds_context.Context) (*xds.Proxy, error) {
	dp, found := envoyContext.Mesh.DataplanesByName[key.Name]
	if !found {
		return nil, core_store.ErrorResourceNotFound(core_mesh.DataplaneType, key.Name, key.Mesh)
	}

	routing, destinations, err := p.resolveRouting(envoyContext.Mesh, dp)
	if err != nil {
		return nil, err
	}

	matchedPolicies, err := p.matchPolicies(envoyContext.Mesh, dp, destinations)
	if err != nil {
		return nil, err
	}

	proxy := &xds.Proxy{
		Id:         xds.FromResourceKey(key),
		APIVersion: p.APIVersion,
		Dataplane:  dp,
		Metadata:   p.MetadataTracker.Metadata(key),
		Routing:    *routing,
		Policies:   *matchedPolicies,
	}
	return proxy, nil
}

func (p *DataplaneProxyBuilder) resolveRouting(meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource) (*xds.Routing, xds.DestinationMap, error) {
	matchedExternalServices, err := permissions.MatchExternalServices(dataplane, meshContext.ExternalServices(), meshContext.TrafficPermissions())
	if err != nil {
		return nil, nil, err
	}

	p.resolveVIPOutbounds(meshContext, dataplane)

	// pick a single the most specific route for each outbound interface
	routes := xds_topology.BuildRouteMap(dataplane, meshContext.TrafficRoutes().Items)

	// create creates a map of selectors to match other dataplanes reachable via given routes
	destinations := xds_topology.BuildDestinationMap(dataplane, routes)

	// resolve all endpoints that match given selectors
	outbound := xds_topology.BuildEndpointMap(meshContext.Resource, p.Zone, meshContext.Dataplanes.Items, meshContext.ZoneIngresses.Items, matchedExternalServices, p.DataSourceLoader)

	routing := &xds.Routing{
		TrafficRoutes:   routes,
		OutboundTargets: outbound,
	}
	return routing, destinations, nil
}

func (p *DataplaneProxyBuilder) resolveVIPOutbounds(meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource) {
	if dataplane.Spec.Networking.GetTransparentProxying() == nil {
		return
	}
	// Update the outbound of the dataplane with the generatedVips
	generatedVips := map[string]bool{}
	for _, ob := range meshContext.VIPOutbounds {
		generatedVips[ob.Address] = true
	}
	outbounds := meshContext.VIPOutbounds
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		if generatedVips[outbound.Address] { // Useful while we still have resources with computed vip outbounds
			continue
		}
		outbounds = append(outbounds, outbound)
	}
	dataplane.Spec.Networking.Outbound = outbounds
}

func (p *DataplaneProxyBuilder) matchPolicies(meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource, outboundSelectors xds.DestinationMap) (*xds.MatchedPolicies, error) {
	additionalInbounds, err := manager_dataplane.AdditionalInbounds(dataplane, meshContext.Resource)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch additional inbounds")
	}
	inbounds := append(dataplane.Spec.GetNetworking().GetInbound(), additionalInbounds...)

	ratelimits := ratelimits.BuildRateLimitMap(dataplane, inbounds, meshContext.RateLimits().Items)
	matchedPolicies := &xds.MatchedPolicies{
		TrafficPermissions: permissions.BuildTrafficPermissionMap(dataplane, inbounds, meshContext.TrafficPermissions().Items),
		TrafficLogs:        logs.BuildTrafficLogMap(dataplane, meshContext.TrafficLogs().Items),
		HealthChecks:       xds_topology.BuildHealthCheckMap(dataplane, outboundSelectors, meshContext.HealthChecks().Items),
		CircuitBreakers:    xds_topology.BuildCircuitBreakerMap(dataplane, outboundSelectors, meshContext.CircuitBreakers().Items),
		TrafficTrace:       xds_topology.SelectTrafficTrace(dataplane, meshContext.TrafficTraces().Items),
		FaultInjections:    faultinjections.BuildFaultInjectionMap(dataplane, inbounds, meshContext.FaultInjections().Items),
		Retries:            xds_topology.BuildRetryMap(dataplane, meshContext.Retries().Items, outboundSelectors),
		Timeouts:           xds_topology.BuildTimeoutMap(dataplane, meshContext.Timeouts().Items),
		RateLimitsInbound:  ratelimits.Inbound,
		RateLimitsOutbound: ratelimits.Outbound,
	}
	return matchedPolicies, nil
}
