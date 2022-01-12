package sync

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/insights"
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

	Zone           string
	APIVersion     envoy.APIVersion
	ConfigManager  config_manager.ConfigManager
	TopLevelDomain string
}

func (p *DataplaneProxyBuilder) Build(key core_model.ResourceKey, envoyContext *xds_context.Context) (*xds.Proxy, error) {
	ctx := context.Background()

	dp, err := p.resolveDataplane(ctx, key)
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
	meshContext xds_context.MeshContext,
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

	var domains []xds.VIPDomains
	outbounds := dataplane.Spec.Networking.Outbound
	if dataplane.Spec.Networking.GetTransparentProxying() != nil {
		pers := vips.NewPersistence(p.CachingResManager, p.ConfigManager)
		virtualOutboundView, err := pers.GetByMesh(dataplane.Meta.GetMesh())
		if err != nil {
			return nil, nil, err
		}
		// resolve all the domains
		domains, outbounds = xds_topology.VIPOutbounds(virtualOutboundView, p.TopLevelDomain)

		// Update the outbound of the dataplane with the generatedVips
		generatedVips := map[string]bool{}
		for _, ob := range outbounds {
			generatedVips[ob.Address] = true
		}
		for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
			if generatedVips[outbound.Address] { // Useful while we still have resources with computed vip outbounds
				continue
			}
			outbounds = append(outbounds, outbound)
		}
	}
	dataplane.Spec.Networking.Outbound = outbounds

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
		VipDomains:      domains,
	}
	return routing, destinations, nil
}

func (p *DataplaneProxyBuilder) matchPolicies(ctx context.Context, meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource, outboundSelectors xds.DestinationMap) (*xds.MatchedPolicies, error) {
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

	retries, err := xds_topology.GetRetries(ctx, dataplane, outboundSelectors, p.CachingResManager)
	if err != nil {
		return nil, err
	}

	matchedPermissions, err := p.PermissionMatcher.Match(ctx, dataplane, meshContext.Resource)
	if err != nil {
		return nil, err
	}

	trafficLogs, err := p.LogsMatcher.Match(ctx, dataplane)
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
		TrafficLogs:        trafficLogs,
		HealthChecks:       healthChecks,
		CircuitBreakers:    circuitBreakers,
		TrafficTrace:       trafficTrace,
		FaultInjections:    faultInjection,
		Retries:            retries,
		Timeouts:           timeouts,
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

	serviceInsight := core_mesh.NewServiceInsightResource()
	if err := p.CachingResManager.Get(ctx, serviceInsight, core_store.GetBy(insights.ServiceInsightKey(key.Mesh))); err != nil {
		if core_store.IsResourceNotFound(err) {
			// Nothing about the TLS readiness has been reported yet
			syncLog.Info("could not determine service TLS readiness", "error", err)
			return tlsReady, nil
		}
		return nil, err
	}

	for svc, insight := range serviceInsight.Spec.GetServices() {
		tlsReady[svc] = insight.IssuedBackends[backend.Name] == insight.Dataplanes.Total
	}

	return tlsReady, nil
}
