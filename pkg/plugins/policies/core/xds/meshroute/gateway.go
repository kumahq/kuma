package meshroute

import (
	"context"
	"slices"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type Sublistener struct {
	Hostname string
	Tags     map[string]string
	TLS      *mesh_proto.MeshGateway_TLS_Conf
}
type listenersHostnames struct {
	listener     *mesh_proto.MeshGateway_Listener
	sublisteners []Sublistener
}

type MapGatewayRulesToHosts func(
	xds_context.ResourceMap,
	rules.GatewayRules,
	string,
	*mesh_proto.MeshGateway_Listener,
	[]Sublistener,
) []plugin_gateway.GatewayHostInfo

func CollectListenerInfos(
	ctx context.Context,
	meshCtx xds_context.MeshContext,
	gateway *core_mesh.MeshGatewayResource,
	proxy *core_xds.Proxy,
	rawRules rules.GatewayRules,
	validProtocols []mesh_proto.MeshGateway_Listener_Protocol,
	mapRules MapGatewayRulesToHosts,
) map[uint32]plugin_gateway.GatewayListenerInfo {
	networking := proxy.Dataplane.Spec.GetNetworking()
	listenersByPort := map[uint32]listenersHostnames{}
	for _, listener := range gateway.Spec.GetConf().GetListeners() {
		if !slices.Contains(validProtocols, listener.Protocol) {
			continue
		}
		listenerAcc, ok := listenersByPort[listener.GetPort()]
		if !ok {
			listenerAcc = listenersHostnames{
				listener: listener,
			}
		}
		hostname := listener.GetNonEmptyHostname()
		listenerAcc.sublisteners = append(listenerAcc.sublisteners, Sublistener{
			Hostname: hostname,
			Tags: mesh_proto.Merge(
				networking.GetGateway().GetTags(),
				gateway.Spec.GetTags(),
				listener.GetTags(),
			),
			TLS: listener.GetTls(),
		})
		listenersByPort[listener.GetPort()] = listenerAcc
	}

	infos := map[uint32]plugin_gateway.GatewayListenerInfo{}

	for port, listener := range listenersByPort {
		externalServices := meshCtx.Resources.ExternalServices()

		matchedExternalServices := permissions.MatchExternalServicesTrafficPermissions(
			proxy.Dataplane, externalServices, meshCtx.Resources.TrafficPermissions(),
		)

		outboundEndpoints := core_xds.EndpointMap{}
		for k, v := range meshCtx.EndpointMap {
			outboundEndpoints[k] = v
		}

		esEndpoints := xds_topology.BuildExternalServicesEndpointMap(
			ctx,
			meshCtx.Resource,
			matchedExternalServices,
			meshCtx.DataSourceLoader,
			proxy.Zone,
		)
		for k, v := range esEndpoints {
			outboundEndpoints[k] = v
		}

		hostInfos := mapRules(
			meshCtx.Resources.MeshLocalResources,
			rawRules,
			networking.Address,
			listener.listener,
			listener.sublisteners,
		)
		var filters []plugin_gateway.GatewayListenerFilter
		for _, sublistener := range listener.sublisteners {
			filters = append(filters, plugin_gateway.GatewayListenerFilter{
				Hostnames: []string{sublistener.Hostname},
				TLS:       sublistener.TLS,
			})
		}
		infos[port] = plugin_gateway.GatewayListenerInfo{
			Proxy:             proxy,
			Gateway:           gateway,
			HostInfos:         hostInfos,
			ExternalServices:  externalServices,
			OutboundEndpoints: outboundEndpoints,
			Listener: plugin_gateway.GatewayListener{
				Port:     port,
				Protocol: listener.listener.GetProtocol(),
				ResourceName: envoy_names.GetGatewayListenerName(
					gateway.Meta.GetName(),
					listener.listener.GetProtocol().String(),
					port,
				),
				CrossMesh: listener.listener.GetCrossMesh(),
				Resources: listener.listener.GetResources(),
				Filters:   filters,
			},
		}
	}

	return infos
}
