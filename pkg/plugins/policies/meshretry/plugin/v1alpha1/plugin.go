package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/plugin/xds"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshRetryType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, _ xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshRetryType]
	if !ok {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)

	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, proxy.Dataplane, proxy.Routing); err != nil {
		return err
	}

	return nil
}

func applyToOutbounds(rules core_xds.ToRules, outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource, routing core_xds.Routing) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)
		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		protocol := policies_xds.InferProtocol(routing, serviceName)

		if err := configure(rules.Rules, core_xds.MeshService(serviceName), protocol, listener); err != nil {
			return err
		}
	}

	return nil
}

func configure(rules core_xds.Rules, subset core_xds.Subset, protocol core_mesh.Protocol, listener *envoy_listener.Listener) error {
	var conf api.Conf
	if computed := rules.Compute(subset); computed != nil {
		conf = computed.Conf.(api.Conf)
	} else {
		return nil
	}

	configurer := plugin_xds.Configurer{
		Retry:    &conf,
		Protocol: protocol,
	}

	for _, filterChain := range listener.FilterChains {
		if err := configurer.Configure(filterChain); err != nil {
			return err
		}
	}

	return nil
}
