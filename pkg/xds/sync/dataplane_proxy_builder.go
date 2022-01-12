package sync

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
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
	dp, found := envoyContext.Mesh.Dataplane(key.Name)
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

	tlsReady, err := p.resolveTLSReadiness(key, envoyContext.Mesh)
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
	matchedPermissions, err := permissions.BuildTrafficPermissionMap(dataplane, meshContext.Snapshot.Mesh(), meshContext.TrafficPermissions().Items)
	if err != nil {
		return nil, err
	}

	faultInjection, err := faultinjections.BuildFaultInjectionMap(dataplane, meshContext.Snapshot.Mesh(), meshContext.FaultInjections().Items)
	if err != nil {
		return nil, err
	}

	ratelimits, err := ratelimits.BuildRateLimitMap(dataplane, meshContext.Snapshot.Mesh(), meshContext.RateLimits().Items)
	if err != nil {
		return nil, err
	}

	matchedPolicies := &xds.MatchedPolicies{
		TrafficPermissions: matchedPermissions,
		TrafficLogs:        logs.BuildTrafficLogMap(dataplane, meshContext.TrafficLogs().Items),
		HealthChecks:       xds_topology.BuildHealthCheckMap(dataplane, outboundSelectors, meshContext.HealthChecks().Items),
		CircuitBreakers:    xds_topology.BuildCircuitBreakerMap(dataplane, outboundSelectors, meshContext.CircuitBreakers().Items),
		TrafficTrace:       xds_topology.SelectTrafficTrace(dataplane, meshContext.TrafficTraces().Items),
		FaultInjections:    faultInjection,
		Retries:            xds_topology.BuildRetryMap(dataplane, meshContext.Retries().Items, outboundSelectors),
		Timeouts:           xds_topology.BuildTimeoutMap(dataplane, meshContext.Timeouts().Items),
		RateLimitsInbound:  ratelimits.Inbound,
		RateLimitsOutbound: ratelimits.Outbound,
	}
	return matchedPolicies, nil
}

func (p *DataplaneProxyBuilder) resolveTLSReadiness(key core_model.ResourceKey, meshContext xds_context.MeshContext) (map[string]bool, error) {
	tlsReady := map[string]bool{}

	backend := meshContext.Resource.GetEnabledCertificateAuthorityBackend()
	// TLS readiness is irrelevant unless we are using PERMISSIVE TLS, so skip
	// checking ServiceInsights if we aren't.
	if backend == nil || backend.Mode != mesh_proto.CertificateAuthorityBackend_PERMISSIVE {
		return tlsReady, nil
	}

	serviceInsight, found := meshContext.ServiceInsight()
	if !found {
		// Nothing about the TLS readiness has been reported yet
		syncLog.Info("could not determine service TLS readiness, ServiceInsight is not yet present")
		return tlsReady, nil
	}
	for svc, insight := range serviceInsight.Spec.GetServices() {
		tlsReady[svc] = insight.IssuedBackends[backend.Name] == insight.Dataplanes.Total
	}
	return tlsReady, nil
}
