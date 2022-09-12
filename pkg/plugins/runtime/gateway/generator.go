package gateway

import (
	"context"
	"fmt"
	"sort"
	"strings"

	envoy_service_runtime_v3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/merge"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

// RoutePolicyTypes specifies the resource types the gateway will bind
// for routes.
var RoutePolicyTypes = []model.ResourceType{
	core_mesh.MeshGatewayRouteType,
}

// ConnectionPolicyTypes specifies the resource types the gateway will
// bind for connection policies.
var ConnectionPolicyTypes = []model.ResourceType{
	core_mesh.CircuitBreakerType,
	core_mesh.FaultInjectionType,
	core_mesh.HealthCheckType,
	core_mesh.RateLimitType,
	core_mesh.RetryType,
	core_mesh.TimeoutType,
}

type GatewayHostInfo struct {
	Host    GatewayHost
	Entries []route.Entry
}

type GatewayHost struct {
	Hostname string
	Routes   []*core_mesh.MeshGatewayRouteResource
	Policies map[model.ResourceType][]match.RankedPolicy
	TLS      *mesh_proto.MeshGateway_TLS_Conf
	// Contains MeshGateway, Listener and Dataplane object tags
	Tags mesh_proto.TagSelector
}

type GatewayListener struct {
	Port         uint32
	Protocol     mesh_proto.MeshGateway_Listener_Protocol
	ResourceName string
	// CrossMesh is important because for generation we need to treat such a
	// listener as if we have HTTPS with the Mesh cert for this Dataplane
	CrossMesh bool
	Resources *mesh_proto.MeshGateway_Listener_Resources // TODO verify these don't conflict when merging
}

// GatewayListenerInfo holds everything needed to generate resources for a
// listener.
type GatewayListenerInfo struct {
	Proxy             *core_xds.Proxy
	Gateway           *core_mesh.MeshGatewayResource
	ExternalServices  *core_mesh.ExternalServiceResourceList
	OutboundEndpoints core_xds.EndpointMap

	Listener  GatewayListener
	HostInfos []GatewayHostInfo
}

// FilterChainGenerator is responsible for handling the filter chain for
// a specific protocol.
// A FilterChainGenerator can be host-specific or shared amongst hosts.
type FilterChainGenerator interface {
	Generate(xds_context.Context, GatewayListenerInfo, []GatewayHost) (*core_xds.ResourceSet, []*envoy_listeners.FilterChainBuilder, error)
}

// Generator generates xDS resources for an entire Gateway.
type Generator struct {
	FilterChainGenerators filterChainGenerators
	ClusterGenerator      ClusterGenerator
	Zone                  string
}

type filterChainGenerators struct {
	FilterChainGenerators map[mesh_proto.MeshGateway_Listener_Protocol]FilterChainGenerator
}

func (g *filterChainGenerators) For(ctx xds_context.Context, info GatewayListenerInfo) FilterChainGenerator {
	gen := g.FilterChainGenerators[info.Listener.Protocol]
	return gen
}

// GatewayListenerInfoFromProxy processes a Dataplane and the corresponding
// Gateway and returns information about the listeners, routes and applied
// policies.
func GatewayListenerInfoFromProxy(
	ctx context.Context, meshCtx xds_context.MeshContext, proxy *core_xds.Proxy, zone string,
) (
	[]GatewayListenerInfo, error,
) {
	gateway := xds_topology.SelectGateway(meshCtx.Resources.Gateways().Items, proxy.Dataplane.Spec.Matches)

	if gateway == nil {
		log.V(1).Info("no matching gateway for dataplane",
			"name", proxy.Dataplane.Meta.GetName(),
			"mesh", proxy.Dataplane.Meta.GetMesh(),
			"service", proxy.Dataplane.Spec.GetIdentifyingService(),
		)

		return nil, nil
	}

	log.V(1).Info(fmt.Sprintf("matched gateway %q to dataplane %q",
		gateway.Meta.GetName(), proxy.Dataplane.Meta.GetName()))

	// Canonicalize the tags on each listener to be the merged resources
	// of dataplane, gateway and listener tags.
	for _, listener := range gateway.Spec.GetConf().GetListeners() {
		listener.Tags = mesh_proto.Merge(
			proxy.Dataplane.Spec.GetNetworking().GetGateway().GetTags(),
			gateway.Spec.GetTags(),
			listener.GetTags(),
		)
	}

	// Multiple listener specifications can have the same port. If
	// they are compatible, then we can collapse those specifications
	// down to a single listener.
	collapsed := map[uint32][]*mesh_proto.MeshGateway_Listener{}
	for _, ep := range gateway.Spec.GetConf().GetListeners() {
		collapsed[ep.GetPort()] = append(collapsed[ep.GetPort()], ep)
	}

	externalServices := meshCtx.Resources.ExternalServices()

	var listenerInfos []GatewayListenerInfo

	matchedExternalServices, err := permissions.MatchExternalServicesTrafficPermissions(
		proxy.Dataplane, meshCtx.Resources.ExternalServices(), meshCtx.Resources.TrafficPermissions(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find external services matched by traffic permissions")
	}

	outboundEndpoints := core_xds.EndpointMap{}
	for k, v := range meshCtx.EndpointMap {
		outboundEndpoints[k] = v
	}

	esEndpoints := xds_topology.BuildExternalServicesEndpointMap(
		ctx,
		meshCtx.Resource,
		matchedExternalServices,
		meshCtx.DataSourceLoader,
		zone,
	)
	for k, v := range esEndpoints {
		outboundEndpoints[k] = v
	}

	// We already validate that listeners are collapsible
	for _, listeners := range collapsed {
		listener, hosts := MakeGatewayListener(meshCtx, gateway, proxy.Dataplane, listeners)

		var hostInfos []GatewayHostInfo
		for _, host := range hosts {
			log.V(1).Info("applying merged traffic routes",
				"listener-port", listener.Port,
				"listener-name", listener.ResourceName,
			)

			hostInfos = append(hostInfos, GatewayHostInfo{
				Host:    host,
				Entries: GenerateEnvoyRouteEntries(host),
			})
		}

		listenerInfos = append(listenerInfos, GatewayListenerInfo{
			Proxy:             proxy,
			Gateway:           gateway,
			ExternalServices:  externalServices,
			OutboundEndpoints: outboundEndpoints,
			Listener:          listener,
			HostInfos:         hostInfos,
		})
	}

	return listenerInfos, nil
}

func (g Generator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	listenerInfos, err := GatewayListenerInfoFromProxy(context.TODO(), ctx.Mesh, proxy, g.Zone)
	if err != nil {
		return nil, errors.Wrap(err, "error generating listener info from Proxy")
	}

	var limits []RuntimeResoureLimitListener

	for _, info := range listenerInfos {
		// This is checked by the gateway validator
		if !SupportsProtocol(info.Listener.Protocol) {
			return nil, errors.New("no support for protocol")
		}

		cdsResources, err := g.generateCDS(ctx, info, info.HostInfos)
		if err != nil {
			return nil, err
		}
		resources.AddSet(cdsResources)

		ldsResources, limit, err := g.generateLDS(ctx, info, info.HostInfos)
		if err != nil {
			return nil, err
		}
		resources.AddSet(ldsResources)

		if limit != nil {
			limits = append(limits, *limit)
		}

		rdsResources, err := g.generateRDS(ctx, info, info.HostInfos)
		if err != nil {
			return nil, err
		}
		resources.AddSet(rdsResources)
	}

	resources.Add(g.generateRTDS(limits))

	return resources, nil
}

func (g Generator) generateRTDS(limits []RuntimeResoureLimitListener) *core_xds.Resource {
	layer := map[string]interface{}{}
	for _, limit := range limits {
		layer[fmt.Sprintf("envoy.resource_limits.listener.%s.connection_limit", limit.Name)] = limit.ConnectionLimit
	}

	res := &core_xds.Resource{
		Name:   "gateway.listeners",
		Origin: OriginGateway,
		Resource: &envoy_service_runtime_v3.Runtime{
			Name:  "gateway.listeners",
			Layer: util_proto.MustStruct(layer),
		},
	}

	return res
}

func (g Generator) generateLDS(ctx xds_context.Context, info GatewayListenerInfo, hostInfos []GatewayHostInfo) (*core_xds.ResourceSet, *RuntimeResoureLimitListener, error) {
	resources := core_xds.NewResourceSet()

	listenerBuilder, limit := GenerateListener(info)

	var gatewayHosts []GatewayHost
	for _, hostInfo := range hostInfos {
		gatewayHosts = append(gatewayHosts, hostInfo.Host)
	}

	protocol := info.Listener.Protocol
	if info.Listener.CrossMesh {
		protocol = mesh_proto.MeshGateway_Listener_HTTPS
	}
	res, filterChainBuilders, err := g.FilterChainGenerators.FilterChainGenerators[protocol].Generate(ctx, info, gatewayHosts)
	if err != nil {
		return nil, limit, err
	}
	resources.AddSet(res)

	for _, filterChainBuilder := range filterChainBuilders {
		listenerBuilder.Configure(envoy_listeners.FilterChain(filterChainBuilder))
	}

	res, err = BuildResourceSet(listenerBuilder)
	if err != nil {
		return nil, limit, errors.Wrapf(err, "failed to build listener resource")
	}
	resources.AddSet(res)

	return resources, limit, nil
}

func (g Generator) generateCDS(ctx xds_context.Context, info GatewayListenerInfo, hostInfos []GatewayHostInfo) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	for _, hostInfo := range hostInfos {
		clusterRes, err := g.ClusterGenerator.GenerateClusters(context.TODO(), ctx, info, hostInfo)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate clusters for dataplane %q", info.Proxy.Id)
		}
		resources.AddSet(clusterRes)
	}

	return resources, nil
}

func (g Generator) generateRDS(ctx xds_context.Context, info GatewayListenerInfo, hostInfos []GatewayHostInfo) (*core_xds.ResourceSet, error) {
	switch info.Listener.Protocol {
	case mesh_proto.MeshGateway_Listener_HTTPS,
		mesh_proto.MeshGateway_Listener_HTTP:
	default:
		return nil, nil
	}

	resources := core_xds.NewResourceSet()
	routeConfig := GenerateRouteConfig(info)

	// Make a pass over the generators for each virtual host.
	for _, hostInfo := range hostInfos {
		vh, err := GenerateVirtualHost(ctx, info, hostInfo.Host, hostInfo.Entries)
		if err != nil {
			return nil, err
		}

		routeConfig.Configure(envoy_routes.VirtualHost(vh))
	}

	res, err := BuildResourceSet(routeConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build route configuration resource")
	}
	resources.AddSet(res)

	return resources, nil
}

// MakeGatewayListener converts a collapsed set of listener configurations
// in to a single configuration with a matched set of route resources. The
// given listeners must have a consistent protocol and port.
// Listeners must be validated for collapsibility in terms of hostnames and
// protocols.
func MakeGatewayListener(
	meshContext xds_context.MeshContext,
	gateway *core_mesh.MeshGatewayResource,
	dataplane *core_mesh.DataplaneResource,
	listeners []*mesh_proto.MeshGateway_Listener,
) (GatewayListener, []GatewayHost) {
	hostsByName := map[string]GatewayHost{}

	listener := GatewayListener{
		Port:     listeners[0].GetPort(),
		Protocol: listeners[0].GetProtocol(),
		ResourceName: envoy_names.GetGatewayListenerName(
			gateway.Meta.GetName(),
			listeners[0].GetProtocol().String(),
			listeners[0].GetPort(),
		),
		CrossMesh: listeners[0].CrossMesh,
		Resources: listeners[0].GetResources(),
	}

	// Hostnames must be unique to a listener to remove ambiguity
	// in policy selection and TLS configuration.
	for _, l := range listeners {
		// An empty hostname is the same as "*", i.e. matches all hosts.
		hostname := l.GetHostname()
		if hostname == "" {
			hostname = mesh_proto.WildcardHostname
		}

		allRoutes := match.Routes(meshContext.Resources.GatewayRoutes().Items, l.GetTags())

		var routes []*core_mesh.MeshGatewayRouteResource
		switch listener.Protocol {
		case mesh_proto.MeshGateway_Listener_HTTP,
			mesh_proto.MeshGateway_Listener_HTTPS,
			mesh_proto.MeshGateway_Listener_TCP:

			routes = route.FilterProtocols(allRoutes, listener.Protocol)
		default:
		}

		host := GatewayHost{
			Hostname: hostname,
			Policies: map[model.ResourceType][]match.RankedPolicy{},
			TLS:      l.GetTls(),
			Tags:     l.Tags,
			Routes:   routes,
		}

		for _, t := range ConnectionPolicyTypes {
			matches := match.ConnectionPoliciesBySource(
				l.GetTags(),
				match.ToConnectionPolicies(meshContext.Resources.MeshLocalResources[t]))
			host.Policies[t] = matches
		}

		hostsByName[hostname] = host
	}

	var hosts []GatewayHost
	for _, host := range hostsByName {
		hosts = append(hosts, host)
	}

	hosts = RedistributeWildcardRoutes(hosts)

	// Sort by reverse hostname, so that fully qualified hostnames sort
	// before wildcard domains, and "*" is last.
	sort.Slice(hosts, func(i, j int) bool {
		return hosts[i].Hostname > hosts[j].Hostname
	})

	return listener, hosts
}

// RedistributeWildcardRoutes takes the routes from the wildcard host
// and redistributes them to hosts with matching names, creating new
// hosts if necessary.
//
// This process is necessary because:
//
//  1. We might have a listener with hostname A and some routes, but also
//     a wildcard listener with routes for hostname A. We want all the routes
//     for hostname A in the same virtual host.
//  2. Routes with hostnames that are attached to a wildcard listener
//     should implicitly create virtual hosts so that we can generate a
//     consistent config. For example, if a wildcard listener has a route for
//     hostname A and a route for hostname B, that doesn't mean that the routes
//     are for hostnames A or B. We still want the routes to match the hostname
//     that they were specified with.
func RedistributeWildcardRoutes(
	hosts []GatewayHost,
) []GatewayHost {
	hostsByName := map[string]GatewayHost{}

	for _, h := range hosts {
		hostsByName[h.Hostname] = h
	}

	wild, ok := hostsByName[mesh_proto.WildcardHostname]
	if !ok {
		return hosts
	}

	wildcardRoutes := wild.Routes
	wild.Routes = nil // We are rebuilding this.
	for _, gw := range wildcardRoutes {
		names := gw.Spec.GetConf().GetHttp().GetHostnames()

		// No hostnames on this route, it stays as a wildcard route.
		if len(names) == 0 {
			wild.Routes = append(wild.Routes, gw)
			continue
		}

		appendRoutesToHost := func(host GatewayHost) {
			// Note that if we already have a virtualhost for this
			// name, and add the route to it, it might be a duplicate.
			host.Routes = append(host.Routes, gw)
			hostsByName[host.Hostname] = host
		}

		for _, n := range names {
			host, ok := hostsByName[n]

			// When generating a new implicit virtualhost,
			// initialize it by shallow copying from the
			// wildcard source.
			if !ok {
				host = wild
				host.Routes = nil
				host.Hostname = n
			}

			appendRoutesToHost(host)

			if suffix := strings.TrimPrefix(n, "*."); len(suffix) != len(n) {
				// An alternative to this might be precalculating
				// a table of suffixes to hosts
				for hostName, host := range hostsByName {
					if !strings.HasSuffix(hostName, suffix) {
						continue
					}
					appendRoutesToHost(host)
				}
			}
		}
	}

	hostsByName[mesh_proto.WildcardHostname] = wild

	var flattened []GatewayHost
	for _, host := range hostsByName {
		flattened = append(flattened, host)
	}

	// return a set of routes for each host
	for i, host := range flattened {
		flattened[i].Routes = merge.UniqueResources(host.Routes)
	}

	return flattened
}
