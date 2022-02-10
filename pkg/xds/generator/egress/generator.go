package egress

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

const (
	EgressProxy = "egress-proxy"

	// OriginEgress is a marker to indicate by which ProxyGenerator resources
	// were generated.
	OriginEgress = "egress"
)

var (
	log = core.Log.WithName("xds").WithName("generator").WithName("egress")
)

// ZoneEgressGenerator is responsible for generating xDS resources for
// a single ZoneEgress.
type ZoneEgressGenerator interface {
	Generate(proxy *core_xds.Proxy, listenerBuilder *envoy_listeners.ListenerBuilder, meshResources *core_xds.MeshResources) (*core_xds.ResourceSet, error)
}

// Generator generates xDS resources for an entire ZoneEgress.
type Generator struct {
	Generators []ZoneEgressGenerator
}

func makeListenerBuilder(
	apiVersion envoy_common.APIVersion,
	zoneEgress *core_mesh.ZoneEgressResource,
) *envoy_listeners.ListenerBuilder {
	networking := zoneEgress.Spec.GetNetworking()

	address := networking.GetAddress()
	port := networking.GetPort()

	return envoy_listeners.NewListenerBuilder(apiVersion).
		Configure(
			envoy_listeners.InboundListener(
				envoy_names.GetInboundListenerName(address, port),
				address, port,
				core_xds.SocketAddressProtocolTCP,
			),
			envoy_listeners.TLSInspector(),
		)
}

func (g Generator) Generate(
	_ xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	listenerBuilder := makeListenerBuilder(
		proxy.APIVersion,
		proxy.ZoneEgressProxy.ZoneEgressResource,
	)

	for _, meshResources := range proxy.ZoneEgressProxy.MeshResourcesList {
		for _, generator := range g.Generators {
			rs, err := generator.Generate(proxy, listenerBuilder, meshResources)
			if err != nil {
				err := errors.Wrapf(
					err,
					"%T failed to generate resources for zone egress %q",
					generator,
					proxy.Id,
				)
				return nil, err
			}

			resources.AddSet(rs)
		}

		listener, err := listenerBuilder.Build()
		if err != nil {
			return nil, err
		}

		resources.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   OriginEgress,
			Resource: listener,
		})
	}

	return resources, nil
}
