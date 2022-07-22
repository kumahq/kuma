package generator

import (
	"net/url"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	model "github.com/kumahq/kuma/pkg/core/xds"
	defaults_mesh "github.com/kumahq/kuma/pkg/defaults/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

type ProtocolInboundPair struct {
	iface    mesh_proto.InboundInterface
	protocol core_mesh.Protocol
}

const (
	// OriginProbes is a marker to indicate by which ProxyGenerator resources were generated.
	OriginProbe  = "probe"
	listenerName = "probe:listener"
)

type ProbeProxyGenerator struct {
}

func (g ProbeProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	probes := proxy.Dataplane.Spec.Probes
	if probes == nil {
		return nil, nil
	}

	virtualHostBuilder := envoy_routes.NewVirtualHostBuilder(proxy.APIVersion).
		Configure(envoy_routes.CommonVirtualHost("probe"))

	inbounds := map[uint32]ProtocolInboundPair{}
	for _, inbound := range proxy.Dataplane.Spec.Networking.Inbound {
		iface := proxy.Dataplane.Spec.GetNetworking().ToInboundInterface(inbound)
		inbounds[inbound.Port] = ProtocolInboundPair{
			protocol: core_mesh.ParseProtocol(inbound.GetProtocol()),
			iface:    iface,
		}
	}
	resources := model.NewResourceSet()
	for i, endpoint := range probes.Endpoints {
		matchURL, err := url.Parse(endpoint.Path)
		if err != nil {
			return nil, err
		}
		newURL, err := url.Parse(endpoint.InboundPath)
		if err != nil {
			return nil, err
		}
		if val, exists := inbounds[endpoint.InboundPort]; exists {
			var clusterName string
			if ctx.ControlPlane.EnableInboundPassthrough &&
				proxy.Dataplane.Spec.GetNetworking() != nil &&
				proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying() != nil &&
				proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying().RedirectPortInbound != 0 {
				clusterName = names.GetProbeClusterName(val.iface.WorkloadPort)
				clusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion).
					Configure(envoy_clusters.ProvidedEndpointCluster(clusterName, false, model.Endpoint{Target: val.iface.WorkloadIP, Port: val.iface.WorkloadPort})).
					Configure(envoy_clusters.Timeout(defaults_mesh.DefaultInboundTimeout(), val.protocol))

				switch val.protocol {
				case core_mesh.ProtocolHTTP:
					clusterBuilder.Configure(envoy_clusters.Http())
				case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
					clusterBuilder.Configure(envoy_clusters.Http2())
				}
				envoyCluster, err := clusterBuilder.Build()
				if err != nil {
					return nil, errors.Wrapf(err, "%s: could not generate cluster %s", validators.RootedAt("dataplane").Field("probe").Field("endpoints").Index(i), clusterName)
				}
				resources.Add(&model.Resource{
					Name:     clusterName,
					Resource: envoyCluster,
					Origin:   OriginProbe,
				})
			} else {
				// we don't have to generate inbound cluster, because they are
				// generated in inbound_proxy_generator
				clusterName = names.GetLocalClusterName(endpoint.InboundPort)
			}

			virtualHostBuilder.Configure(
				envoy_routes.Route(matchURL.Path, newURL.Path, clusterName, true))
		} else {
			// On Kubernetes we are overriding probes for every container, but there is no guarantee that given
			// probe will have an equivalent in inbound interface (ex. sidecar that is not selected by any service).
			// In this situation there is no local cluster therefore we are sending redirect to a real destination.
			// System responsible for using virtual probes needs to support redirect (kubelet on K8S supports it).
			virtualHostBuilder.Configure(
				envoy_routes.Redirect(matchURL.Path, newURL.Path, true, endpoint.InboundPort))
		}
	}

	probeListener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
		Configure(envoy_listeners.InboundListener(listenerName, proxy.Dataplane.Spec.GetNetworking().GetAddress(), probes.Port, model.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
			Configure(envoy_listeners.HttpConnectionManager(listenerName, false)).
			Configure(envoy_listeners.HttpStaticRoute(envoy_routes.NewRouteConfigurationBuilder(proxy.APIVersion).
				Configure(envoy_routes.VirtualHost(virtualHostBuilder)))))).
		Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener %s", listenerName)
	}

	resources.Add(&model.Resource{
		Name:     listenerName,
		Resource: probeListener,
		Origin:   OriginProbe,
	})

	return resources, nil
}
