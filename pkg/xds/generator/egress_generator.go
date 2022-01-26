package generator

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

const (
	EgressProxy = "egress-proxy"

	// OriginEgress is a marker to indicate by which ProxyGenerator resources were generated.
	OriginEgress = "egress"
)

type EgressGenerator struct {
}

func (g EgressGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()

	listener, err := g.generateLDS(proxy, proxy.ZoneEgressProxy.ZoneEgressResource, proxy.APIVersion)
	if err != nil {
		return nil, err
	}

	resources.Add(&model.Resource{
		Name:     listener.GetName(),
		Origin:   OriginEgress,
		Resource: listener,
	})

	return resources, nil
}

// generateLDS generates one Ingress Listener
// It assumes that mTLS is on. Using TLSInspector we sniff SNI value.
// SNI value has service name and tag values specified with the following format: "backend{cluster=2,version=1}"
func (g EgressGenerator) generateLDS(
	proxy *model.Proxy,
	egress *core_mesh.ZoneEgressResource,
	apiVersion envoy_common.APIVersion,
) (envoy_common.NamedResource, error) {
	networking := egress.Spec.GetNetworking()

	listenerBuilder := envoy_listeners.NewListenerBuilder(apiVersion).
		Configure(envoy_listeners.InboundListener(
			// TODO (bartsmykla): move the name generation somewhere else probably
			"egress:listener",
			networking.GetAddress(),
			networking.GetPort(),
			model.SocketAddressProtocolTCP,
		)).
		Configure(envoy_listeners.TLSInspector()).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(apiVersion)))

	return listenerBuilder.Build()
}
