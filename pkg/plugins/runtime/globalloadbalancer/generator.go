package globalloadbalancer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
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
	return GatewayListener{
		Port:     listeners[0].GetPort(),
		Protocol: listeners[0].GetProtocol(),
		ResourceName: envoy_names.GetGatewayListenerName(
			gateway.Meta.GetName(),
			listeners[0].GetProtocol().String(),
			listeners[0].GetPort(),
		),
		Resources: listeners[0].GetResources(),
	}
}

func (g Generator) Generate(
	ctx context.Context,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	_, err := GatewayListenerInfoFromProxy(ctx, xdsCtx.Mesh, proxy, g.Zone)
	if err != nil {
		return nil, errors.Wrap(err, "error generating listener info from Proxy")
	}

	// var limits []RuntimeResoureLimitListener

	// for _, info := range listenerInfos {
	// 	cdsResources, err := g.generateCDS(ctx, xdsCtx, proxy)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	resources.AddSet(cdsResources)

	// 	ldsResources, limit, err := g.generateLDS(xdsCtx, info)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	resources.AddSet(ldsResources)

	// 	if limit != nil {
	// 		limits = append(limits, *limit)
	// 	}

	// 	rdsResources, err := g.generateRDS(xdsCtx, proxy, info)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	resources.AddSet(rdsResources)
	// }

	// resources.Add(g.generateRTDS(limits))

	return resources, nil
}
