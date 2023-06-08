package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/plugin/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshAccessLogType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshAccessLogType]
	if !ok {
		return nil
	}

	endpoints := &plugin_xds.EndpointAccumulator{}

	listeners := policies_xds.GatherListeners(rs)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, proxy.Dataplane, endpoints); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, proxy.Dataplane, endpoints); err != nil {
		return err
	}
	if err := applyToTransparentProxyListeners(policies, listeners.Ipv4Passthrough, listeners.Ipv6Passthrough, proxy.Dataplane, endpoints); err != nil {
		return err
	}
	if err := applyToDirectAccess(policies.ToRules, listeners.DirectAccess, proxy.Dataplane, endpoints); err != nil {
		return err
	}
	if err := applyToGateway(policies.ToRules, listeners.Gateway, ctx.Mesh.Resources.MeshLocalResources, proxy.Dataplane, endpoints); err != nil {
		return err
	}

	if err := plugin_xds.HandleClusters(*endpoints, rs, proxy); err != nil {
		return errors.Wrap(err, "unable to handle clusters for policy")
	}

	return nil
}

func applyToInbounds(
	rules core_rules.FromRules, inboundListeners map[core_rules.InboundListener]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
	backends *plugin_xds.EndpointAccumulator,
) error {
	for _, inbound := range dataplane.Spec.GetNetworking().GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := core_rules.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}
		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}

		if err := configureInbound(rules.Rules[listenerKey], dataplane, listener, backends); err != nil {
			return err
		}
	}
	return nil
}

func applyToOutbounds(
	rules core_rules.ToRules, outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
	backends *plugin_xds.EndpointAccumulator,
) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]

		if err := configureOutbound(rules, dataplane, core_rules.MeshService(serviceName), serviceName, listener, backends); err != nil {
			return err
		}
	}

	return nil
}

func applyToTransparentProxyListeners(
	policies xds.TypedMatchingPolicies, ipv4 *envoy_listener.Listener, ipv6 *envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
	backends *plugin_xds.EndpointAccumulator,
) error {
	if ipv4 != nil {
		if err := configureOutbound(
			policies.ToRules,
			dataplane,
			core_rules.MeshService(core_mesh.PassThroughService),
			"external",
			ipv4,
			backends,
		); err != nil {
			return err
		}
	}

	if ipv6 != nil {
		return configureOutbound(
			policies.ToRules,
			dataplane,
			core_rules.MeshService(core_mesh.PassThroughService),
			"external",
			ipv6,
			backends,
		)
	}

	return nil
}

func applyToDirectAccess(
	rules core_rules.ToRules, directAccess map[generator.Endpoint]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
	backends *plugin_xds.EndpointAccumulator,
) error {
	for endpoint, listener := range directAccess {
		name := generator.DirectAccessEndpointName(endpoint)
		return configureOutbound(
			rules,
			dataplane,
			core_rules.MeshService(core_mesh.PassThroughService),
			name,
			listener,
			backends,
		)
	}

	return nil
}

func applyToGateway(
	rules core_rules.ToRules, gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener, resources xds_context.ResourceMap, dataplane *core_mesh.DataplaneResource,
	backends *plugin_xds.EndpointAccumulator,
) error {
	var gateways *core_mesh.MeshGatewayResourceList
	if rawList := resources[core_mesh.MeshGatewayType]; rawList != nil {
		gateways = rawList.(*core_mesh.MeshGatewayResourceList)
	} else {
		return nil
	}

	gateway := xds_topology.SelectGateway(gateways.Items, dataplane.Spec.Matches)
	if gateway == nil {
		return nil
	}

	for _, listener := range gateway.Spec.GetConf().GetListeners() {
		address := dataplane.Spec.GetNetworking().Address
		port := listener.GetPort()
		listener, ok := gatewayListeners[core_rules.InboundListener{
			Address: address,
			Port:    port,
		}]
		if !ok {
			continue
		}

		if err := configureOutbound(
			rules,
			dataplane,
			core_rules.Subset{},
			mesh_proto.MatchAllTag,
			listener,
			backends,
		); err != nil {
			return err
		}
	}

	return nil
}

func configureInbound(
	fromRules core_rules.Rules,
	dataplane *core_mesh.DataplaneResource,
	listener *envoy_listener.Listener,
	backendsAcc *plugin_xds.EndpointAccumulator,
) error {
	serviceName := dataplane.Spec.GetIdentifyingService()

	var conf api.Conf
	// `from` section of MeshAccessLog only allows Mesh targetRef
	if computed := fromRules.Compute(core_rules.MeshSubset()); computed != nil {
		conf = computed.Conf.(api.Conf)
	} else {
		return nil
	}

	for _, backend := range pointer.Deref(conf.Backends) {
		configurer := plugin_xds.Configurer{
			Mesh:               dataplane.GetMeta().GetMesh(),
			TrafficDirection:   envoy.TrafficDirectionInbound,
			SourceService:      mesh_proto.ServiceUnknown,
			DestinationService: serviceName,
			Backend:            backend,
			Dataplane:          dataplane,
		}

		for _, chain := range listener.FilterChains {
			if err := configurer.Configure(chain, backendsAcc); err != nil {
				return err
			}
		}
	}

	return nil
}

func configureOutbound(
	toRules core_rules.ToRules,
	dataplane *core_mesh.DataplaneResource,
	subset core_rules.Subset,
	destinationServiceName string,
	listener *envoy_listener.Listener,
	backendsAcc *plugin_xds.EndpointAccumulator,
) error {
	sourceService := dataplane.Spec.GetIdentifyingService()

	var conf api.Conf
	if computed := toRules.Rules.Compute(subset); computed != nil {
		conf = computed.Conf.(api.Conf)
	} else {
		return nil
	}

	for _, backend := range pointer.Deref(conf.Backends) {
		configurer := plugin_xds.Configurer{
			Mesh:               dataplane.GetMeta().GetMesh(),
			TrafficDirection:   envoy.TrafficDirectionOutbound,
			SourceService:      sourceService,
			DestinationService: destinationServiceName,
			Backend:            backend,
			Dataplane:          dataplane,
		}

		for _, chain := range listener.FilterChains {
			if err := configurer.Configure(chain, backendsAcc); err != nil {
				return err
			}
		}
	}

	return nil
}
