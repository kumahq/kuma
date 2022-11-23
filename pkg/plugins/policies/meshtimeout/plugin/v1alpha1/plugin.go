package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/plugin/xds"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_common "github.com/kumahq/kuma/pkg/xds/generator"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var _ core_plugins.PolicyPlugin = &plugin{}
var log = core.Log.WithName("MeshTimeout")

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTimeoutType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshTimeoutType]
	if !ok {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, clusters.Inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, clusters.Outbound, proxy.Dataplane, proxy.Routing); err != nil {
		return err
	}
	if err := applyToGateway(policies.FromRules, listeners.Gateway, clusters.Gateway, ctx.Mesh.Resources.MeshLocalResources, proxy.Dataplane); err != nil {
		return err
	}
	return nil
}

func applyToInbounds(fromRules core_xds.FromRules, inboundListeners map[core_xds.InboundListener]*envoy_listener.Listener, inboundClusters map[string]*envoy_cluster.Cluster, dataplane *core_mesh.DataplaneResource) error {
	for _, inbound := range dataplane.Spec.Networking.GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := xds.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}

		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}

		rules, ok := fromRules.Rules[listenerKey]
		if !ok {
			continue
		}

		cluster, ok := inboundClusters[envoy_names.GetLocalClusterName(listenerKey.Port)]
		if !ok {
			continue
		}

		protocol := core_mesh.ParseProtocol(inbound.GetProtocol())
		if err := configure(rules, xds.MeshSubset(), protocol, listener, cluster); err != nil {
			return err
		}
	}

	return nil
}

func applyToOutbounds(rules core_xds.ToRules, outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener, outboundClusters map[string]*envoy_cluster.Cluster, dataplane *core_mesh.DataplaneResource, routing core_xds.Routing) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]
		cluster, ok := outboundClusters[serviceName]
		if !ok {
			continue
		}

		protocol := inferProtocol(routing, serviceName)
		if err := configure(rules.Rules, xds.MeshService(serviceName), protocol, listener, cluster); err != nil {
			return err
		}
		if serviceName == "test-server" {
			log.Info("e2e test", "rules", rules.Rules, "protocol", protocol, "listener", listener, "cluster", cluster)
		}
	}

	return nil
}

func applyToGateway(
	fromRules xds.FromRules, gatewayListeners map[xds.InboundListener]*envoy_listener.Listener, gatewayClusters map[string]*envoy_cluster.Cluster, resources xds_context.ResourceMap, dataplane *core_mesh.DataplaneResource,
) error {
	var gateways *core_mesh.MeshGatewayResourceList
	if rawList := resources[core_mesh.MeshGatewayType]; rawList != nil {
		gateways = rawList.(*core_mesh.MeshGatewayResourceList)
	} else {
		return nil
	}
	log.Info("gateways configured")
	gateway := xds_topology.SelectGateway(gateways.Items, dataplane.Spec.Matches)
	if gateway == nil {
		return nil
	}
	// TODO remove log
	log.Info("gateway selected", "gateway", gateway)

	for _, listener := range gateway.Spec.GetConf().GetListeners() {
		address := dataplane.Spec.GetNetworking().Address
		port := listener.GetPort()
		listenerKey := xds.InboundListener{
			Address: address,
			Port:    port,
		}
		gatewayListener, ok := gatewayListeners[listenerKey]
		if !ok {
			continue
		}

		rules, ok := fromRules.Rules[listenerKey]
		if !ok {
			continue
		}

		// TODO find a way to extract clusters for gateway route
		cluster, ok := gatewayClusters[""]
		if !ok {
			continue
		}

		// TODO remove log
		log.Info("gateway listener", "gatewayListener", gatewayListener)
		// TODO fix method invocation
		if err := configure(
			rules,
			xds.MeshSubset(),
			toProtocol(listener.GetProtocol()),
			gatewayListener,
			cluster,
		); err != nil {
			return err
		}
	}

	return nil
}

func configure(rules core_xds.Rules, subset core_xds.Subset, protocol core_mesh.Protocol, listener *envoy_listener.Listener, cluster *envoy_cluster.Cluster) error {
	var conf api.Conf
	if computed := rules.Compute(subset); computed != nil {
		conf = computed.Conf.(api.Conf)
	} else {
		return nil
	}

	configurer := plugin_xds.Configurer{
		Conf:     conf,
		Protocol: protocol,
	}

	for _, chain := range listener.FilterChains {
		if err := configurer.ConfigureListener(chain); err != nil {
			return err
		}
	}

	if err := configurer.ConfigureCluster(cluster); err != nil {
		return err
	}
	return nil
}

func inferProtocol(routing core_xds.Routing, serviceName string) core_mesh.Protocol {
	var allEndpoints []core_xds.Endpoint
	endpoints := core_xds.EndpointList(routing.OutboundTargets[serviceName])
	allEndpoints = append(allEndpoints, endpoints...)
	endpoints = routing.ExternalServiceOutboundTargets[serviceName]
	allEndpoints = append(allEndpoints, endpoints...)

	return envoy_common.InferServiceProtocol(allEndpoints)
}

func toProtocol(p mesh_proto.MeshGateway_Listener_Protocol) core_mesh.Protocol {
	return core_mesh.ParseProtocol(mesh_proto.MeshGateway_Listener_Protocol_name[int32(p.Number())])
}
