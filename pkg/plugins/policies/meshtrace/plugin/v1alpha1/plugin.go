package v1alpha1

import (
	net_url "net/url"
	"strconv"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/plugin/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

const MeshTraceCluster = "meshTrace"
const MeshTraceOrigin = "meshTrace"

var _ core_plugins.PolicyPlugin = &plugin{}

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

	if err := applyToInbounds(policies.SingleItemRules, listeners.inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.SingleItemRules, listeners.outbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToClusters(policies.SingleItemRules, rs, proxy); err != nil {
		return err
	}

	return nil
}

type listeners struct {
	inbound  map[xds.InboundListener]*envoy_listener.Listener
	outbound map[mesh_proto.OutboundInterface]*envoy_listener.Listener
}

func gatherListeners(rs *xds.ResourceSet) listeners {
	listeners := listeners{
		inbound:  map[xds.InboundListener]*envoy_listener.Listener{},
		outbound: map[mesh_proto.OutboundInterface]*envoy_listener.Listener{},
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

func applyToInbounds(rules xds.SingleItemRules, inboundListeners map[xds.InboundListener]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource) error {
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

		if err := configureListener(rules.Rules, dataplane, xds.MeshService(serviceName), listener); err != nil {
			return err
		}
	}

	return nil
}

func applyToOutbounds(
	rules xds.SingleItemRules, outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]

		if err := configureListener(rules.Rules, dataplane, xds.MeshService(serviceName), listener); err != nil {
			return err
		}
	}

	return nil
}

func configureListener(
	rules xds.Rules,
	dataplane *core_mesh.DataplaneResource,
	subset xds.Subset,
	listener *envoy_listener.Listener,
) error {
	serviceName := dataplane.Spec.GetIdentifyingService()

	var conf *api.MeshTrace_Conf
	if computed := rules.Compute(subset); computed != nil {
		conf = computed.(*api.MeshTrace_Conf)
	} else {
		return nil
	}

	configurer := plugin_xds.Configurer{
		Conf:        conf,
		Service:     serviceName,
		ClusterName: MeshTraceCluster,
	}

	for _, chain := range listener.FilterChains {
		if err := configurer.Configure(chain); err != nil {
			return err
		}
	}

	return nil
}

func applyToClusters(rules xds.SingleItemRules, rs *xds.ResourceSet, proxy *xds.Proxy) error {
	for _, rule := range rules.Rules {
		conf := rule.Conf.(*api.MeshTrace_Conf)
		backend := conf.GetBackends()[0]
		var endpoint *xds.Endpoint
		var err error

		if backend.GetZipkin() != nil {
			endpoint, err = endpointForZipkin(backend.GetZipkin())
			if err != nil {
				return errors.Wrap(err, "could not generate zipkin cluster")
			}
		} else {
			endpoint, err = endpointForDatadog(backend.GetDatadog())
			if err != nil {
				return errors.Wrap(err, "could not generate zipkin cluster")
			}
		}

		res, err := clusters.NewClusterBuilder(proxy.APIVersion).
			Configure(clusters.ProvidedEndpointCluster(MeshTraceCluster, proxy.Dataplane.IsIPv6(), *endpoint)).
			Configure(clusters.ClientSideTLS([]xds.Endpoint{*endpoint})).
			Configure(clusters.DefaultTimeout()).
			Build()
		if err != nil {
			return err
		}

		rs.Add(&xds.Resource{Name: MeshTraceCluster, Origin: MeshTraceOrigin, Resource: res})
	}

	return nil
}

func endpointForZipkin(cfg *api.MeshTrace_ZipkinBackend) (*xds.Endpoint, error) {
	url, err := net_url.ParseRequestURI(cfg.Url)
	if err != nil {
		return nil, errors.Wrap(err, "invalid URL of Zipkin")
	}
	port, err := strconv.Atoi(url.Port())
	if err != nil {
		return nil, err
	}
	return &xds.Endpoint{
		Target: url.Hostname(),
		Port:   uint32(port),
		ExternalService: &xds.ExternalService{
			TLSEnabled:         url.Scheme == "https",
			AllowRenegotiation: true,
		},
	}, nil
}

func endpointForDatadog(cfg *api.MeshTrace_DatadogBackend) (*xds.Endpoint, error) {
	if cfg.Port > 0xFFFF || cfg.Port < 1 {
		return nil, errors.Errorf("invalid Datadog port number %d. Must be in range 1-65535", cfg.Port)
	}
	return &xds.Endpoint{
		Target: cfg.Address,
		Port:   cfg.Port,
	}, nil
}
