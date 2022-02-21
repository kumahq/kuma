package sync

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
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
	"github.com/kumahq/kuma/pkg/xds/template"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var syncLog = core.Log.WithName("sync")

type DataplaneProxyBuilder struct {
	DataSourceLoader datasource.Loader
	MetadataTracker  DataplaneMetadataTracker

	Zone       string
	APIVersion envoy.APIVersion
}

func (p *DataplaneProxyBuilder) Build(key core_model.ResourceKey, meshContext xds_context.MeshContext) (*xds.Proxy, error) {
	dp, found := meshContext.DataplanesByName[key.Name]
	if !found {
		return nil, core_store.ErrorResourceNotFound(core_mesh.DataplaneType, key.Name, key.Mesh)
	}

	routing, destinations, err := p.resolveRouting(meshContext, dp)
	if err != nil {
		return nil, err
	}

	matchedPolicies, err := p.matchPolicies(meshContext, dp, destinations)
	if err != nil {
		return nil, err
	}

	matchedPolicies.TrafficRoutes = routing.TrafficRoutes

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
	matchedExternalServices, err := permissions.MatchExternalServices(dataplane, meshContext.Resources.ExternalServices(), meshContext.Resources.TrafficPermissions())
	if err != nil {
		return nil, nil, err
	}

	p.resolveVIPOutbounds(meshContext, dataplane)

	// pick a single the most specific route for each outbound interface
	routes := xds_topology.BuildRouteMap(dataplane, meshContext.Resources.TrafficRoutes().Items)

	// create a map of selectors to match other dataplanes reachable via given routes
	destinations := xds_topology.BuildDestinationMap(dataplane, routes)

	// resolve all endpoints that match given selectors
	outbound := xds_topology.BuildEndpointMap(
		meshContext.Resource,
		p.Zone,
		meshContext.Resources.Dataplanes().Items,
		meshContext.Resources.ZoneIngresses().Items,
		meshContext.Resources.ZoneEgresses().Items,
		matchedExternalServices, p.DataSourceLoader,
	)

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
	reachableServices := map[string]bool{}
	for _, reachableService := range dataplane.Spec.Networking.TransparentProxying.ReachableServices {
		reachableServices[reachableService] = true
	}

	// Update the outbound of the dataplane with the generatedVips
	generatedVips := map[string]bool{}
	for _, ob := range meshContext.VIPOutbounds {
		generatedVips[ob.Address] = true
	}
	var outbounds []*mesh_proto.Dataplane_Networking_Outbound
	for _, outbound := range meshContext.VIPOutbounds {
		service := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]
		if len(reachableServices) == 0 || reachableServices[service] { // ignore VIP outbound if reachableServices is defined and not specified
			outbounds = append(outbounds, outbound)
		}
	}
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

	resources := meshContext.Resources
	ratelimits := ratelimits.BuildRateLimitMap(dataplane, inbounds, resources.RateLimits().Items)
	matchedPolicies := &xds.MatchedPolicies{
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
	}
	return matchedPolicies, nil
}
