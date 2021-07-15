package sync

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/ratelimits"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var syncLog = core.Log.WithName("sync")

type DataplaneProxyBuilder struct {
	CachingResManager     manager.ReadOnlyResourceManager
	NonCachingResManager  manager.ReadOnlyResourceManager
	LookupIP              lookup.LookupIPFunc
	DataSourceLoader      datasource.Loader
	MetadataTracker       DataplaneMetadataTracker
	PermissionMatcher     permissions.TrafficPermissionsMatcher
	LogsMatcher           logs.TrafficLogsMatcher
	FaultInjectionMatcher faultinjections.FaultInjectionMatcher
	RateLimitMatcher      ratelimits.RateLimitMatcher

	Zone       string
	apiVersion envoy.APIVersion
}

func (p *DataplaneProxyBuilder) build(key core_model.ResourceKey, meshCtx *xds_context.MeshContext) (*xds.Proxy, error) {
	ctx := context.Background()

	dp, err := p.resolveDataplane(ctx, key)
	if err != nil {
		return nil, err
	}

	routing, destinations, err := p.resolveRouting(ctx, meshCtx, dp)
	if err != nil {
		return nil, err
	}

	matchedPolicies, err := p.matchPolicies(ctx, meshCtx, dp, destinations)
	if err != nil {
		return nil, err
	}

	proxy := &xds.Proxy{
		Id:         xds.FromResourceKey(key),
		APIVersion: p.apiVersion,
		Dataplane:  dp,
		Metadata:   p.MetadataTracker.Metadata(key),
		Routing:    *routing,
		Policies:   *matchedPolicies,
	}
	return proxy, nil
}

func (p *DataplaneProxyBuilder) resolveDataplane(ctx context.Context, key core_model.ResourceKey) (*core_mesh.DataplaneResource, error) {
	dataplane := core_mesh.NewDataplaneResource()

	// we use non-cached ResourceManager to always fetch fresh version of the Dataplane.
	// Otherwise, technically MeshCache can use newer version because it uses List operation instead of Get
	if err := p.NonCachingResManager.Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
		return nil, err
	}

	// Envoy requires IPs instead of Hostname therefore we need to resolve an address. Consider moving this outside of this component.
	resolvedDp, err := xds_topology.ResolveAddress(p.LookupIP, dataplane)
	if err != nil {
		return nil, err
	}
	return resolvedDp, nil
}

func (p *DataplaneProxyBuilder) resolveRouting(
	ctx context.Context,
	meshContext *xds_context.MeshContext,
	dataplane *core_mesh.DataplaneResource,
) (*xds.Routing, xds.DestinationMap, error) {
	externalServices := &core_mesh.ExternalServiceResourceList{}
	if err := p.CachingResManager.List(ctx, externalServices, core_store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return nil, nil, err
	}

	zoneIngresses, err := xds_topology.GetZoneIngresses(syncLog, ctx, p.CachingResManager, p.LookupIP)
	if err != nil {
		return nil, nil, err
	}

	matchedExternalServices, err := p.PermissionMatcher.MatchExternalServices(ctx, dataplane, externalServices)
	if err != nil {
		return nil, nil, err
	}

	// pick a single the most specific route for each outbound interface
	routes, err := xds_topology.GetRoutes(ctx, dataplane, p.CachingResManager)
	if err != nil {
		return nil, nil, err
	}

	// create creates a map of selectors to match other dataplanes reachable via given routes
	destinations := xds_topology.BuildDestinationMap(dataplane, routes)

	// resolve all endpoints that match given selectors
	outbound := xds_topology.BuildEndpointMap(meshContext.Resource, p.Zone, meshContext.Dataplanes.Items, zoneIngresses.Items, matchedExternalServices, p.DataSourceLoader)

	routing := &xds.Routing{
		TrafficRoutes:   routes,
		OutboundTargets: outbound,
	}
	return routing, destinations, nil
}

func (p *DataplaneProxyBuilder) matchPolicies(ctx context.Context, meshContext *xds_context.MeshContext, dataplane *core_mesh.DataplaneResource, outboundSelectors xds.DestinationMap) (*xds.MatchedPolicies, error) {
	healthChecks, err := xds_topology.GetHealthChecks(ctx, dataplane, outboundSelectors, p.CachingResManager)
	if err != nil {
		return nil, err
	}

	circuitBreakers, err := xds_topology.GetCircuitBreakers(ctx, dataplane, outboundSelectors, p.CachingResManager)
	if err != nil {
		return nil, err
	}

	trafficTrace, err := xds_topology.GetTrafficTrace(ctx, dataplane, p.CachingResManager)
	if err != nil {
		return nil, err
	}
	var tracingBackend *mesh_proto.TracingBackend
	if trafficTrace != nil {
		tracingBackend = meshContext.Resource.GetTracingBackend(trafficTrace.Spec.GetConf().GetBackend())
	}

	retries, err := xds_topology.GetRetries(ctx, dataplane, outboundSelectors, p.CachingResManager)
	if err != nil {
		return nil, err
	}

	matchedPermissions, err := p.PermissionMatcher.Match(ctx, dataplane, meshContext.Resource)
	if err != nil {
		return nil, err
	}

	matchedLogs, err := p.LogsMatcher.Match(ctx, dataplane)
	if err != nil {
		return nil, err
	}

	faultInjection, err := p.FaultInjectionMatcher.Match(ctx, dataplane, meshContext.Resource)
	if err != nil {
		return nil, err
	}

	timeouts, err := xds_topology.GetTimeouts(ctx, dataplane, p.CachingResManager)
	if err != nil {
		return nil, err
	}

	ratelimits, err := p.RateLimitMatcher.Match(ctx, dataplane, meshContext.Resource)
	if err != nil {
		return nil, err
	}

	matchedPolicies := &xds.MatchedPolicies{
		TrafficPermissions: matchedPermissions,
		Logs:               matchedLogs,
		HealthChecks:       healthChecks,
		CircuitBreakers:    circuitBreakers,
		TrafficTrace:       trafficTrace,
		TracingBackend:     tracingBackend,
		FaultInjections:    faultInjection,
		Retries:            retries,
		Timeouts:           timeouts,
		RateLimits:         ratelimits,
	}
	return matchedPolicies, nil
}
