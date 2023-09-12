package globalloadbalancer

import (
	"context"
	"fmt"

	envoy_service_runtime_v3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/globalloadbalancer/metadata"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

const (
	GlobalLoadBalancerRoutesName = "global-load-balancer-routes"
)

// FilterChainGenerator is responsible for handling the filter chain for
// a specific protocol.
// A FilterChainGenerator can be host-specific or shared amongst hosts.
type FilterChainGenerator interface {
	Generate(xdsCtx xds_context.Context, info GatewayListenerInfo) (*core_xds.ResourceSet, []*envoy_listeners.FilterChainBuilder, error)
}

type GatewayListener struct {
	Port         uint32
	Protocol     mesh_proto.MeshGateway_Listener_Protocol
	ResourceName string
	Resources    *mesh_proto.MeshGateway_Listener_Resources // TODO verify these don't conflict when merging
	TLS          *mesh_proto.MeshGateway_TLS_Conf
}

// GatewayListenerInfo holds everything needed to generate resources for a
// listener.
type GatewayListenerInfo struct {
	Proxy   *core_xds.Proxy
	Gateway *core_mesh.MeshGatewayResource

	Listener GatewayListener
}

// Generator generates xDS resources for an entire Global Load Balancer.
type Generator struct {
	Zone                  string
	FilterChainGenerators FilterChainGenerators
	ClusterGenerator      *ClusterGenerator
}

type FilterChainGenerators struct {
	FilterChainGenerators map[mesh_proto.MeshGateway_Listener_Protocol]FilterChainGenerator
}

func (g *FilterChainGenerators) For(ctx xds_context.Context, info GatewayListenerInfo) FilterChainGenerator {
	gen := g.FilterChainGenerators[info.Listener.Protocol]
	return gen
}

type Route struct {
	Mesh            string
	Service         string
	DeploymentGroup string
}

// GatewayListenerInfoFromProxy processes a Dataplane and the corresponding
// Gateway and returns information about the listeners, routes and applied
// policies.
// NOTE(nicoche)  This ^ is the original comment. In practice, we use it
// only for the listener part. The routes/endpoint part is not computed here
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

	var listenerInfos []GatewayListenerInfo
	// We already validate that listeners are collapsible
	for _, listeners := range collapsed {
		listener := MakeGatewayListener(gateway, listeners)

		listenerInfos = append(listenerInfos, GatewayListenerInfo{
			Proxy:    proxy,
			Gateway:  gateway,
			Listener: listener,
		})
	}

	return listenerInfos, nil
}

func MakeGatewayListener(
	gateway *core_mesh.MeshGatewayResource,
	listeners []*mesh_proto.MeshGateway_Listener,
) GatewayListener {
	listener := GatewayListener{
		Port:     listeners[0].GetPort(),
		Protocol: listeners[0].GetProtocol(),
		ResourceName: envoy_names.GetGatewayListenerName(
			gateway.Meta.GetName(),
			listeners[0].GetProtocol().String(),
			listeners[0].GetPort(),
		),
		Resources: listeners[0].GetResources(),
	}

	if listener.Protocol == mesh_proto.MeshGateway_Listener_HTTPS {
		listener.TLS = listeners[0].GetTls()
	}

	return listener
}

func (g Generator) Generate(
	ctx context.Context,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	listenerInfos, err := GatewayListenerInfoFromProxy(ctx, xdsCtx.Mesh, proxy, g.Zone)
	if err != nil {
		return nil, errors.Wrap(err, "error generating listener info from Proxy")
	}

	var limits []RuntimeResoureLimitListener

	for _, info := range listenerInfos {
		cdsResources, err := g.generateCDS(ctx, xdsCtx, proxy)
		if err != nil {
			return nil, err
		}
		resources.AddSet(cdsResources)

		ldsResources, limit, err := g.generateLDS(xdsCtx, info)
		if err != nil {
			return nil, err
		}
		resources.AddSet(ldsResources)

		if limit != nil {
			limits = append(limits, *limit)
		}

		rdsResources, err := g.generateRDS(proxy, info)
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
		Name:   "globalloadbalancer.listeners",
		Origin: metadata.OriginGlobalLoadBalancer,
		Resource: &envoy_service_runtime_v3.Runtime{
			Name:  "globalloadbalancer.listeners",
			Layer: util_proto.MustStruct(layer),
		},
	}

	return res
}

func (g Generator) generateLDS(xdsCtx xds_context.Context, info GatewayListenerInfo) (*core_xds.ResourceSet, *RuntimeResoureLimitListener, error) {
	resources := core_xds.NewResourceSet()

	listenerBuilder, limit := GenerateListener(info)

	protocol := info.Listener.Protocol
	if protocol != mesh_proto.MeshGateway_Listener_HTTP && protocol != mesh_proto.MeshGateway_Listener_HTTPS {
		return nil, nil, errors.New("only HTTP and HTTPS are supported by Koyeb Global Load Balancer")
	}

	res, filterChainBuilders, err := g.FilterChainGenerators.FilterChainGenerators[protocol].Generate(xdsCtx, info)
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

func (g Generator) generateCDS(
	ctx context.Context,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	clusterRes, err := g.ClusterGenerator.GenerateClusters(ctx, xdsCtx, proxy)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate clusters for dataplane %q", proxy.Id)
	}
	resources.AddSet(clusterRes)

	return resources, nil
}

func (g Generator) generateRDS(proxy *core_xds.Proxy, info GatewayListenerInfo) (*core_xds.ResourceSet, error) {
	switch info.Listener.Protocol {
	case mesh_proto.MeshGateway_Listener_HTTPS,
		mesh_proto.MeshGateway_Listener_HTTP:
	default:
		return nil, nil
	}

	resources := core_xds.NewResourceSet()
	routeConfig := GenerateRouteConfig(info)

	// For each koyeb app, add a VirtualHost with the domains of that app.
	// For each virtual host:
	//   * attach the right X-Koyeb-Route header, so the Ingress Gateway knows,
	//   in turn, who to route tp
	//   * choose the right aggregate cluster to route to, depending  on the
	//   path
	for _, koyebApp := range proxy.GlobalLoadBalancerProxy.KoyebApps {
		routeBuilders, err := GenerateRouteBuilders(proxy, koyebApp.Services)
		if err != nil {
			return nil, err
		}

		vh, err := GenerateVirtualHost(proxy, routeBuilders, koyebApp.Domains)
		if err != nil {
			return nil, err
		}

		routeConfig.Configure(envoy_routes.VirtualHost(vh))
	}

	// Add a fallback virtual host which catchs every other request.
	vh := GenerateFallbackVirtualHost(proxy)
	routeConfig.Configure(envoy_routes.VirtualHost(vh))

	res, err := BuildResourceSet(routeConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build route configuration resource")
	}
	resources.AddSet(res)

	return resources, nil
}
