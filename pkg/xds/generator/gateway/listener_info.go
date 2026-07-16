package gateway

import (
	"context"
	"maps"
	"sort"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	util_maps "github.com/kumahq/kuma/v3/pkg/util/maps"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/gateway/match"
	xds_topology "github.com/kumahq/kuma/v3/pkg/xds/topology"
)

// GatewayListenerInfoFromProxy processes a built-in gateway dataplane and its
// matching MeshGateway into listener metadata consumed by gateway policies.
func GatewayListenerInfoFromProxy(
	ctx context.Context,
	meshCtx xds_context.MeshContext,
	proxy *core_xds.Proxy,
) map[uint32]GatewayListenerInfo {
	gateway := xds_topology.SelectGateway(meshCtx.Resources.Gateways().Items, proxy.Dataplane.Spec.Matches)
	if gateway == nil {
		log.V(1).Info("no matching gateway for dataplane",
			"name", proxy.Dataplane.Meta.GetName(),
			"mesh", proxy.Dataplane.Meta.GetMesh(),
			"service", proxy.Dataplane.IdentifyingName(false),
		)
		return nil
	}

	log.V(1).Info("matched gateway to dataplane", "gateway", gateway.Meta.GetName(), "dataplane", proxy.Dataplane.Meta.GetName())

	for _, listener := range gateway.Spec.GetConf().GetListeners() {
		listener.Tags = mesh_proto.Merge(
			proxy.Dataplane.Spec.GetNetworking().GetGateway().GetTags(),
			gateway.Spec.GetTags(),
			listener.GetTags(),
		)
	}

	collapsed := map[uint32][]*mesh_proto.MeshGateway_Listener{}
	for _, listener := range gateway.Spec.GetConf().GetListeners() {
		collapsed[listener.GetPort()] = append(collapsed[listener.GetPort()], listener)
	}

	externalServices := meshCtx.Resources.ExternalServices()

	outboundEndpoints := core_xds.EndpointMap{}
	maps.Copy(outboundEndpoints, meshCtx.EndpointMap)
	esEndpoints := xds_topology.BuildExternalServicesEndpointMap(
		ctx,
		meshCtx.Resource,
		externalServices.Items,
		meshCtx.DataSourceLoader,
		proxy.Zone,
	)
	maps.Copy(outboundEndpoints, esEndpoints)

	listenerInfos := map[uint32]GatewayListenerInfo{}
	for _, listeners := range collapsed {
		listener, hostInfos := MakeGatewayListener(&meshCtx, gateway, listeners)
		listenerInfos[listener.Port] = GatewayListenerInfo{
			Proxy:             proxy,
			Gateway:           gateway,
			ExternalServices:  externalServices,
			OutboundEndpoints: outboundEndpoints,
			Listener:          listener,
			ListenerHostnames: hostInfos,
		}
	}

	return listenerInfos
}

// MakeGatewayListener converts a collapsed set of listener configurations into
// one listener model with host metadata and matched gateway policies.
func MakeGatewayListener(
	meshContext *xds_context.MeshContext,
	gateway *core_mesh.MeshGatewayResource,
	listeners []*mesh_proto.MeshGateway_Listener,
) (GatewayListener, []GatewayListenerHostname) {
	listener := GatewayListener{
		Port:     listeners[0].GetPort(),
		Protocol: listeners[0].GetProtocol(),
		EnvoyListenerName: envoy_names.GetGatewayListenerName(
			gateway.Meta.GetName(),
			listeners[0].GetProtocol().String(),
			listeners[0].GetPort(),
		),
		CrossMesh: listeners[0].GetCrossMesh(),
		Resources: listeners[0].GetResources(),
	}

	type hostAcc struct {
		hosts []GatewayHost
		tls   *mesh_proto.MeshGateway_TLS_Conf
	}
	hostsByName := map[string]hostAcc{}

	for _, rawListener := range listeners {
		hostname := rawListener.GetNonEmptyHostname()

		host := GatewayHost{
			Hostname: hostname,
			Tags:     rawListener.GetTags(),
		}

		hostnameKey := mesh_proto.WildcardHostname
		switch rawListener.Protocol {
		case mesh_proto.MeshGateway_Listener_HTTPS, mesh_proto.MeshGateway_Listener_TLS:
			hostnameKey = hostname
		}

		acc := hostsByName[hostnameKey]
		if acc.tls == nil {
			acc.tls = rawListener.GetTls()
		}
		acc.hosts = append(acc.hosts, host)
		hostsByName[hostnameKey] = acc
	}

	var listenerHostnames []GatewayListenerHostname
	for _, hostname := range match.SortHostnamesByExactnessDec(util_maps.AllKeys(hostsByName)) {
		hostAcc := hostsByName[hostname]
		hosts := hostAcc.hosts

		sort.Slice(hosts, func(i, j int) bool {
			return hosts[i].Hostname > hosts[j].Hostname
		})

		log.V(1).Info("applying merged gateway listener metadata",
			"listener-port", listener.Port,
			"listener-name", listener.EnvoyListenerName,
		)

		var hostInfos []GatewayHostInfo
		for _, host := range hosts {
			hostInfos = append(hostInfos, GatewayHostInfo{
				Host: host,
			})
		}

		listenerHostnames = append(listenerHostnames, GatewayListenerHostname{
			Hostname:  hostname,
			Protocol:  listeners[0].GetProtocol(),
			TLS:       hostAcc.tls,
			HostInfos: hostInfos,
		})
	}

	return listener, listenerHostnames
}
