package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/plugin/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ core_plugins.PolicyPlugin = &plugin{}
var log = core.Log.WithName("MeshTrace")

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTraceType, dataplane, resources)
}

func (p plugin) Apply(rs *xds.ResourceSet, ctx xds_context.Context, proxy *xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshTraceType]
	if !ok {
		return nil
	}

	listeners := gatherListeners(rs)

	if err := applyToInbounds(policies.FromRules, listeners.inbound, proxy.Dataplane); err != nil {
		return err
	}

	log.Info("apply is not implemented")
	return nil
}

type listeners struct {
	inbound  map[xds.InboundListener]*envoy_listener.Listener
	outbound map[mesh_proto.OutboundInterface]*envoy_listener.Listener
}

func gatherListeners(rs *xds.ResourceSet) listeners {
	listeners := listeners{
		inbound:      map[xds.InboundListener]*envoy_listener.Listener{},
		outbound:     map[mesh_proto.OutboundInterface]*envoy_listener.Listener{},
	}

	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		listener := res.Resource.(*envoy_listener.Listener)
		address := listener.GetAddress().GetSocketAddress()

		switch res.Origin {
		case generator.OriginOutbound:
			listeners.outbound[mesh_proto.OutboundInterface{
				DataplaneIP:   address.GetAddress(),
				DataplanePort: address.GetPortValue(),
			}] = listener
		case generator.OriginInbound:
			listeners.inbound[xds.InboundListener{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		default:
			continue
		}
	}

	return listeners
}

func applyToInbounds(
	rules xds.FromRules, inboundListeners map[xds.InboundListener]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
) error {
	for _, inbound := range dataplane.Spec.GetNetworking().GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := xds.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}
		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}

		serviceName := inbound.GetTags()[mesh_proto.ServiceTag]

		if err := configureInbound(rules.Rules[listenerKey], dataplane, xds.MeshService(serviceName), listener); err != nil {
			return err
		}
	}

	return nil
}

func configureInbound(
	fromRules xds.Rules,
	dataplane *core_mesh.DataplaneResource,
	subset xds.Subset,
	listener *envoy_listener.Listener,
) error {
	serviceName := dataplane.Spec.GetIdentifyingService()

	var conf *api.MeshTrace_Conf
	if computed := fromRules.Compute(subset); computed != nil {
		conf = computed.(*api.MeshTrace_Conf)
	} else {
		return nil
	}

	configurer := plugin_xds.Configurer{
		Conf: conf,
		Service: serviceName,
	}

	for _, chain := range listener.FilterChains {
		if err := configurer.Configure(chain); err != nil {
			return err
		}
	}

	return nil
}
