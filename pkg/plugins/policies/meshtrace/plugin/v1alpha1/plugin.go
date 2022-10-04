package v1alpha1

import (
	net_url "net/url"
	"strconv"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"

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
	if err := applyToListeners(policies.SingleItemRules, listeners, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToClusters(policies.SingleItemRules, rs, proxy); err != nil {
		return err
	}

	return nil
}

func gatherListeners(rs *xds.ResourceSet) []*xds.Resource {
	listeners := []*xds.Resource{}

	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		switch res.Origin {
		case generator.OriginOutbound:
			listeners = append(listeners, res)
		case generator.OriginInbound:
			listeners = append(listeners, res)
		default:
			continue
		}
	}

	return listeners
}

func applyToListeners(rules xds.SingleItemRules, inboundListeners []*xds.Resource, dataplane *core_mesh.DataplaneResource) error {
	for _, inboundListener := range inboundListeners {
		if err := configureListener(rules, dataplane, inboundListener); err != nil {
			return err
		}
	}

	return nil
}

func configureListener(rules xds.SingleItemRules, dataplane *core_mesh.DataplaneResource, listenerResource *xds.Resource) error {
	serviceName := dataplane.Spec.GetIdentifyingService()
	listener := listenerResource.Resource.(*envoy_listener.Listener)
	rawConf := rules.Rules[0].Conf
	conf := rawConf.(*api.MeshTrace_Conf)
	destination := ""

	if listenerResource.Origin == generator.OriginOutbound {
		i := slices.IndexFunc(dataplane.Spec.Networking.Outbound, func(outbound *mesh_proto.Dataplane_Networking_Outbound) bool {
			outboundInterface := dataplane.Spec.Networking.ToOutboundInterface(outbound)
			address := listener.GetAddress().GetSocketAddress()
			return outboundInterface.DataplaneIP == address.GetAddress() && outboundInterface.DataplanePort == address.GetPortValue()
		})

		if i != -1 {
			outbound := dataplane.Spec.Networking.GetOutbound()[i]
			destination = outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]
		}
	}

	configurer := plugin_xds.Configurer{
		Conf:             conf,
		Service:          serviceName,
		TrafficDirection: listener.TrafficDirection,
		Destination:      destination,
	}

	for _, chain := range listener.FilterChains {
		if err := configurer.Configure(chain); err != nil {
			return err
		}
	}

	return nil
}

func applyToClusters(rules xds.SingleItemRules, rs *xds.ResourceSet, proxy *xds.Proxy) error {
	rawConf := rules.Rules[0].Conf
	conf := rawConf.(*api.MeshTrace_Conf)

	backend := conf.GetBackends()[0]
	if backend == nil {
		return nil
	}

	var endpoint *xds.Endpoint
	var err error
	var provider string

	if backend.GetZipkin() != nil {
		endpoint, err = endpointForZipkin(backend.GetZipkin())
		provider = "zipkin"
		if err != nil {
			return errors.Wrap(err, "could not generate zipkin cluster")
		}
	} else {
		endpoint, err = endpointForDatadog(backend.GetDatadog())
		provider = "datadog"
		if err != nil {
			return errors.Wrap(err, "could not generate zipkin cluster")
		}
	}

	res, err := clusters.NewClusterBuilder(proxy.APIVersion).
		Configure(clusters.ProvidedEndpointCluster(plugin_xds.GetTracingClusterName(provider), proxy.Dataplane.IsIPv6(), *endpoint)).
		Configure(clusters.ClientSideTLS([]xds.Endpoint{*endpoint})).
		Configure(clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return err
	}

	rs.Add(&xds.Resource{Name: plugin_xds.GetTracingClusterName(provider), Origin: MeshTraceOrigin, Resource: res})

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
