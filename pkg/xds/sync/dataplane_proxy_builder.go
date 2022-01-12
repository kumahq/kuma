package sync

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/insights"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var syncLog = core.Log.WithName("sync")

type DataplaneProxyBuilder struct {
	LookupIP         lookup.LookupIPFunc
	DataSourceLoader datasource.Loader
	MetadataTracker  DataplaneMetadataTracker

	Zone       string
	APIVersion envoy.APIVersion
}

func (p *DataplaneProxyBuilder) Build(key core_model.ResourceKey, envoyContext *xds_context.Context) (*xds.Proxy, error) {
	ctx := context.Background()

	dp, err := p.resolveDataplane(ctx, key, envoyContext.Mesh.Snapshot)
	if err != nil {
		return nil, err
	}

	routing, destinations, err := p.resolveRouting(ctx, envoyContext.Mesh, dp)
	if err != nil {
		return nil, err
	}

	matchedPolicies, err := p.matchPolicies(ctx, envoyContext.Mesh, dp, destinations)
	if err != nil {
		return nil, err
	}

	tlsReady, err := p.resolveTLSReadiness(ctx, key, envoyContext)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't determine TLS readiness of services")
	}

	proxy := &xds.Proxy{
		Id:                  xds.FromResourceKey(key),
		APIVersion:          p.APIVersion,
		Dataplane:           dp,
		Metadata:            p.MetadataTracker.Metadata(key),
		Routing:             *routing,
		Policies:            *matchedPolicies,
		ServiceTLSReadiness: tlsReady,
	}
	return proxy, nil
}

func (p *DataplaneProxyBuilder) resolveDataplane(ctx context.Context, key core_model.ResourceKey, snapshot xds_context.MeshSnapshot) (*core_mesh.DataplaneResource, error) {
	// we use non-cached ResourceManager to always fetch fresh version of the Dataplane.
	// Otherwise, technically MeshCache can use newer version because it uses List operation instead of Get
	resource, found := snapshot.Resource(core_mesh.DataplaneType, key)
	if !found {
		return nil, core_store.ErrorResourceNotFound(core_mesh.DataplaneType, key.Name, key.Mesh)
	}
	dataplane := resource.(*core_mesh.DataplaneResource)

	// todo move resolve address to mesh snapshot?
	// Envoy requires IPs instead of Hostname therefore we need to resolve an address. Consider moving this outside of this component.
	resolvedDp, err := xds_topology.ResolveAddress(p.LookupIP, dataplane)
	if err != nil {
		return nil, err
	}
	return resolvedDp, nil
}

func (p *DataplaneProxyBuilder) resolveRouting(
	ctx context.Context,
	meshContext xds_context.MeshContext,
	dataplane *core_mesh.DataplaneResource,
) (*xds.Routing, xds.DestinationMap, error) {
	externalServices := meshContext.Snapshot.Resources(core_mesh.ExternalServiceType).(*core_mesh.ExternalServiceResourceList)
	trafficPermissions := meshContext.Snapshot.Resources(core_mesh.TrafficPermissionType).(*core_mesh.TrafficPermissionResourceList)

	zoneIngresses := meshContext.Snapshot.Resources(core_mesh.ZoneIngressType).(*core_mesh.ZoneIngressResourceList)
	zoneIngresses.Items = xds_topology.ResolveZoneIngressAddresses(syncLog, p.LookupIP, zoneIngresses.Items) // todo resolve in mesh snapshot?

	matchedExternalServices, err := permissions.MatchExternalServices(dataplane, externalServices, trafficPermissions)
	if err != nil {
		return nil, nil, err
	}

	if dataplane.Spec.Networking.GetTransparentProxying() != nil {
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

	// pick a single the most specific route for each outbound interface
	trafficRoutes := meshContext.Snapshot.Resources(core_mesh.TrafficRouteType).(*core_mesh.TrafficRouteResourceList)
	routes := xds_topology.BuildRouteMap(dataplane, trafficRoutes.Items)

	// create creates a map of selectors to match other dataplanes reachable via given routes
	destinations := xds_topology.BuildDestinationMap(dataplane, routes)

	// resolve all endpoints that match given selectors
	outbound := xds_topology.BuildEndpointMap(meshContext.Resource, p.Zone, meshContext.Dataplanes.Items, zoneIngresses.Items, matchedExternalServices, p.DataSourceLoader)

	routing := &xds.Routing{
		TrafficRoutes:   routes,
		OutboundTargets: outbound,
		VipDomains:      meshContext.VIPDomains, // todo this can be removed since it's available in MeshContext#VIPs domains
	}
	return routing, destinations, nil
}

func (p *DataplaneProxyBuilder) matchPolicies(ctx context.Context, meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource, outboundSelectors xds.DestinationMap) (*xds.MatchedPolicies, error) {
	// todo should those be moved to functions in MeshSnapshot?
	healthChecks := meshContext.Snapshot.Resources(core_mesh.HealthCheckType).(*core_mesh.HealthCheckResourceList).Items
	circuitBreakers := meshContext.Snapshot.Resources(core_mesh.CircuitBreakerType).(*core_mesh.CircuitBreakerResourceList).Items
	trafficTraces := meshContext.Snapshot.Resources(core_mesh.TrafficTraceType).(*core_mesh.TrafficTraceResourceList).Items
	retries := meshContext.Snapshot.Resources(core_mesh.RetryType).(*core_mesh.RetryResourceList).Items
	trafficPermissions := meshContext.Snapshot.Resources(core_mesh.TrafficPermissionType).(*core_mesh.TrafficPermissionResourceList).Items
	trafficLogs := meshContext.Snapshot.Resources(core_mesh.TrafficLogType).(*core_mesh.TrafficLogResourceList).Items
	faultInjections := meshContext.Snapshot.Resources(core_mesh.FaultInjectionType).(*core_mesh.FaultInjectionResourceList).Items
	timeouts := meshContext.Snapshot.Resources(core_mesh.TimeoutType).(*core_mesh.TimeoutResourceList).Items
	rateLimits := meshContext.Snapshot.Resources(core_mesh.RateLimitType).(*core_mesh.RateLimitResourceList).Items

	matchedPermissions, err := permissions.BuildTrafficPermissionMap(dataplane, meshContext.Snapshot.Mesh(), trafficPermissions)
	if err != nil {
		return nil, err
	}

	faultInjection, err := faultinjections.BuildFaultInjectionMap(dataplane, meshContext.Snapshot.Mesh(), faultInjections)
	if err != nil {
		return nil, err
	}

	ratelimits, err := ratelimits.BuildRateLimitMap(dataplane, meshContext.Snapshot.Mesh(), rateLimits)
	if err != nil {
		return nil, err
	}

	matchedPolicies := &xds.MatchedPolicies{
		TrafficPermissions: matchedPermissions,
		TrafficLogs:        logs.BuildTrafficLogMap(dataplane, trafficLogs),
		HealthChecks:       xds_topology.BuildHealthCheckMap(dataplane, outboundSelectors, healthChecks),
		CircuitBreakers:    xds_topology.BuildCircuitBreakerMap(dataplane, outboundSelectors, circuitBreakers),
		TrafficTrace:       xds_topology.SelectTrafficTrace(dataplane, trafficTraces),
		FaultInjections:    faultInjection,
		Retries:            xds_topology.BuildRetryMap(dataplane, retries, outboundSelectors),
		Timeouts:           xds_topology.BuildTimeoutMap(dataplane, timeouts),
		RateLimitsInbound:  ratelimits.Inbound,
		RateLimitsOutbound: ratelimits.Outbound,
	}
	return matchedPolicies, nil
}

func (p *DataplaneProxyBuilder) resolveTLSReadiness(
	ctx context.Context, key core_model.ResourceKey, envoyContext *xds_context.Context,
) (map[string]bool, error) {
	tlsReady := map[string]bool{}

	backend := envoyContext.Mesh.Resource.GetEnabledCertificateAuthorityBackend()
	// TLS readiness is irrelevant unless we are using PERMISSIVE TLS, so skip
	// checking ServiceInsights if we aren't.
	if backend == nil || backend.Mode != mesh_proto.CertificateAuthorityBackend_PERMISSIVE {
		return tlsReady, nil
	}

	resource, found := envoyContext.Mesh.Snapshot.Resource(core_mesh.ServiceInsightType, insights.ServiceInsightKey(key.Mesh))
	if !found {
		// Nothing about the TLS readiness has been reported yet
		syncLog.Info("could not determine service TLS readiness, ServiceInsight is not yet present")
		return tlsReady, nil
	}
	serviceInsight := resource.(*core_mesh.ServiceInsightResource)
	for svc, insight := range serviceInsight.Spec.GetServices() {
		tlsReady[svc] = insight.IssuedBackends[backend.Name] == insight.Dataplanes.Total
	}
	return tlsReady, nil
}
