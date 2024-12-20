package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
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

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshAccessLogType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshAccessLogType]
	if !ok {
		return nil
	}

	endpoints := &plugin_xds.EndpointAccumulator{}

	listeners := policies_xds.GatherListeners(rs)

	accessLogSocketPath := core_xds.AccessLogSocketName(proxy.Metadata.WorkDir, proxy.Id.ToResourceKey().Name, proxy.Id.ToResourceKey().Mesh)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, proxy.Dataplane, endpoints, accessLogSocketPath); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, proxy.Outbounds, proxy.Dataplane, endpoints, accessLogSocketPath); err != nil {
		return err
	}
	if err := applyToTransparentProxyListeners(policies, listeners.Ipv4Passthrough, listeners.Ipv6Passthrough, proxy.Dataplane, endpoints, accessLogSocketPath); err != nil {
		return err
	}
	if err := applyToDirectAccess(policies.ToRules, listeners.DirectAccess, proxy.Dataplane, endpoints, accessLogSocketPath); err != nil {
		return err
	}
	if err := applyToGateway(policies.GatewayRules, listeners.Gateway, ctx.Mesh.Resources.MeshLocalResources, proxy, endpoints, accessLogSocketPath); err != nil {
		return err
	}
	if err := applyToRealResources(rs, policies.ToRules.ResourceRules, ctx.Mesh, proxy.Dataplane, endpoints, accessLogSocketPath); err != nil {
		return err
	}

	if err := plugin_xds.AddLogBackendConf(*endpoints, rs, proxy); err != nil {
		return errors.Wrap(err, "unable to add configuration for MeshAccessLog backends")
	}

	return nil
}

func applyToInbounds(rules core_rules.FromRules, inboundListeners map[core_rules.InboundListener]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource, backends *plugin_xds.EndpointAccumulator, path string) error {
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

		if err := configureInbound(rules.Rules[listenerKey], dataplane, listener, backends, path); err != nil {
			return err
		}
	}
	return nil
}

func applyToOutbounds(
	rules core_rules.ToRules,
	outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener,
	outbounds xds_types.Outbounds,
	dataplane *core_mesh.DataplaneResource,
	backends *plugin_xds.EndpointAccumulator,
	path string,
) error {
	for _, outbound := range outbounds.Filter(xds_types.NonBackendRefFilter) {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound.LegacyOutbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.LegacyOutbound.GetService()

		if err := configureOutbound(rules.Rules, dataplane, core_rules.MeshServiceElement(serviceName), serviceName, listener, backends, path); err != nil {
			return err
		}
	}

	return nil
}

func applyToTransparentProxyListeners(
	policies xds.TypedMatchingPolicies, ipv4 *envoy_listener.Listener, ipv6 *envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
	backends *plugin_xds.EndpointAccumulator, path string,
) error {
	if ipv4 != nil {
		if err := configureOutbound(
			policies.ToRules.Rules,
			dataplane,
			core_rules.MeshServiceElement(core_mesh.PassThroughService),
			"external",
			ipv4,
			backends,
			path,
		); err != nil {
			return err
		}
	}

	if ipv6 != nil {
		return configureOutbound(
			policies.ToRules.Rules,
			dataplane,
			core_rules.MeshServiceElement(core_mesh.PassThroughService),
			"external",
			ipv6,
			backends,
			path,
		)
	}

	return nil
}

func applyToDirectAccess(
	rules core_rules.ToRules, directAccess map[generator.Endpoint]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
	backends *plugin_xds.EndpointAccumulator, path string,
) error {
	for endpoint, listener := range directAccess {
		name := generator.DirectAccessEndpointName(endpoint)
		return configureOutbound(
			rules.Rules,
			dataplane,
			core_rules.MeshServiceElement(core_mesh.PassThroughService),
			name,
			listener,
			backends,
			path,
		)
	}

	return nil
}

func applyToGateway(
	rules core_rules.GatewayRules,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	resources xds_context.ResourceMap,
	proxy *core_xds.Proxy,
	backends *plugin_xds.EndpointAccumulator,
	path string,
) error {
	var gateways *core_mesh.MeshGatewayResourceList
	if rawList := resources[core_mesh.MeshGatewayType]; rawList != nil {
		gateways = rawList.(*core_mesh.MeshGatewayResourceList)
	} else {
		return nil
	}

	gateway := xds_topology.SelectGateway(gateways.Items, proxy.Dataplane.Spec.Matches)
	if gateway == nil {
		return nil
	}

	for _, listenerInfo := range gateway_plugin.ExtractGatewayListeners(proxy) {
		address := proxy.Dataplane.Spec.GetNetworking().Address
		port := listenerInfo.Listener.Port
		listenerKey := core_rules.InboundListener{
			Address: address,
			Port:    port,
		}
		listener, ok := gatewayListeners[listenerKey]
		if !ok {
			continue
		}

		if toListenerRules, ok := rules.ToRules.ByListener[listenerKey]; ok {
			if err := configureOutbound(
				toListenerRules.Rules,
				proxy.Dataplane,
				core_rules.MeshElement(),
				mesh_proto.MatchAllTag,
				listener,
				backends,
				path,
			); err != nil {
				return err
			}
		}

		if fromListenerRules, ok := rules.FromRules[listenerKey]; ok {
			if err := configureInbound(fromListenerRules, proxy.Dataplane, listener, backends, path); err != nil {
				return err
			}
		}
	}

	return nil
}

func configureInbound(
	fromRules core_rules.Rules,
	dataplane *core_mesh.DataplaneResource,
	listener *envoy_listener.Listener,
	backendsAcc *plugin_xds.EndpointAccumulator,
	accessLogSocketPath string,
) error {
	serviceName := dataplane.Spec.GetIdentifyingService()

	// `from` section of MeshAccessLog only allows Mesh targetRef
	conf := core_rules.NewComputeConf[api.Conf](fromRules, core_rules.MeshElement())
	if conf == nil {
		return nil
	}

	for _, backend := range pointer.Deref(conf.Backends) {
		configurer := plugin_xds.Configurer{
			Mesh:                dataplane.GetMeta().GetMesh(),
			TrafficDirection:    envoy.TrafficDirectionInbound,
			SourceService:       mesh_proto.ServiceUnknown,
			DestinationService:  serviceName,
			Backend:             backend,
			Dataplane:           dataplane,
			AccessLogSocketPath: accessLogSocketPath,
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
	toRules core_rules.Rules,
	dataplane *core_mesh.DataplaneResource,
	element core_rules.Element,
	destinationServiceName string,
	listener *envoy_listener.Listener,
	backendsAcc *plugin_xds.EndpointAccumulator,
	path string,
) error {
	sourceService := dataplane.Spec.GetIdentifyingService()

	conf := core_rules.NewComputeConf[api.Conf](toRules, element)
	if conf == nil {
		return nil
	}

	for _, backend := range pointer.Deref(conf.Backends) {
		configurer := plugin_xds.Configurer{
			Mesh:                dataplane.GetMeta().GetMesh(),
			TrafficDirection:    envoy.TrafficDirectionOutbound,
			SourceService:       sourceService,
			DestinationService:  destinationServiceName,
			Backend:             backend,
			Dataplane:           dataplane,
			AccessLogSocketPath: path,
		}

		for _, chain := range listener.FilterChains {
			if err := configurer.Configure(chain, backendsAcc); err != nil {
				return err
			}
		}
	}

	return nil
}

func applyToRealResources(rs *core_xds.ResourceSet, rules core_rules.ResourceRules, meshCtx xds_context.MeshContext, dataplane *core_mesh.DataplaneResource, backendsAcc *plugin_xds.EndpointAccumulator, accessLogSocketPath string) error {
	for uri, resType := range rs.IndexByOrigin() {
		conf := rules.Compute(uri, meshCtx.Resources)
		if conf == nil {
			continue
		}

		for typ, resources := range resType {
			switch typ {
			case envoy_resource.ListenerType:
				err := configureListeners(resources, conf.Conf[0].(api.Conf), dataplane, backendsAcc, accessLogSocketPath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func configureListeners(resources []*core_xds.Resource, conf api.Conf, dataplane *core_mesh.DataplaneResource, backendsAcc *plugin_xds.EndpointAccumulator, accessLogSocketPath string) error {
	sourceService := dataplane.Spec.GetIdentifyingService()
	for _, backend := range pointer.Deref(conf.Backends) {
		for _, resource := range resources {
			configurer := plugin_xds.Configurer{
				Mesh:                dataplane.GetMeta().GetMesh(),
				TrafficDirection:    envoy.TrafficDirectionOutbound,
				SourceService:       sourceService,
				DestinationService:  resource.ResourceOrigin.Name,
				Backend:             backend,
				Dataplane:           dataplane,
				AccessLogSocketPath: accessLogSocketPath,
			}

			for _, chain := range resource.Resource.(*envoy_listener.Listener).FilterChains {
				if err := configurer.Configure(chain, backendsAcc); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
