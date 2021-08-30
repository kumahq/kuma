package generator

import (
	"net/url"

	"github.com/pkg/errors"

	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

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

	portSet := map[uint32]bool{}
	for _, inbound := range proxy.Dataplane.Spec.Networking.Inbound {
		portSet[proxy.Dataplane.Spec.Networking.ToInboundInterface(inbound).WorkloadPort] = true
	}
	for _, endpoint := range probes.Endpoints {
		matchURL, err := url.Parse(endpoint.Path)
		if err != nil {
			return nil, err
		}
		newURL, err := url.Parse(endpoint.InboundPath)
		if err != nil {
			return nil, err
		}
		if portSet[endpoint.InboundPort] {
			virtualHostBuilder.Configure(
				envoy_routes.Route(matchURL.Path, newURL.Path, names.GetLocalClusterName(endpoint.InboundPort), true))
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

	resources := model.NewResourceSet()
	resources.Add(&model.Resource{
		Name:     listenerName,
		Resource: probeListener,
		Origin:   OriginProbe,
	})

	return resources, nil
}
